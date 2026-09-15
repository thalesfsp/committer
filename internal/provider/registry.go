package provider

import (
	"maps"
	"os"
	"slices"
	"strings"
)

//////
// Const, vars, types.
//////

// Kind identifies which Fantasy provider package backs a Spec.
type Kind string

// Supported kinds.
const (
	KindAnthropic    Kind = "anthropic"
	KindAzure        Kind = "azure"
	KindBedrock      Kind = "bedrock"
	KindGoogle       Kind = "google"
	KindOpenAI       Kind = "openai"
	KindOpenAICompat Kind = "openai-compat"
	KindOpenRouter   Kind = "openrouter"
)

// Names of the supported providers, as accepted by the --provider flag.
const (
	Anthropic        = "anthropic"
	Azure            = "azure"
	Bedrock          = "bedrock"
	DeepSeek         = "deepseek"
	Google           = "google"
	Groq             = "groq"
	HuggingFace      = "huggingface"
	Mistral          = "mistral"
	Ollama           = "ollama"
	OpenAI           = "openai"
	OpenAICompatible = "openai-compatible"
	OpenRouter       = "openrouter"
	XAI              = "xai"
)

// Spec describes how committer reaches one LLM provider.
type Spec struct {
	// Name is the identifier accepted by the --provider flag.
	Name string

	// Description is a short, human readable summary shown in help.
	Description string

	// Kind selects the Fantasy provider package used to talk to the API.
	Kind Kind

	// APIKeyEnvs are the environment variables checked, in order, for the
	// API key. The first non-empty value wins.
	APIKeyEnvs []string

	// APIKeyRequired makes initialization fail early when no API key is
	// found. Providers whose SDK resolves credentials on its own (for
	// example Anthropic auth tokens, or AWS credentials for Bedrock) or that
	// need no key at all (Ollama) leave it false.
	APIKeyRequired bool

	// BaseURLEnvs are the environment variables checked, in order, for an
	// endpoint override. The --base-url flag takes precedence over them.
	BaseURLEnvs []string

	// DefaultBaseURL is used when neither the flag nor the environment
	// provide one. Empty means the SDK default.
	DefaultBaseURL string

	// BaseURLRequired makes initialization fail early when no base URL is
	// available.
	BaseURLRequired bool

	// DefaultModel is used when --model is not set. Empty means the model
	// must be provided explicitly.
	DefaultModel string
}

// specs lists the supported providers in display order.
var specs = []Spec{
	{
		Name:           OpenAI,
		Description:    "OpenAI (GPT models)",
		Kind:           KindOpenAI,
		APIKeyEnvs:     []string{"OPENAI_API_KEY"},
		APIKeyRequired: true,
		BaseURLEnvs:    []string{"OPENAI_BASE_URL"},
		DefaultBaseURL: "https://api.openai.com/v1",
		DefaultModel:   "gpt-5.4-mini",
	},
	{
		Name:           Anthropic,
		Description:    "Anthropic (Claude models)",
		Kind:           KindAnthropic,
		APIKeyEnvs:     []string{"ANTHROPIC_API_KEY"},
		BaseURLEnvs:    []string{"ANTHROPIC_BASE_URL"},
		DefaultBaseURL: "https://api.anthropic.com",
		DefaultModel:   "claude-opus-5",
	},
	{
		Name:           Google,
		Description:    "Google (Gemini models)",
		Kind:           KindGoogle,
		APIKeyEnvs:     []string{"GEMINI_API_KEY", "GOOGLE_API_KEY"},
		APIKeyRequired: true,
		BaseURLEnvs:    []string{"GOOGLE_GEMINI_BASE_URL", "GEMINI_BASE_URL"},
		DefaultBaseURL: "https://generativelanguage.googleapis.com",
		DefaultModel:   "gemini-3.5-flash",
	},
	{
		Name:           XAI,
		Description:    "xAI (Grok models)",
		Kind:           KindOpenAICompat,
		APIKeyEnvs:     []string{"XAI_API_KEY"},
		APIKeyRequired: true,
		BaseURLEnvs:    []string{"XAI_BASE_URL"},
		DefaultBaseURL: "https://api.x.ai/v1",
		DefaultModel:   "grok-4.5",
	},
	{
		Name:           Groq,
		Description:    "Groq",
		Kind:           KindOpenAICompat,
		APIKeyEnvs:     []string{"GROQ_API_KEY"},
		APIKeyRequired: true,
		BaseURLEnvs:    []string{"GROQ_BASE_URL"},
		DefaultBaseURL: "https://api.groq.com/openai/v1",
		DefaultModel:   "moonshotai/kimi-k2-instruct-0905",
	},
	{
		Name:           DeepSeek,
		Description:    "DeepSeek",
		Kind:           KindOpenAICompat,
		APIKeyEnvs:     []string{"DEEPSEEK_API_KEY"},
		APIKeyRequired: true,
		BaseURLEnvs:    []string{"DEEPSEEK_BASE_URL"},
		DefaultBaseURL: "https://api.deepseek.com/v1",
		DefaultModel:   "deepseek-v4-flash",
	},
	{
		Name:           Mistral,
		Description:    "Mistral AI",
		Kind:           KindOpenAICompat,
		APIKeyEnvs:     []string{"MISTRAL_API_KEY"},
		APIKeyRequired: true,
		BaseURLEnvs:    []string{"MISTRAL_BASE_URL"},
		DefaultBaseURL: "https://api.mistral.ai/v1",
		DefaultModel:   "mistral-small-latest",
	},
	{
		Name:           OpenRouter,
		Description:    "OpenRouter (many models behind one key)",
		Kind:           KindOpenRouter,
		APIKeyEnvs:     []string{"OPENROUTER_API_KEY"},
		APIKeyRequired: true,
		DefaultBaseURL: "https://openrouter.ai/api/v1",
		DefaultModel:   "openrouter/auto",
	},
	{
		Name:           HuggingFace,
		Description:    "Hugging Face Inference Providers",
		Kind:           KindOpenAICompat,
		APIKeyEnvs:     []string{"HF_TOKEN", "HUGGINGFACE_API_KEY"},
		APIKeyRequired: true,
		BaseURLEnvs:    []string{"HUGGINGFACE_BASE_URL"},
		DefaultBaseURL: "https://router.huggingface.co/v1",
		DefaultModel:   "openai/gpt-oss-20b",
	},
	{
		Name:           Ollama,
		Description:    "Ollama (local, offline)",
		Kind:           KindOpenAICompat,
		APIKeyEnvs:     []string{"OLLAMA_API_KEY"},
		BaseURLEnvs:    []string{"OLLAMA_HOST", "OLLAMA_ENDPOINT"},
		DefaultBaseURL: "http://localhost:11434/v1",
		DefaultModel:   "llama3.2",
	},
	{
		Name:            Azure,
		Description:     "Azure OpenAI (the model is the deployment name)",
		Kind:            KindAzure,
		APIKeyEnvs:      []string{"AZURE_OPENAI_API_KEY"},
		APIKeyRequired:  true,
		BaseURLEnvs:     []string{"AZURE_OPENAI_ENDPOINT", "AZURE_OPENAI_BASE_URL"},
		BaseURLRequired: true,
	},
	{
		Name:        Bedrock,
		Description: "Amazon Bedrock (Anthropic models, AWS credentials)",
		Kind:        KindBedrock,
		APIKeyEnvs:  []string{"AWS_BEARER_TOKEN_BEDROCK"},
		BaseURLEnvs: []string{"BEDROCK_BASE_URL"},
	},
	{
		Name:            OpenAICompatible,
		Description:     "Any OpenAI-compatible endpoint (LM Studio, vLLM, LiteLLM, Cerebras, ...)",
		Kind:            KindOpenAICompat,
		APIKeyEnvs:      []string{"OPENAI_COMPATIBLE_API_KEY"},
		BaseURLEnvs:     []string{"OPENAI_COMPATIBLE_BASE_URL"},
		BaseURLRequired: true,
	},
}

// aliases maps friendly names to canonical provider names.
var aliases = map[string]string{
	"claude":        Anthropic,
	"custom":        OpenAICompatible,
	"gemini":        Google,
	"grok":          XAI,
	"hf":            HuggingFace,
	"openai-compat": OpenAICompatible,
}

//////
// Exported functionalities.
//////

// Specs returns the supported providers in display order.
func Specs() []Spec {
	return slices.Clone(specs)
}

// Aliases returns a copy of the alias to canonical provider name map.
func Aliases() map[string]string {
	return maps.Clone(aliases)
}

// Names returns the canonical provider names in display order.
func Names() []string {
	names := make([]string, 0, len(specs))

	for _, spec := range specs {
		names = append(names, spec.Name)
	}

	return names
}

// Lookup finds a provider by canonical name or alias. Matching is
// case-insensitive and ignores surrounding whitespace.
func Lookup(name string) (Spec, bool) {
	normalized := strings.ToLower(strings.TrimSpace(name))

	if canonical, ok := aliases[normalized]; ok {
		normalized = canonical
	}

	for _, spec := range specs {
		if spec.Name == normalized {
			return spec, true
		}
	}

	return Spec{}, false
}

// APIKey returns the first API key found in the environment, or an empty
// string.
func (s Spec) APIKey() string {
	return firstEnv(s.APIKeyEnvs...)
}

// BaseURL resolves the endpoint to use: the override (from the --base-url
// flag) wins, then the environment, then the provider default.
func (s Spec) BaseURL(override string) string {
	url := strings.TrimSpace(override)

	if url == "" {
		url = firstEnv(s.BaseURLEnvs...)
	}

	if url == "" {
		url = s.DefaultBaseURL
	}

	if s.Name == Ollama {
		return normalizeOllamaURL(url)
	}

	return url
}

//////
// Internal functionalities.
//////

// firstEnv returns the value of the first non-empty environment variable.
func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}

	return ""
}

// normalizeOllamaURL accepts the many ways people write the Ollama address
// (OLLAMA_HOST style "localhost:11434", a bare origin, or the legacy
// "/api/chat" endpoint) and returns the OpenAI-compatible "/v1" base URL.
func normalizeOllamaURL(raw string) string {
	url := strings.TrimSpace(raw)
	if url == "" {
		return url
	}

	if !strings.Contains(url, "://") {
		url = "http://" + url
	}

	url = strings.TrimRight(url, "/")

	for _, suffix := range []string{"/api/chat", "/api", "/v1"} {
		url = strings.TrimSuffix(url, suffix)
	}

	return url + "/v1"
}
