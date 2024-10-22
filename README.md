# committer

Committer CLI is a command-line tool designed to streamline the process of generating meaningful commit messages. It leverages large language models (LLMs) to automatically create concise and descriptive commit messages based on the changes staged in a Git repository.

## Features

- **Generate Commit Messages**: Automatically generates meaningful commit messages using LLMs based on staged Git changes.
- **Interactive CLI**: Provides an interactive command-line interface to guide users through the commit process.
- **Git Integration**: Checks for uncommitted changes and stages files before committing.
- **Chunking Large Diffs**: Splits large diffs into chunks for efficient processing and message generation.
- **Customizable Prompts**: Allows users to customize the message generation with additional instructions.
- **Retry Mechanism**: Offers options to regenerate commit messages or manually edit them.
- **Push and Tag Commits**: Capable of pushing commits and tagging versions in the Git repository.
- **Provider Flexibility**: Supports multiple LLM providers, including OpenAI, Anthropic, and Ollama.

## Architecture Overview

The application is built using Go and utilizes the Cobra library for CLI interactions. It integrates with Git for version control operations and uses various LLM providers for generating commit messages. The codebase incorporates the Bubble Tea framework for interactive CLI components and uses Lipgloss for styling. The CLI is designed to be extensible, supporting multiple LLM providers via an interface pattern.

## Install

### CLI

`curl -s https://raw.githubusercontent.com/thalesfsp/committer/main/resources/install.sh | sh`

Setting target destination:

`curl -s https://raw.githubusercontent.com/thalesfsp/committer/main/resources/install.sh | BIN_DIR=ABSOLUTE_DIR_PATH sh`

Setting version:

`curl -s https://raw.githubusercontent.com/thalesfsp/committer/main/resources/install.sh | VERSION=v{M.M.P} sh`

Example:

`curl -s https://raw.githubusercontent.com/thalesfsp/committer/main/resources/install.sh | BIN_DIR=/usr/local/bin VERSION=v1.3.17 sh`

### Programmatically

Install dependency:

`go get -u github.com/thalesfsp/committer`

## Usage

### CLI

`$ committer --help`

### Programmatically

See `*_test.go` files for examples.

### Documentation

Run `$ make doc` or check out [online](https://pkg.go.dev/github.com/thalesfsp/committer).

## Contributing

1. Fork
2. Clone
3. Create a branch
4. Make changes following the same standards as the project
5. Run `make ci`
6. Create a merge request

### Release flow

1. Update [CHANGELOG](CHANGELOG.md) accordingly.
2. Once changes from MR are merged.
3. Just tag. Don't need to create a release, it's automatically created by CI.

## Roadmap

Check out [CHANGELOG](CHANGELOG.md).
