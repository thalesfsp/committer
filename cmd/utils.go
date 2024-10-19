package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/thalesfsp/customerror"
	"github.com/thalesfsp/inference/provider"
)

// isDebugMode checks if the CLI is running in debug mode.
func isDebugMode() bool {
	syplLevel := os.Getenv("SYPL_LEVEL")

	return syplLevel == "debug"
}

// isCurrentDirectoryGitRepo checks if the current directory is a Git repository.
func isCurrentDirectoryGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")

	return cmd.Run() == nil
}

// hasStagedChanges checks if there are any staged changes.
func hasStagedChanges() bool {
	cmd := exec.Command("git", "diff", "--staged", "--quiet")

	return cmd.Run() != nil
}

// gitAddAll adds all changes to staging.
func gitAddAll() error {
	cmd := exec.Command("git", "add", ".")

	return cmd.Run()
}

// getGitDiff gets the staged diff.
func getGitDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged", "--unified=0")

	out, err := cmd.Output()
	if err != nil {
		return "", ErrorCatalog.MustGet(ErrFailedToGitDiff, customerror.WithError(err))
	}

	return string(out), nil
}

// getGitStats gets the git stats.
func getGitStats() (string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--stat")

	out, err := cmd.Output()
	if err != nil {
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

// generateCommitMessage generates a commit message using LLM API.
//
//nolint:lll
func generateCommitMessage(
	ctx context.Context,
	providerInUse provider.IProvider,
	stats, diff string,
	chunkNumber, totalChunks int,
) (string, error) {
	var prompt string
	if totalChunks > 1 {
		// Diff is chunked
		prompt = fmt.Sprintf(`Please generate a concise and descriptive commit message based on the following staged changes. Note that the diff is too big, so we chunked it into smaller parts:

Change Statistics:
%s

Chunk %d of %d:
%s`, stats, chunkNumber, totalChunks, diff)
	} else {
		// Diff is not chunked; use standard prompt
		prompt = fmt.Sprintf(`Please generate a concise and descriptive commit message based on the following staged changes:

Change Statistics:
%s

Code Changes:
%s`, stats, diff)
	}

	cliLogger.Debuglnf("Prompting LLM with the following prompt:\n%s", prompt)

	// Call LLM API
	message, err := callLLM(ctx, providerInUse, prompt)
	if err != nil {
		return "", err
	}

	return message, nil
}

// callLLM calls the LLM API (OpenAI or Ollama).
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
//
// TODO: Output to /dev/null.
func gitCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)

	cmd.Stdout = os.Stdout

	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// gitPush pushes the commits.
func gitPush() error {
	cmd := exec.Command("git", "push")

	cmd.Stdout = os.Stdout

	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// gitTag tags the commit with the provided tag name.
func gitTag(tag string) error {
	cmd := exec.Command("git", "tag", tag)

	return cmd.Run()
}

// gitPushTags pushes the tags to the remote repository.
func gitPushTags() error {
	cmd := exec.Command("git", "push", "--tags")

	cmd.Stdout = os.Stdout

	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// promptYesNo prompts the user with a yes/no question.
func promptYesNo(question string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("\n%s (y/n): ", question)

		response, _ := reader.ReadString('\n')

		response = strings.TrimSpace(strings.ToLower(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

// promptForInput prompts the user for input.
func promptForInput(prompt string) string {
	fmt.Println(prompt)

	reader := bufio.NewReader(os.Stdin)

	input, err := reader.ReadString('\n')
	if err != nil {
		cliLogger.Fatalln(err)
	}

	return strings.TrimSpace(input)
}
