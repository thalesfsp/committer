package provider

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/thalesfsp/committer/internal/errorcatalog"
	"github.com/thalesfsp/committer/internal/shared"
	"github.com/thalesfsp/committer/internal/textsplitter"
	"github.com/thalesfsp/committer/internal/tui"
	"github.com/thalesfsp/customerror"
	"github.com/thalesfsp/inference/anthropic"
	"github.com/thalesfsp/inference/ollama"
	"github.com/thalesfsp/inference/openai"
	"github.com/thalesfsp/inference/provider"
)

//go:embed commit.prompt
var commitPrompt string

// InitializeLLMProvider initialize the LLM provider.
func InitializeLLMProvider(
	llmProvider string,
	llmModel string,
) (provider.IProvider, error) {
	var providerInUse provider.IProvider

	switch llmProvider {
	case openai.Name:
		oai, err := openai.NewDefault(provider.WithDefaulModel(llmModel))
		if err != nil {
			return nil, errorcatalog.MustGet(errorcatalog.ErrFailedToSetupLLM).New()
		}

		providerInUse = oai
	case anthropic.Name:
		anth, err := anthropic.NewDefault(provider.WithDefaulModel(llmModel))
		if err != nil {
			return nil, errorcatalog.MustGet(errorcatalog.ErrFailedToSetupLLM).New()
		}

		providerInUse = anth
	case ollama.Name:
		oll, err := ollama.NewDefault(provider.WithDefaulModel(llmModel))
		if err != nil {
			return nil, errorcatalog.MustGet(errorcatalog.ErrFailedToSetupLLM).New()
		}

		providerInUse = oll
	default:
		return nil, errorcatalog.MustGet(errorcatalog.ErrInvalidProvider).New()
	}

	return providerInUse, nil
}

// ChunkDiff chunks the diff if it's too big.
func ChunkDiff(maxChars int, diff string) ([]string, error) {
	// Should do nothing if the diff is smaller than the threshold.
	if len(diff) <= maxChars {
		return []string{diff}, nil
	}

	splitter := textsplitter.NewTokenSplitter(maxChars)

	chunks, err := splitter.SplitText(diff)
	if err != nil {
		return nil, errorcatalog.MustGet(errorcatalog.ErrFailedToChunkDiff, customerror.WithError(err))
	}

	return chunks, nil
}

// CallLLM calls the LLM API (OpenAI, Anthropic, or Ollama, etc).
func CallLLM(
	ctx context.Context,
	providerInUse provider.IProvider,
	llmAPICallTimeout time.Duration,
	prompt string,
) (string, error) {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), llmAPICallTimeout)
	defer cancel()

	response, err := providerInUse.Completion(
		ctxWithTimeout,
		provider.WithUserMessages(prompt),
	)
	if err != nil {
		return "", err
	}

	return response, nil
}

// handleTryAgain handles the "Try again" choice.
//
//nolint:lll
func HandleTryAgain() string {
	changeChoice := tui.MustPromptWithChoices("What would you like to change?", []string{
		"Make more succinct",
		"Make more technical",
		"Make less technical",
		"Write what should change",
	})

	switch changeChoice {
	case "Make more succinct":
		return "Please make the commit message more succinct while still conveying the essence of the change."
	case "Make more technical":
		return `Please make the commit message more technical, adding IF POSSIBLE, more context and details aiding engineering comprehension:

1. Include relevant technical terms, e.g., function names, data structures, or algorithms modified.
2. Specify the exact files or modules affected.
3. For bug fixes, IF possible, briefly describe the root cause and solution.
4. For new features, IF possible, outline the core implementation approach.
5. Use concise language while maintaining technical accuracy.

Examples:
- "Optimized database query in user_auth.py using indexing"
- "Implemented red-black tree for efficient sorting in data_processor.cpp"
- "Fixed race condition in thread pool by adding mutex lock in worker.java"

Aim for a balance between technical depth and clarity. Prioritize information that aids code review and future maintenance. No more than 1000 characters!`
	case "Make less technical":
		return `Please make commit messages non-technical, suitable for general audiences. Aim for brevity while still conveying the essence of the change. Examples:

- For updating dependencies: "Updated dependencies"
- For fixing a bug: "Fixed login issue"
- For adding a feature: "Added dark mode"
- For refactoring: "Improved code structure"

For complex changes, summarize the overall impact rather than listing technical details. If multiple significant changes are present, use a bulleted list.`
	case "Write what should change":
		return tui.MustPromptForInputTea("Describe what should change:")
	default:
		return ""
	}
}

// GenerateCommitMessageLoop definition.
func GenerateCommitMessageLoop(
	providerInUse provider.IProvider,
	llmAPICallTimeout time.Duration,
	stats string, chunks []string,
) (string, error) {
	totalChunks := len(chunks)

	additionalInstructions := ""

	maxAttempts := 5 // Define a maximum number of attempts to prevent infinite loops

	for attempt := 0; attempt < maxAttempts; attempt++ {
		for i, chunk := range chunks {
			tui.SprinnerStart("Generating commit message...")

			message, err := GenerateCommitMessage(
				context.Background(),
				providerInUse,
				llmAPICallTimeout,
				stats, chunk,
				i+1, totalChunks,
				additionalInstructions,
			)
			if err != nil {
				return "", fmt.Errorf("failed to generate commit message: %w", err)
			}

			tui.SprinnerStop()

			fmt.Printf("%s\n\n%s\n\n", tui.QuestionStyle.Render("Generated Commit Message:"), message)

			choice := tui.MustPromptWithChoices("What would you like to do?", []string{
				"Approve commit message",
				"Try again",
				"Write commit message yourself",
				"Exit",
			})

			switch choice {
			case "Approve commit message":
				return message, nil
			case "Try again":
				additionalInstructions = HandleTryAgain()

				break // Break inner loop to regenerate
			case "Write commit message yourself":
				content, err := tui.CommitMessageTextArea()
				if err != nil {
					return "", fmt.Errorf("failed to get commit message: %w", err)
				}

				return content + "\n", nil
			case "Exit":
				shared.NothingToDo()
			}

			break // Break inner loop if "Try again" was selected
		}
	}

	return "", fmt.Errorf("maximum attempts reached")
}

// GenerateCommitMessage generates a commit message using LLM API with
// additional instructions.
func GenerateCommitMessage(
	ctx context.Context,
	providerInUse provider.IProvider,
	llmAPICallTimeout time.Duration,
	stats, diff string,
	chunkNumber, totalChunks int,
	additionalInstructions string,
) (string, error) {
	var prompt string
	if totalChunks > 1 {
		// Diff is chunked.
		prompt = fmt.Sprintf(commitPrompt,
			"Git diff is too big, so we chunked it into smaller parts!",
			stats,
			fmt.Sprintf("Chunk %d of %d:", chunkNumber, totalChunks),
			diff,
			fmt.Sprintf("**%s**", additionalInstructions),
		)
	} else {
		// Diff is not chunked.
		prompt = fmt.Sprintf(commitPrompt,
			"",
			stats,
			"",
			diff,
			fmt.Sprintf("**%s**", additionalInstructions),
		)
	}

	// Call LLM API
	message, err := CallLLM(ctx, providerInUse, llmAPICallTimeout, prompt)
	if err != nil {
		return "", err
	}

	return message, nil
}
