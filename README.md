# committer

Committer is a beautiful command-line tool (CLI) designed to leverage large language models (LLMs) to streamline the process of generating meaningful, concise, and descriptive commit messages.

## Features

- **Generate Commit Messages**: Automatically generates commit messages using LLMs based on staged changes.
- **Provider Flexibility**: Supports multiple LLM providers, including OpenAI, Anthropic, and Ollama (offline).
- **Interactive CLI**: Provides an interactive TUI to guide users through the process.
- **Retry Mechanism**: Offers options to regenerate commit messages, change the prompt on-the-fly by making it more or less technical or any additional custom instruction, or manually edit that.
- **Chunking Large Diffs**: Smart chunking properly splits large diffs into chunks for efficient processing, and message generation.
- **Git Flow**: Capable of seamlessly stage files, commit, push, and tag changes.
- **Native Git Integration**: Built-in safe sanity checks, importantly, it respect `.gitignore`!

## Architecture Overview

The application is built using Go and utilizes the Cobra library for CLI interactions. It integrates with Git for version control operations and uses various LLM providers for generating commit messages. It also incorporates the Bubble Tea framework for TUI providing interactive CLI components and uses Lipgloss for styling. The CLI is designed to be extensible, supporting multiple LLM providers via a standard interface pattern.

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

1. Set the API key for the LLM provider, for example, `OPENAI_API_KEY`.
2. Run `$ committer`

### More Information

Checkout our well-crafted help by running `$ committer --help`.

## Contributing

1. Fork
2. Clone
3. Create a branch
4. Make changes following the same standards as the project
5. Run `make ci`
6. Create a merge request
