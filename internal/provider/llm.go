package provider

import (
	"cmp"
	"context"
	"fmt"
	"strings"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/anthropic"
	"charm.land/fantasy/providers/azure"
	"charm.land/fantasy/providers/bedrock"
	"charm.land/fantasy/providers/google"
	"charm.land/fantasy/providers/openai"
	"charm.land/fantasy/providers/openaicompat"
	"charm.land/fantasy/providers/openrouter"
	"github.com/thalesfsp/committer/internal/errorcatalog"
)

//////
// Const, vars, types.
//////

// DefaultMaxOutputTokens bounds the size of a generated commit message. It is
// generous on purpose: reasoning models spend part of this budget thinking
// before they answer.
const DefaultMaxOutputTokens int64 = 16384

// placeholderAPIKey is sent to endpoints that don't authenticate (Ollama, LM
// Studio, ...) so the SDK still builds a well-formed request.
const placeholderAPIKey = "no-key"

// Config selects and configures the LLM provider.
type Config struct {
	// Provider is a canonical provider name or alias, see Lookup.
	Provider string

	// Model overrides the provider's default model.
	Model string

	// BaseURL overrides the provider endpoint.
	BaseURL string

	// MaxOutputTokens caps the response size. Zero means
	// DefaultMaxOutputTokens.
	MaxOutputTokens int64
}

// LLM is the minimal language model surface committer needs. It hides the
// underlying SDK so the rest of the application, and the tests, don't depend
// on it.
type LLM interface {
	// Generate returns the model's text response to a single user prompt.
	Generate(ctx context.Context, prompt string) (string, error)

	// Provider returns the provider name.
	Provider() string

	// Model returns the model identifier.
	Model() string
}

// fantasyLLM adapts a Fantasy agent to the LLM interface.
type fantasyLLM struct {
	agent    fantasy.Agent
	provider string
	model    string
}

//////
// Implements the LLM interface.
//////

// Generate returns the model's text response to a single user prompt. The
// underlying agent retries transient failures (rate limits, 5xx, network
// errors) with exponential backoff.
func (f *fantasyLLM) Generate(ctx context.Context, prompt string) (string, error) {
	result, err := f.agent.Generate(ctx, fantasy.AgentCall{Prompt: prompt})
	if err != nil {
		return "", fmt.Errorf("%w: %w",
			errorcatalog.MustGet(errorcatalog.ErrFailedToCallLLM).NewFailedToError(), err)
	}

	text := strings.TrimSpace(result.Response.Content.Text())
	if text == "" {
		return "", fmt.Errorf("%w: %s/%s returned no text (finish reason: %s)",
			errorcatalog.MustGet(errorcatalog.ErrFailedToCallLLM).NewFailedToError(),
			f.provider, f.model, result.Response.FinishReason)
	}

	return text, nil
}

// Provider returns the provider name.
func (f *fantasyLLM) Provider() string {
	return f.provider
}

// Model returns the model identifier.
func (f *fantasyLLM) Model() string {
	return f.model
}

//////
// Exported functionalities.
//////

// InitializeLLMProvider resolves the configuration against the registry and
// the environment, and returns a ready to use LLM.
func InitializeLLMProvider(ctx context.Context, cfg Config) (LLM, error) {
	spec, ok := Lookup(cfg.Provider)
	if !ok {
		return nil, fmt.Errorf("%w: %q, allowed: %s",
			errorcatalog.MustGet(errorcatalog.ErrInvalidProvider).NewInvalidError(),
			cfg.Provider, strings.Join(Names(), ", "))
	}

	model := cmp.Or(strings.TrimSpace(cfg.Model), spec.DefaultModel)
	if model == "" {
		return nil, fmt.Errorf("%w: %s has no default model, set --model (-m)",
			errorcatalog.MustGet(errorcatalog.ErrMissingModel).NewMissingError(), spec.Name)
	}

	apiKey := spec.APIKey()
	if apiKey == "" && spec.APIKeyRequired {
		return nil, fmt.Errorf("%w: set %s",
			errorcatalog.MustGet(errorcatalog.ErrMissingAPIKey).NewMissingError(),
			strings.Join(spec.APIKeyEnvs, " or "))
	}

	baseURL := spec.BaseURL(cfg.BaseURL)
	if baseURL == "" && spec.BaseURLRequired {
		return nil, fmt.Errorf("%w: set --base-url or %s",
			errorcatalog.MustGet(errorcatalog.ErrMissingBaseURL).NewMissingError(),
			strings.Join(spec.BaseURLEnvs, " or "))
	}

	fantasyProvider, err := newFantasyProvider(spec, apiKey, baseURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w",
			errorcatalog.MustGet(errorcatalog.ErrFailedToSetupLLM).NewFailedToError(), err)
	}

	languageModel, err := fantasyProvider.LanguageModel(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("%w: %w",
			errorcatalog.MustGet(errorcatalog.ErrFailedToSetupLLM).NewFailedToError(), err)
	}

	maxOutputTokens := cfg.MaxOutputTokens
	if maxOutputTokens <= 0 {
		maxOutputTokens = DefaultMaxOutputTokens
	}

	return &fantasyLLM{
		agent:    fantasy.NewAgent(languageModel, fantasy.WithMaxOutputTokens(maxOutputTokens)),
		provider: spec.Name,
		model:    model,
	}, nil
}

//////
// Internal functionalities.
//////

// newFantasyProvider builds the Fantasy provider matching the spec's kind.
func newFantasyProvider(spec Spec, apiKey, baseURL string) (fantasy.Provider, error) {
	switch spec.Kind {
	case KindOpenAI:
		return openai.New(openai.WithAPIKey(apiKey), openai.WithBaseURL(baseURL))
	case KindAnthropic:
		return anthropic.New(anthropic.WithAPIKey(apiKey), anthropic.WithBaseURL(baseURL))
	case KindGoogle:
		return google.New(google.WithGeminiAPIKey(apiKey), google.WithBaseURL(baseURL))
	case KindOpenAICompat:
		return openaicompat.New(
			openaicompat.WithName(spec.Name),
			openaicompat.WithBaseURL(baseURL),
			openaicompat.WithAPIKey(cmp.Or(apiKey, placeholderAPIKey)),
		)
	case KindOpenRouter:
		if baseURL != "" && baseURL != openrouter.DefaultURL {
			return nil, fmt.Errorf("%s does not support a custom base URL", spec.Name)
		}

		return openrouter.New(openrouter.WithAPIKey(apiKey))
	case KindAzure:
		opts := []azure.Option{azure.WithAPIKey(apiKey), azure.WithBaseURL(baseURL)}

		if version := firstEnv("AZURE_OPENAI_API_VERSION"); version != "" {
			opts = append(opts, azure.WithAPIVersion(version))
		}

		return azure.New(opts...)
	case KindBedrock:
		opts := []bedrock.Option{}

		if apiKey != "" {
			opts = append(opts, bedrock.WithAPIKey(apiKey))
		}

		if region := firstEnv("AWS_REGION", "AWS_DEFAULT_REGION"); region != "" {
			opts = append(opts, bedrock.WithRegion(region))
		}

		if baseURL != "" {
			opts = append(opts, bedrock.WithBaseURL(baseURL))
		}

		return bedrock.New(opts...)
	default:
		return nil, fmt.Errorf("unsupported provider kind %q", spec.Kind)
	}
}
