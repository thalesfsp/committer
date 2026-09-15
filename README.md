# committer

![molonelaveh_A_futuristic_command-line_interface_terminal_floa_d86d758a-2a56-4187-8e2a-052bbf1135cd_0](https://github.com/user-attachments/assets/c27172d2-11f7-4e55-ae38-a3cb85673ae8)

Committer is a beautiful command-line tool (CLI) designed to leverage large language models (LLMs) to streamline the process of generating meaningful, concise, and descriptive commit messages.

## Features

- **Generate Commit Messages**: Automatically generates commit messages using LLMs based on staged changes.
- **Provider Flexibility**: Works with OpenAI, Anthropic (Claude), Google (Gemini), xAI (Grok), Groq, DeepSeek, Mistral, OpenRouter, Hugging Face, Ollama (offline), Azure OpenAI, Amazon Bedrock, and any OpenAI-compatible endpoint.
- **Interactive CLI**: Provides an interactive TUI to guide users through the process.
- **Retry Mechanism**: Offers options to regenerate commit messages, change the prompt on-the-fly by making it more or less technical or any additional custom instruction, or manually edit that.
- **Chunking Large Diffs**: Smart chunking properly splits large diffs into chunks for efficient processing, and message generation.
- **Git Flow**: Capable of seamlessly stage files, commit, push, and tag changes.
- **Native Git Integration**: Built-in safe sanity checks, importantly, it respect `.gitignore`!

## Architecture Overview

The application is built using Go and utilizes the Cobra library for CLI interactions. It integrates with Git for version control operations and uses various LLM providers for generating commit messages. It also incorporates the Bubble Tea framework for TUI providing interactive CLI components and uses Lipgloss for styling.

LLM access goes through [Fantasy](https://github.com/charmbracelet/fantasy), Charm's multi-provider AI SDK for Go (the same library behind [Crush](https://github.com/charmbracelet/crush)). Committer keeps a small registry (`internal/provider/registry.go`) describing each provider: its API key and endpoint environment variables, its default endpoint, and its default model. Adding a provider is one entry in that table.

## Install

### CLI

`curl -s https://raw.githubusercontent.com/thalesfsp/committer/main/resources/install.sh | sh`

Setting target destination:

`curl -s https://raw.githubusercontent.com/thalesfsp/committer/main/resources/install.sh | BIN_DIR=ABSOLUTE_DIR_PATH sh`

Setting version:

`curl -s https://raw.githubusercontent.com/thalesfsp/committer/main/resources/install.sh | VERSION=v{M.M.P} sh`

Example:

`curl -s https://raw.githubusercontent.com/thalesfsp/committer/main/resources/install.sh | BIN_DIR=/usr/local/bin VERSION=v1.3.17 sh`

## Usage

1. Set the API key for the LLM provider in the environment variable, example: `export OPENAI_API_KEY=sk-...`

_Note: Update your shell (Fish, Bash, ZSH) config to persist the change_

2. Run `$ committer`
3. Happy work!

Pick another provider with `-p` (and optionally a model with `-m`):

```sh
export ANTHROPIC_API_KEY=sk-ant-...
committer -p anthropic                       # Claude, default model

export XAI_API_KEY=xai-...
committer -p grok -m grok-4.5                # xAI, "grok" is an alias of "xai"

committer -p ollama -m qwen3                 # local Ollama, no API key needed

export OPENAI_COMPATIBLE_API_KEY=...
committer -p openai-compatible --base-url https://api.cerebras.ai/v1 -m gpt-oss-120b
```

Auto-accept mode (`-a`) never prompts, so it also works without a terminal, for example from a git hook or a CI job, printing plain progress lines instead of the spinner.

To stop repeating the flags, set the defaults once in your shell config:

```sh
export COMMITTER_PROVIDER=anthropic
export COMMITTER_MODEL=claude-sonnet-5
```

### Providers

Run `committer providers` for the up-to-date list. Every provider has a default model, so `-m` is only needed to pick a different one.

| Provider (`-p`) | Aliases | API key | Endpoint override | Default model |
| --- | --- | --- | --- | --- |
| `openai` | | `OPENAI_API_KEY` | `OPENAI_BASE_URL` | `gpt-5.4-mini` |
| `anthropic` | `claude` | `ANTHROPIC_API_KEY` | `ANTHROPIC_BASE_URL` | `claude-opus-5` |
| `google` | `gemini` | `GEMINI_API_KEY` or `GOOGLE_API_KEY` | `GOOGLE_GEMINI_BASE_URL` | `gemini-3.5-flash` |
| `xai` | `grok` | `XAI_API_KEY` | `XAI_BASE_URL` | `grok-4.5` |
| `groq` | | `GROQ_API_KEY` | `GROQ_BASE_URL` | `moonshotai/kimi-k2-instruct-0905` |
| `deepseek` | | `DEEPSEEK_API_KEY` | `DEEPSEEK_BASE_URL` | `deepseek-v4-flash` |
| `mistral` | | `MISTRAL_API_KEY` | `MISTRAL_BASE_URL` | `mistral-small-latest` |
| `openrouter` | | `OPENROUTER_API_KEY` | | `openrouter/auto` |
| `huggingface` | `hf` | `HF_TOKEN` or `HUGGINGFACE_API_KEY` | `HUGGINGFACE_BASE_URL` | `openai/gpt-oss-20b` |
| `ollama` | | `OLLAMA_API_KEY` (optional) | `OLLAMA_HOST` (default `http://localhost:11434`) | `llama3.2` |
| `azure` | | `AZURE_OPENAI_API_KEY` | `AZURE_OPENAI_ENDPOINT` (required) | none, `-m` is the deployment name |
| `bedrock` | | AWS credentials, or `AWS_BEARER_TOKEN_BEDROCK` | `BEDROCK_BASE_URL` | none, e.g. `us.anthropic.claude-opus-5` |
| `openai-compatible` | `custom` | `OPENAI_COMPATIBLE_API_KEY` (optional) | `OPENAI_COMPATIBLE_BASE_URL` or `--base-url` (required) | none |

Notes:

- `--base-url` (`-u`) overrides the endpoint of any provider, handy for proxies such as LiteLLM or self-hosted servers such as LM Studio and vLLM.
- Anthropic also honours the credentials its SDK understands (`ANTHROPIC_AUTH_TOKEN`, or a profile from `ant auth login`). Bedrock uses the AWS default credential chain and `AWS_REGION`. Azure reads `AZURE_OPENAI_API_VERSION` when set.
- `--max-tokens` caps the response size (default 16384). Reasoning models spend part of that budget thinking, so keep it generous.
- `--llm-api-call-timeout` (`-t`, default 60s) bounds each call, including the automatic retries on rate limits and transient errors.

### More Information

Checkout our well-crafted help by running `$ committer --help`.

## Contributing

1. Fork
2. Clone
3. Create a branch
4. Make changes following the same standards as the project (Go 1.27+, `golangci-lint` v2)
5. Run `make ci`
6. Create a merge request
