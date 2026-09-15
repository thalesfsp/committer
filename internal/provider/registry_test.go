package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	testAPIKey = "test-key"
	testModel  = "test-model"
	testReply  = "feat(x): add y"
	ollamaV1   = "http://localhost:11434/v1"
)

// capturedRequest records what the fake API server received.
type capturedRequest struct {
	path   string
	auth   string
	apiKey string
	model  string
	body   string
}

// newFakeServer serves a JSON API that records each request in captured and
// answers with the body produced by respond for the requested model.
func newFakeServer(t *testing.T, captured *capturedRequest, respond func(model string) string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}

		var body struct {
			Model string `json:"model"`
		}

		if err := json.Unmarshal(raw, &body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		captured.path = r.URL.Path
		captured.auth = r.Header.Get("Authorization")
		captured.apiKey = r.Header.Get("x-api-key")
		captured.model = body.Model
		captured.body = string(raw)

		w.Header().Set("Content-Type", "application/json")

		_, _ = io.WriteString(w, respond(body.Model))
	}))
}

// chatCompletionsReply is a minimal OpenAI-compatible chat completion.
func chatCompletionsReply(model string) string {
	return `{"id":"chatcmpl-1","object":"chat.completion","created":1,"model":"` + model + `",` +
		`"choices":[{"index":0,"message":{"role":"assistant","content":"` + testReply + `"},"finish_reason":"stop"}],` +
		`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`
}

// anthropicMessagesReply is a minimal Anthropic Messages API response.
func anthropicMessagesReply(model string) string {
	return `{"id":"msg_1","type":"message","role":"assistant","model":"` + model + `",` +
		`"content":[{"type":"text","text":"` + testReply + `"}],"stop_reason":"end_turn","stop_sequence":null,` +
		`"usage":{"input_tokens":1,"output_tokens":1}}`
}

func TestLookup(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{in: "openai", want: OpenAI, ok: true},
		{in: " Anthropic ", want: Anthropic, ok: true},
		{in: "claude", want: Anthropic, ok: true},
		{in: "GROK", want: XAI, ok: true},
		{in: "gemini", want: Google, ok: true},
		{in: "hf", want: HuggingFace, ok: true},
		{in: "custom", want: OpenAICompatible, ok: true},
		{in: "openai-compat", want: OpenAICompatible, ok: true},
		{in: "nope", want: "", ok: false},
		{in: "", want: "", ok: false},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			spec, ok := Lookup(tc.in)
			if ok != tc.ok {
				t.Fatalf("Lookup(%q) ok = %v, want %v", tc.in, ok, tc.ok)
			}

			if spec.Name != tc.want {
				t.Errorf("Lookup(%q) name = %q, want %q", tc.in, spec.Name, tc.want)
			}
		})
	}
}

func TestSpecs_AreConsistent(t *testing.T) {
	seen := map[string]bool{}

	for _, spec := range Specs() {
		if seen[spec.Name] {
			t.Errorf("duplicate provider name %q", spec.Name)
		}

		seen[spec.Name] = true

		if spec.Description == "" {
			t.Errorf("%s: missing description", spec.Name)
		}

		if spec.Kind == "" {
			t.Errorf("%s: missing kind", spec.Name)
		}

		if spec.APIKeyRequired && len(spec.APIKeyEnvs) == 0 {
			t.Errorf("%s: API key required but no env var listed", spec.Name)
		}

		if spec.BaseURLRequired && len(spec.BaseURLEnvs) == 0 {
			t.Errorf("%s: base URL required but no env var listed", spec.Name)
		}

		if _, ok := Lookup(spec.Name); !ok {
			t.Errorf("%s: not found by Lookup", spec.Name)
		}
	}

	if len(Names()) != len(Specs()) {
		t.Errorf("Names() has %d entries, Specs() has %d", len(Names()), len(Specs()))
	}

	for alias, name := range Aliases() {
		if _, ok := Lookup(name); !ok {
			t.Errorf("alias %q points to unknown provider %q", alias, name)
		}
	}
}

func TestSpec_APIKey_Precedence(t *testing.T) {
	spec, _ := Lookup(HuggingFace)

	t.Setenv("HF_TOKEN", "")
	t.Setenv("HUGGINGFACE_API_KEY", "")

	if got := spec.APIKey(); got != "" {
		t.Errorf("expected empty API key, got %q", got)
	}

	t.Setenv("HUGGINGFACE_API_KEY", "legacy")

	if got := spec.APIKey(); got != "legacy" {
		t.Errorf("expected legacy env var to be used, got %q", got)
	}

	t.Setenv("HF_TOKEN", "primary")

	if got := spec.APIKey(); got != "primary" {
		t.Errorf("expected first env var to win, got %q", got)
	}
}

func TestSpec_BaseURL_Precedence(t *testing.T) {
	spec, _ := Lookup(XAI)

	t.Setenv("XAI_BASE_URL", "")

	if got := spec.BaseURL(""); got != "https://api.x.ai/v1" {
		t.Errorf("expected default base URL, got %q", got)
	}

	t.Setenv("XAI_BASE_URL", "http://from-env")

	if got := spec.BaseURL(""); got != "http://from-env" {
		t.Errorf("expected env base URL, got %q", got)
	}

	if got := spec.BaseURL("http://from-flag"); got != "http://from-flag" {
		t.Errorf("expected flag base URL to win, got %q", got)
	}
}

func TestNormalizeOllamaURL(t *testing.T) {
	cases := map[string]string{
		"":                                "",
		"localhost:11434":                 ollamaV1,
		"http://localhost:11434":          ollamaV1,
		"http://localhost:11434/":         ollamaV1,
		"http://localhost:11434/v1":       ollamaV1,
		"http://localhost:11434/api/chat": ollamaV1,
		"https://ollama.example.com/api":  "https://ollama.example.com/v1",
	}

	for in, want := range cases {
		if got := normalizeOllamaURL(in); got != want {
			t.Errorf("normalizeOllamaURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestInitializeLLMProvider_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("unknown provider", func(t *testing.T) {
		_, err := InitializeLLMProvider(ctx, Config{Provider: "nope"})
		if err == nil || !strings.Contains(err.Error(), "invalid provider") {
			t.Fatalf("expected invalid provider error, got %v", err)
		}
	})

	t.Run("missing API key", func(t *testing.T) {
		t.Setenv("XAI_API_KEY", "")

		_, err := InitializeLLMProvider(ctx, Config{Provider: XAI})
		if err == nil || !strings.Contains(err.Error(), "XAI_API_KEY") {
			t.Fatalf("expected missing API key error naming XAI_API_KEY, got %v", err)
		}
	})

	t.Run("missing model", func(t *testing.T) {
		t.Setenv("AZURE_OPENAI_API_KEY", testAPIKey)
		t.Setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")

		_, err := InitializeLLMProvider(ctx, Config{Provider: Azure})
		if err == nil || !strings.Contains(err.Error(), "missing model") {
			t.Fatalf("expected missing model error, got %v", err)
		}
	})

	t.Run("missing base URL", func(t *testing.T) {
		t.Setenv("AZURE_OPENAI_API_KEY", testAPIKey)
		t.Setenv("AZURE_OPENAI_ENDPOINT", "")
		t.Setenv("AZURE_OPENAI_BASE_URL", "")

		_, err := InitializeLLMProvider(ctx, Config{Provider: Azure, Model: "deployment"})
		if err == nil || !strings.Contains(err.Error(), "missing base URL") {
			t.Fatalf("expected missing base URL error, got %v", err)
		}
	})

	t.Run("openrouter rejects a custom base URL", func(t *testing.T) {
		t.Setenv("OPENROUTER_API_KEY", testAPIKey)

		_, err := InitializeLLMProvider(ctx, Config{Provider: OpenRouter, BaseURL: "http://localhost:1"})
		if err == nil || !strings.Contains(err.Error(), "custom base URL") {
			t.Fatalf("expected custom base URL error, got %v", err)
		}
	})
}

func TestInitializeLLMProvider_OpenAICompatibleRoundTrip(t *testing.T) {
	captured := &capturedRequest{}
	server := newFakeServer(t, captured, chatCompletionsReply)

	defer server.Close()

	t.Setenv("OPENAI_COMPATIBLE_API_KEY", testAPIKey)

	llm, err := InitializeLLMProvider(context.Background(), Config{
		Provider: OpenAICompatible,
		Model:    testModel,
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if llm.Provider() != OpenAICompatible || llm.Model() != testModel {
		t.Errorf("unexpected identity %s/%s", llm.Provider(), llm.Model())
	}

	got, err := CallLLM(context.Background(), llm, 10*time.Second, "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != testReply {
		t.Errorf("expected %q, got %q", testReply, got)
	}

	if captured.path != "/chat/completions" {
		t.Errorf("expected /chat/completions, got %q", captured.path)
	}

	if captured.auth != "Bearer "+testAPIKey {
		t.Errorf("expected bearer auth with the configured key, got %q", captured.auth)
	}

	if captured.model != testModel {
		t.Errorf("expected model %q, got %q", testModel, captured.model)
	}

	if !strings.Contains(captured.body, "hello world") {
		t.Errorf("expected the prompt in the request body, got %s", captured.body)
	}
}

func TestInitializeLLMProvider_OllamaWithoutKey(t *testing.T) {
	captured := &capturedRequest{}
	server := newFakeServer(t, captured, chatCompletionsReply)

	defer server.Close()

	t.Setenv("OLLAMA_API_KEY", "")
	t.Setenv("OLLAMA_HOST", strings.TrimPrefix(server.URL, "http://"))
	t.Setenv("OLLAMA_ENDPOINT", "")

	llm, err := InitializeLLMProvider(context.Background(), Config{Provider: Ollama})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if llm.Model() != "llama3.2" {
		t.Errorf("expected the default Ollama model, got %q", llm.Model())
	}

	got, err := CallLLM(context.Background(), llm, 10*time.Second, "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != testReply {
		t.Errorf("expected %q, got %q", testReply, got)
	}

	if captured.path != "/v1/chat/completions" {
		t.Errorf("expected the OpenAI-compatible Ollama path, got %q", captured.path)
	}
}

func TestInitializeLLMProvider_AnthropicRoundTrip(t *testing.T) {
	captured := &capturedRequest{}
	server := newFakeServer(t, captured, anthropicMessagesReply)

	defer server.Close()

	t.Setenv("ANTHROPIC_API_KEY", testAPIKey)

	llm, err := InitializeLLMProvider(context.Background(), Config{
		Provider: "claude",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if llm.Provider() != Anthropic || llm.Model() != "claude-opus-5" {
		t.Errorf("unexpected identity %s/%s", llm.Provider(), llm.Model())
	}

	got, err := CallLLM(context.Background(), llm, 10*time.Second, "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != testReply {
		t.Errorf("expected %q, got %q", testReply, got)
	}

	if captured.path != "/v1/messages" {
		t.Errorf("expected /v1/messages, got %q", captured.path)
	}

	if captured.apiKey != testAPIKey {
		t.Errorf("expected x-api-key header with the configured key, got %q", captured.apiKey)
	}

	if captured.model != "claude-opus-5" {
		t.Errorf("expected model claude-opus-5, got %q", captured.model)
	}
}
