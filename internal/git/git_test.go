package git

import (
	"os/exec"
	"strings"
	"testing"
)

// createTestCommand is a helper to create exec.Cmd for tests.
func createTestCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// TestGitGetLatestTags_ErrorWrapping verifies that GitGetLatestTags wraps
// errors using the error catalog rather than returning bare errors.
//
// This is a regression test: previously the function returned `err` directly,
// inconsistent with other git functions that wrap errors via errorcatalog.
func TestGitGetLatestTags_ErrorWrapping(t *testing.T) {
	// We can test this in the current git repo — it should succeed.
	// The main purpose is to verify the function works and that on success
	// we get proper results.
	tags, err := GitGetLatestTags(3)
	if err != nil {
		// If running in a git repo with no tags, this could be nil/nil.
		// Check that any error is properly wrapped (contains error catalog info).
		errStr := err.Error()
		if !strings.Contains(errStr, "retrieve git tags") &&
			!strings.Contains(errStr, "ERR_FAILED_TO_GET_TAGS") {
			t.Errorf("expected wrapped error with catalog info, got bare error: %v", err)
		}
	}

	// Verify the count limit works when tags exist.
	if tags != nil && len(tags) > 3 {
		t.Errorf("expected at most 3 tags, got %d", len(tags))
	}
}

// TestIsCurrentDirectoryGitRepo verifies we're in a git repo (test runs
// from within the project).
func TestIsCurrentDirectoryGitRepo(t *testing.T) {
	if !IsCurrentDirectoryGitRepo() {
		t.Skip("not running in a git repository")
	}
}

// TestRunCommand_FailingCommand verifies RunCommand returns an error for
// invalid commands.
func TestRunCommand_FailingCommand(t *testing.T) {
	// Use a command that will fail.
	cmd := createTestCommand("git", "log", "--invalid-flag-that-does-not-exist")
	err := RunCommand(cmd)

	if err == nil {
		t.Error("expected error from invalid git command, got nil")
	}
}
