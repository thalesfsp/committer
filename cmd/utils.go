/*
Git Operations and UI Interactions

This file implements Git operations, LLM interactions, and terminal UI components
for a commit message generation tool. It handles various operations including Git
status checks, diff generation, commit operations, and user interactions through
a terminal UI.
*/

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/thalesfsp/committer/internal/shared"
	"github.com/thalesfsp/committer/internal/tea"
	"github.com/thalesfsp/customerror"
	"github.com/thalesfsp/inference/provider"
)

//////
// Environment Variables
//////

// syplLevel stores the logging level from environment.
var syplLevel = os.Getenv("SYPL_LEVEL")

//////
// Helper functions.
//////

// isDebugMode checks if the CLI is running in debug mode based on SYPL_LEVEL.
func isDebugMode() bool {
	return syplLevel == "debug"
}

// isCurrentDirectoryGitRepo verifies if the current directory is a Git repository.
//
// Returns:
//   - bool: True if current directory is a Git repository, false otherwise.
func isCurrentDirectoryGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprint(os.Stderr, stderr.String())

		return false
	}

	return true
}

// isDirty checks for uncommitted changes in the repository.
//
// Returns:
//   - bool: True if there are uncommitted changes, false otherwise.
func isDirty() bool {
	cmd := exec.Command("git", "diff", "--quiet")

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	return cmd.Run() != nil
}

// hasStagedChanges verifies if there are staged changes ready for commit.
//
// Returns:
//   - bool: True if there are staged changes, false otherwise.
func hasStagedChanges() bool {
	cmd := exec.Command("git", "diff", "--staged", "--quiet")

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	return cmd.Run() != nil
}

//////
// Git Operations
//////

// gitAddAll stages all changes in the repository.
//
// Returns:
//   - error: Error if staging fails.
func gitAddAll() error {
	return runCommand(exec.Command("git", "add", "."))
}

// gitCommit commits staged changes with the provided message.
//
// Parameters:
//   - message: The commit message to use.
//
// Returns:
//   - error: Error if commit fails.
func gitCommit(message string) error {
	return runCommand(exec.Command("git", "commit", "-m", message))
}

// gitPush pushes commits to the remote repository.
//
// Returns:
//   - error: Error if push fails.
func gitPush() error {
	return runCommand(exec.Command("git", "push"))
}

// gitTag creates a new tag at the current commit.
//
// Parameters:
//   - tag: Name of the tag to create.
//
// Returns:
//   - error: Error if tagging fails.
func gitTag(tag string) error {
	return runCommand(exec.Command("git", "tag", tag))
}

// gitPushTags pushes all tags to the remote repository.
//
// Returns:
//   - error: Error if pushing tags fails.
func gitPushTags() error {
	return runCommand(exec.Command("git", "push", "--tags"))
}

//////
// Git Information Retrieval
//////

// getGitDiff retrieves the staged changes diff.
//
// Returns:
//   - string: The diff output.
//   - error: Error if diff generation fails.
func getGitDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged", "--unified=0")

	out, err := cmd.Output()
	if err != nil {
		fmt.Fprint(os.Stderr, string(out))

		return "", shared.ErrorCatalog.MustGet(
			shared.ErrFailedToGitDiff,
			customerror.WithError(err),
		)
	}

	return string(out), nil
}

// getGitStats retrieves statistics about staged changes.
//
// Returns:
//   - string: The stats output.
//   - error: Error if stats generation fails.
func getGitStats() (string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--stat")

	out, err := cmd.Output()
	if err != nil {
		fmt.Fprint(os.Stderr, string(out))

		return "", shared.ErrorCatalog.MustGet(
			shared.ErrFailedToGitStats,
			customerror.WithError(err),
		)
	}

	return string(out), nil
}

//////
// LLM Integration
//////

// callLLM invokes the LLM API to generate text content.
//
// Parameters:
//   - ctx: Context for the API call.
//   - providerInUse: The LLM provider to use.
//   - prompt: The prompt to send to the LLM.
//
// Returns:
//   - string: The generated text.
//   - error: Error if the API call fails.
func callLLM(
	ctx context.Context,
	providerInUse provider.IProvider,
	prompt string,
) (string, error) {
	contextWithTimeout, cancel := context.WithTimeout(
		ctx,
		llmAPICallTimeout,
	)
	defer cancel()

	return providerInUse.Completion(
		contextWithTimeout,
		provider.WithUserMessages(prompt),
	)
}

//////
// Text Processing
//////

// chunkDiff splits a large diff into smaller chunks.
//
// Parameters:
//   - maxChars: Maximum characters per chunk.
//   - diff: The diff text to split.
//
// Returns:
//   - []string: Array of text chunks.
//   - error: Error if chunking fails.
func chunkDiff(maxChars int, diff string) ([]string, error) {
	if len(diff) <= maxChars {
		return []string{diff}, nil
	}

	splitter := NewTokenSplitter(maxChars)

	chunks, err := splitter.SplitText(diff)
	if err != nil {
		return nil, shared.ErrorCatalog.MustGet(
			shared.ErrFailedToChunkDiff,
			customerror.WithError(err),
		)
	}

	return chunks, nil
}

//////
// Command Execution
//////

// runCommand executes a command and handles its output.
//
// Parameters:
//   - cmd: The command to execute.
//
// Returns:
//   - error: Error if command execution fails.
func runCommand(cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprint(os.Stderr, stderr.String())

		return err
	}

	return nil
}

//////
// File Operations
//////

// OpenFile opens a file for reading.
//
// Parameters:
//   - path: Path to the file.
//
// Returns:
//   - *os.File: The opened file.
//   - error: Error if opening fails.
//
// NOTE: Caller is responsible for closing the file.
func OpenFile(path string) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, shared.ErrorCatalog.MustGet(
			shared.ErrFailedToOpenFile,
			customerror.WithError(err),
		).NewFailedToError()
	}

	return f, nil
}

// ReadFile reads the entire content of a file.
//
// Parameters:
//   - file: The file to read.
//
// Returns:
//   - string: The file content.
//   - error: Error if reading fails.
func ReadFile(file *os.File) (string, error) {
	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", shared.ErrorCatalog.MustGet(
			shared.ErrFailedToReadFile,
			customerror.WithError(err),
		).NewFailedToError()
	}

	return string(bytes), nil
}

//////
// Message Generation
//////

// generateCommitMessageLoop manages the commit message generation workflow.
//
// Parameters:
//   - providerInUse: The LLM provider to use.
//   - stats: Git statistics about the changes.
//   - chunks: Chunks of the diff to process.
//
// Returns:
//   - string: The final commit message.
//   - error: Error if generation fails.
func generateCommitMessageLoop(
	providerInUse provider.IProvider,
	stats string,
	chunks []string,
) (string, error) {
	totalChunks := len(chunks)
	additionalInstructions := ""
	maxAttempts := 5

	for attempt := 0; attempt < maxAttempts; attempt++ {
		for i, chunk := range chunks {
			ctxWithTimeout, cancel := context.WithTimeout(
				context.Background(),
				llmAPICallTimeout,
			)
			defer cancel()

			tea.SpinnerStart("Generating commit message...")

			message, err := generateCommitMessage(
				ctxWithTimeout,
				providerInUse,
				stats,
				chunk,
				i+1,
				totalChunks,
				additionalInstructions,
			)
			if err != nil {
				return "", fmt.Errorf(
					"failed to generate commit message: %w",
					err,
				)
			}

			tea.SpinnerStop()

			fmt.Printf(
				"%s\n\n%s\n\n",
				tea.QuestionStyle.Render("Generated Commit Message:"),
				message,
			)

			choice := tea.PromptWithChoices(
				"What would you like to do?",
				[]string{
					"Approve commit message",
					"Try again",
					"Write commit message yourself",
					"Exit",
				},
			)

			switch choice {
			case "Approve commit message":
				return message, nil
			case "Try again":
				additionalInstructions = handleTryAgain()

				break
			case "Write commit message yourself":
				content, err := tea.NewMessageTextArea()
				if err != nil {
					return "", fmt.Errorf(
						"failed to get commit message: %w",
						err,
					)
				}

				return content + "\n", nil
			case "Exit":
				fmt.Println("Nothing to do, exiting...")

				os.Exit(0)
			}

			break // Break inner loop if "Try again" was selected
		}
	}

	return "", fmt.Errorf("maximum attempts reached")
}

// generateCommitMessage creates a commit message using the LLM API.
//
// Parameters:
//   - ctx: Context for the API call.
//   - providerInUse: The LLM provider to use.
//   - stats: Git statistics about the changes.
//   - diff: The diff content.
//   - chunkNumber: Current chunk number.
//   - totalChunks: Total number of chunks.
//   - additionalInstructions: Extra instructions for the LLM.
//
// Returns:
//   - string: The generated commit message.
//   - error: Error if generation fails.
func generateCommitMessage(
	ctx context.Context,
	providerInUse provider.IProvider,
	stats, diff string,
	chunkNumber, totalChunks int,
	additionalInstructions string,
) (string, error) {
	var prompt string

	if totalChunks > 1 {
		prompt = buildChunkedPrompt(
			stats,
			diff,
			chunkNumber,
			totalChunks,
			additionalInstructions,
		)
	} else {
		prompt = buildStandardPrompt(
			stats,
			diff,
			additionalInstructions,
		)
	}

	cliLogger.Tracelnf("Prompting LLM with the following prompt:\n%s", prompt)

	return callLLM(ctx, providerInUse, prompt)
}

// buildChunkedPrompt creates a prompt for chunked diffs.
//
// Parameters:
//   - stats: Git statistics about the changes.
//   - diff: The diff content.
//   - chunkNumber: Current chunk number.
//   - totalChunks: Total number of chunks.
//   - additionalInstructions: Extra instructions for the LLM.
//
// Returns:
//   - string: The formatted prompt.
func buildChunkedPrompt(
	stats, diff string,
	chunkNumber, totalChunks int,
	additionalInstructions string,
) string {
	return fmt.Sprintf(
		`Please generate a concise and descriptive commit message based on the 
following staged changes. Note that the diff is too big, so we chunked it into
smaller parts:

Change Statistics:
%s

Chunk %d of %d:
%s

%s`,
		stats,
		chunkNumber,
		totalChunks,
		diff,
		additionalInstructions,
	)
}

// buildStandardPrompt creates a prompt for standard diffs.
//
// Parameters:
//   - stats: Git statistics about the changes.
//   - diff: The diff content.
//   - additionalInstructions: Extra instructions for the LLM.
//
// Returns:
//   - string: The formatted prompt.
func buildStandardPrompt(
	stats, diff string,
	additionalInstructions string,
) string {
	return fmt.Sprintf(
		`Please generate a concise and descriptive commit message based on the 
following staged changes:

Change Statistics:
%s

Code Changes:
%s

%s`,
		stats,
		diff,
		additionalInstructions,
	)
}

// handleTryAgain manages the retry flow for commit message generation.
//
// Returns:
//   - string: Additional instructions based on user choice.
func handleTryAgain() string {
	changeChoice := tea.PromptWithChoices(
		"What would you like to change?",
		[]string{
			"Make more succinct",
			"Make more technical",
			"Make less technical",
			"Write what should change",
		},
	)

	return getChangeInstructions(changeChoice)
}

// getChangeInstructions returns instructions based on the change type.
//
// Parameters:
//   - changeType: The type of change requested.
//
// Returns:
//   - string: The corresponding instructions.
func getChangeInstructions(changeType string) string {
	switch changeType {
	case "Make more succinct":
		return "Please make the commit message more succinct while still " +
			"conveying the essence of the change."
	case "Make more technical":
		return buildTechnicalInstructions()
	case "Make less technical":
		return buildNonTechnicalInstructions()
	case "Write what should change":
		return tea.PromptForInputTea("Describe what should change:")
	default:
		return ""
	}
}

// buildTechnicalInstructions creates instructions for technical commit messages.
//
// Returns:
//   - string: Detailed technical instructions.
func buildTechnicalInstructions() string {
	return `Please make the commit message more technical, adding IF POSSIBLE, 
more context and details aiding engineering comprehension:

1. Include relevant technical terms, e.g., function names, data structures, or 
   algorithms modified.
2. Specify the exact files or modules affected.
3. For bug fixes, IF possible, briefly describe the root cause and solution.
4. For new features, IF possible, outline the core implementation approach.
5. Use concise language while maintaining technical accuracy.

Examples:
- "Optimized database query in user_auth.py using indexing"
- "Implemented red-black tree for efficient sorting in data_processor.cpp"
- "Fixed race condition in thread pool by adding mutex lock in worker.java"

Aim for a balance between technical depth and clarity. Prioritize information 
that aids code review and future maintenance. No more than 1000 characters!`
}

// buildNonTechnicalInstructions creates instructions for non-technical messages.
//
// Returns:
//   - string: Non-technical formatting instructions.
func buildNonTechnicalInstructions() string {
	return `Please make commit messages non-technical, suitable for general 
audiences. Aim for brevity while still conveying the essence of the change. 
Examples:

- For updating dependencies: "Updated dependencies"
- For fixing a bug: "Fixed login issue"
- For adding a feature: "Added dark mode"
- For refactoring: "Improved code structure"

For complex changes, summarize the overall impact rather than listing technical 
details. If multiple significant changes are present, use a bulleted list.`
}

// generateDocumentation generates documentation for the codebase.
//
// Parameters:
//   - ctx: Context for the API call.
//   - providerInUse: The LLM provider to use.
//   - content: The codebase content.
//   - chunkNumber: Current chunk number.
//   - totalChunks: Total number of chunks.
//
// Returns:
//   - string: The generated documentation.
//   - error: Error if generation fails.
func generateDocumentation(
	ctx context.Context,
	providerInUse provider.IProvider,
	content string,
	chunkNumber, totalChunks int,
) (string, error) {
	prompt := buildDocumentationPrompt(content, chunkNumber, totalChunks)

	cliLogger.Tracelnf("Prompting LLM with the following prompt:\n%s", prompt)

	return callLLM(ctx, providerInUse, prompt)
}

// buildDocumentationPrompt creates a prompt for documentation generation.
//
// Parameters:
//   - content: The codebase content.
//   - chunkNumber: Current chunk number.
//   - totalChunks: Total number of chunks.
//
// Returns:
//   - string: The formatted documentation prompt.
func buildDocumentationPrompt(
	content string,
	chunkNumber, totalChunks int,
) string {
	template := `Please generate a markdown document, based on the following 
codebase. The document must contain the following sections:
- An overview of the what is the codebase, for example: "a web application that 
  allows users to send email campaigns."
- An exhaustive list of features, high-level, each feature description with no 
  more than 240 characters, for example: "Users can create and send email 
  campaigns", "Users can track the performance of their campaigns.", and "A 
  dashboard with graphs and metrics provides insights on campaign performance."
- Architecture overview, for example: "The application is built using a 
  microservices architecture, with a React frontend and a Go backend. Run on 
  Docker. It uses a PostgreSQL database, Redis for caching and Elasticsearch for 
  search."

%s`

	if totalChunks > 1 {
		return fmt.Sprintf(
			template,
			fmt.Sprintf(
				"Note that the codebase is too big, so we chunked it into "+
					"smaller parts:\n\nChunk %d of %d:\n%s",
				chunkNumber,
				totalChunks,
				content,
			),
		)
	}

	return fmt.Sprintf(template, content)
}
