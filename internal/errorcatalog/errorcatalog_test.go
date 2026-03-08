package errorcatalog

import (
	"testing"
)

// TestErrorCatalog_ErrFailedToGetTags verifies the error catalog contains
// the new ErrFailedToGetTags entry for consistent error wrapping in git
// tag operations.
func TestErrorCatalog_ErrFailedToGetTags(t *testing.T) {
	err := MustGet(ErrFailedToGetTags)
	if err == nil {
		t.Fatal("expected non-nil error from catalog for ErrFailedToGetTags")
	}

	errStr := err.Error()
	if errStr == "" {
		t.Error("expected non-empty error string")
	}
}

// TestErrorCatalog_AllEntriesExist verifies all error catalog constants can
// be retrieved without panicking.
func TestErrorCatalog_AllEntriesExist(t *testing.T) {
	entries := []string{
		ErrEmptyCommitMessage,
		ErrFailedToCallLLM,
		ErrFailedToChunkDiff,
		ErrFailedToCreateHTTPClient,
		ErrFailedToGetTags,
		ErrFailedToGitDiff,
		ErrFailedToGitStats,
		ErrFailedToInitChunker,
		ErrFailedToInitTea,
		ErrFailedToRunTeaProgram,
		ErrFailedToSetupLLM,
		ErrFailedToStageFiles,
		ErrInvalidProvider,
		ErrNotGitRepo,
	}

	for _, code := range entries {
		t.Run(code, func(t *testing.T) {
			err := MustGet(code)
			if err == nil {
				t.Errorf("MustGet(%q) returned nil", code)
			}
		})
	}
}
