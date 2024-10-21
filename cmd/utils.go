package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/thalesfsp/customerror"
	"github.com/thalesfsp/inference/provider"
)

var syplLevel = os.Getenv("SYPL_LEVEL")

// isDebugMode checks if the CLI is running in debug mode.
func isDebugMode() bool {
	return syplLevel == "debug"
}

// isCurrentDirectoryGitRepo checks if the current directory is a Git repository.
func isCurrentDirectoryGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Fprint(os.Stderr, stderr.String())

		return false
	}
	return true
}

// isDirty checks if there are any uncommitted changes.
func isDirty() bool {
	cmd := exec.Command("git", "diff", "--quiet")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return true
	}
	return false
}

// hasStagedChanges checks if there are any staged changes.
func hasStagedChanges() bool {
	cmd := exec.Command("git", "diff", "--staged", "--quiet")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// If there are staged changes, git diff --quiet exits with status 1.
		// So we return true without printing the error.
		return true
	}
	return false
}

// gitAddAll adds all changes to staging.
func gitAddAll() error {
	cmd := exec.Command("git", "add", ".")

	return runCommand(cmd)
}

// getGitDiff gets the staged diff.
func getGitDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged", "--unified=0")

	out, err := cmd.Output()
	if err != nil {
		fmt.Fprint(os.Stderr, string(out))

		return "", ErrorCatalog.MustGet(ErrFailedToGitDiff, customerror.WithError(err))
	}

	return string(out), nil
}

// getGitStats gets the git stats.
func getGitStats() (string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--stat")

	out, err := cmd.Output()
	if err != nil {
		fmt.Fprint(os.Stderr, string(out))

		return "", ErrorCatalog.MustGet(ErrFailedToGitStats, customerror.WithError(err))
	}

	return string(out), nil
}

// chunkDiff chunks the diff if it's too big.
func chunkDiff(maxChars int, diff string) ([]string, error) {
	if len(diff) <= maxChars {
		return []string{diff}, nil
	}

	splitter := NewTokenSplitter()

	chunks, err := splitter.SplitText(diff)
	if err != nil {
		return nil, ErrorCatalog.MustGet(ErrFailedToChunkDiff, customerror.WithError(err))
	}

	return chunks, nil
}

// callLLM calls the LLM API (OpenAI, Anthropic, or Ollama).
func callLLM(
	ctx context.Context,
	providerInUse provider.IProvider,
	prompt string,
) (string, error) {
	response, err := providerInUse.Completion(
		ctx,
		provider.WithModel(model),
		provider.WithUserMessages(prompt),
	)
	if err != nil {
		return "", err
	}

	return response, nil
}

// gitCommit commits the changes with the provided message.
func gitCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	return runCommand(cmd)
}

// gitPush pushes the commits.
func gitPush() error {
	cmd := exec.Command("git", "push")
	return runCommand(cmd)
}

// gitTag tags the commit with the provided tag name.
func gitTag(tag string) error {
	cmd := exec.Command("git", "tag", tag)
	return runCommand(cmd)
}

// gitPushTags pushes the tags to the remote repository.
func gitPushTags() error {
	cmd := exec.Command("git", "push", "--tags")
	return runCommand(cmd)
}

// runCommand executes a command and outputs errors to os.Stderr if any.
func runCommand(cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Fprint(os.Stderr, stderr.String())
		return err
	}
	return nil
}

func handleTryAgain() string {
	changeChoice := promptWithChoices("What would you like to change?", []string{
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
		return promptForInputTea("Describe what should change:")
	default:
		return ""
	}
}

func generateCommitMessageLoop(providerInUse provider.IProvider, stats string, chunks []string) (string, error) {
	totalChunks := len(chunks)
	additionalInstructions := ""
	maxAttempts := 5 // Define a maximum number of attempts to prevent infinite loops

	for attempt := 0; attempt < maxAttempts; attempt++ {
		for i, chunk := range chunks {
			ctxWithTimeout, cancel := context.WithTimeout(context.Background(), llmAPICallTimeout)

			spinnerStart("Generating commit message...")
			message, err := generateCommitMessage(
				ctxWithTimeout,
				providerInUse,
				stats, chunk,
				i+1, totalChunks,
				additionalInstructions,
			)
			cancel() // Cancel the context right after use

			if err != nil {
				return "", fmt.Errorf("failed to generate commit message: %w", err)
			}

			spinnerStop()

			fmt.Printf("%s\n\n%s\n\n", questionStyle.Render("Generated Commit Message:"), message)

			choice := promptWithChoices("What would you like to do?", []string{
				"Approve commit message",
				"Try again",
				"Write commit message yourself",
				"Exit",
			})

			switch choice {
			case "Approve commit message":
				return message, nil
			case "Try again":
				additionalInstructions = handleTryAgain()
				break // Break inner loop to regenerate
			case "Write commit message yourself":
				content, err := commitMessageTextArea()
				if err != nil {
					return "", fmt.Errorf("failed to get commit message: %w", err)
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

// generateCommitMessage generates a commit message using LLM API with additional instructions.
func generateCommitMessage(
	ctx context.Context,
	providerInUse provider.IProvider,
	stats, diff string,
	chunkNumber, totalChunks int,
	additionalInstructions string,
) (string, error) {
	var prompt string
	if totalChunks > 1 {
		// Diff is chunked
		prompt = fmt.Sprintf(`Please generate a concise and descriptive commit message based on the following staged changes. Note that the diff is too big, so we chunked it into smaller parts:

Change Statistics:
%s

Chunk %d of %d:
%s

%s`, stats, chunkNumber, totalChunks, diff, additionalInstructions)
	} else {
		// Diff is not chunked; use standard prompt
		prompt = fmt.Sprintf(`Please generate a concise and descriptive commit message based on the following staged changes:

Change Statistics:
%s

Code Changes:
%s

%s`, stats, diff, additionalInstructions)
	}

	cliLogger.Tracelnf("Prompting LLM with the following prompt:\n%s", prompt)

	// Call LLM API
	message, err := callLLM(ctx, providerInUse, prompt)
	if err != nil {
		return "", err
	}

	return message, nil
}

// commitMessageTextArea is a Tea model for the commit message text area.
func commitMessageTextArea() (string, error) {
	p := tea.NewProgram(initializeTextAreModel())
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	// Check if the program finished due to Ctrl+Enter
	if m.(textAreModel).done {
		return m.(textAreModel).textarea.Value(), nil
	}

	return "", nil
}

// promptYesNoTea prompts a yes/no question using Tea.
func promptYesNoTea(question string, defaultChoice bool) bool {
	choices := []string{"Yes", "No"}

	defaultIndex := 1
	if defaultChoice {
		defaultIndex = 0
	}

	m := cliModel{
		question:      question,
		choices:       choices,
		defaultChoice: defaultIndex,
		cursor:        defaultIndex, // Set initial cursor to default choice
	}

	p := tea.NewProgram(m)

	model, err := p.Run()
	if err != nil {
		cliLogger.Fatalln(ErrorCatalog.
			MustGet(ErrFailedToInitTea).
			NewFailedToError(customerror.WithError(err)),
		)
	}

	if m, ok := model.(cliModel); ok && m.choice != "" {
		return m.choice == "Yes"
	}

	return false
}

// promptForInputTea prompts the user for input using Tea.
func promptForInputTea(prompt string) string {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.Prompt = inputStyle.Render("> ")

	m := inputModel{
		textinput: ti,
		prompt:    prompt,
	}

	p := tea.NewProgram(m)
	model, err := p.Run()
	if err != nil {
		cliLogger.Fatalln(ErrorCatalog.
			MustGet(ErrFailedToInitTea).
			NewFailedToError(customerror.WithError(err)),
		)
	}

	if m, ok := model.(inputModel); ok && m.input != "" {
		return m.input
	}
	return ""
}

// promptWithChoices prompts the user with multiple choices using Tea.
func promptWithChoices(question string, choices []string) string {
	m := cliModel{
		question: question,
		choices:  choices,
	}

	p := tea.NewProgram(m)
	model, err := p.Run()
	if err != nil {
		cliLogger.Fatalln(ErrorCatalog.
			MustGet(ErrFailedToInitTea).
			NewFailedToError(customerror.WithError(err)),
		)
	}

	if m, ok := model.(cliModel); ok && m.choice != "" {
		return m.choice
	}
	return ""
}
