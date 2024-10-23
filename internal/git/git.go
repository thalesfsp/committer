package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/thalesfsp/committer/internal/errorcatalog"
	"github.com/thalesfsp/customerror"
)

// IsCurrentDirectoryGitRepo determines if the current directory is a Git repository.
// It does this by attempting to run 'git rev-parse --is-inside-work-tree' which
// returns true if the directory is part of a Git repository work tree.
func IsCurrentDirectoryGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// If the command returns an error, the current directory is not a Git repo.
	if err := cmd.Run(); err != nil {
		// Output the standard error to the console and return false.
		fmt.Fprint(os.Stderr, stderr.String())

		return false
	}

	return true
}

// IsDirty checks if there are any uncommitted changes in the working directory.
// 'git diff --quiet' will return a non-zero exit code if there are changes, so
// this function returns true in that case.
func IsDirty() bool {
	cmd := exec.Command("git", "diff", "--quiet")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// If there is an error, it implies there are changes, hence returns true.
	if err := cmd.Run(); err != nil {
		return true
	}

	return false
}

// HasStagedChanges checks if there are any staged but not yet committed changes.
// Uses 'git diff --staged --quiet', which returns error if there are changes.
func HasStagedChanges() bool {
	cmd := exec.Command("git", "diff", "--staged", "--quiet")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Returns true for any non-zero exit, meaning there are staged changes.
	if err := cmd.Run(); err != nil {
		return true
	}

	return false
}

// GitAddAll stages all changes (including new, modified, and deleted files).
// It runs the command 'git add .' which adds all files in the current directory.
func GitAddAll() error {
	return RunCommand(exec.Command("git", "add", "."))
}

// GetGitDiff retrieves the staged differences.
// This function runs 'git diff --staged --unified=0' to show zero lines of
// context around differences in the output. The diff is returned as a string.
func GetGitDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged", "--unified=0")

	out, err := cmd.Output()
	if err != nil {
		fmt.Fprint(os.Stderr, string(out))

		return "", errorcatalog.MustGet(errorcatalog.ErrFailedToGitDiff, customerror.WithError(err))
	}

	return string(out), nil
}

// GetGitStats provides statistics of staged changes. It uses the command
// 'git diff --cached --stat' to show file statistics (insertions, deletions)
// for staged changes.
func GetGitStats() (string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--stat")

	out, err := cmd.Output()
	if err != nil {
		fmt.Fprint(os.Stderr, string(out))

		return "", errorcatalog.MustGet(errorcatalog.ErrFailedToGitStats, customerror.WithError(err))
	}

	return string(out), nil
}

// GitCommit commits staged changes with a provided commit message.
// Uses 'git commit -m <message>' to perform a commit.
func GitCommit(message string) error {
	return RunCommand(exec.Command("git", "commit", "-m", message))
}

// GitPush pushes commits to the remote repository.
// Runs 'git push' to push changes to the default push target.
func GitPush() error {
	return RunCommand(exec.Command("git", "push"))
}

// GitTag creates a new tag on the latest commit with the specified tag name.
// Uses 'git tag <tag>' to attach a tag to the current commit.
func GitTag(tag string) error {
	return RunCommand(exec.Command("git", "tag", tag))
}

// GitPushTags pushes tags to the remote repository.
// Executes 'git push --tags' to push all tags to the remote.
func GitPushTags() error {
	return RunCommand(exec.Command("git", "push", "--tags"))
}

// RunCommand executes a given command and outputs its standard error content to
// os.Stderr if the command fails. This is a helper function to reduce repetition
// of error handling logic.
func RunCommand(cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// If the command fails, print the standard error to the console and return
	// the error.
	if err := cmd.Run(); err != nil {
		fmt.Fprint(os.Stderr, stderr.String())

		return err
	}

	return nil
}
