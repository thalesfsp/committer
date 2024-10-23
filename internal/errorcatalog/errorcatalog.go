package errorcatalog

import (
	"github.com/thalesfsp/committer/internal/shared"
	"github.com/thalesfsp/customerror"
)

//////
// Const, vars, types.
//////

const (
	ErrEmptyCommitMessage       = "ERR_EMPTY_COMMIT_MESSAGE"         // Missing.
	ErrFailedToCallLLM          = "ERR_FAILED_TO_CALL_LLM"           // FailedTo.
	ErrFailedToChunkDiff        = "ERR_FAILED_TO_CHUNK_DIFF"         // FailedTo.
	ErrFailedToCreateHTTPClient = "ERR_FAILED_TO_CREATE_HTTP_CLIENT" // FailedTo.
	ErrFailedToGitDiff          = "ERR_FAILED_TO_GIT_DIFF"           // FailedTo.
	ErrFailedToGitStats         = "ERR_FAILED_TO_GIT_STATS"          // FailedTo.
	ErrFailedToInitChunker      = "ERR_FAILED_TO_INIT_CHUNKER"       // FailedTo.
	ErrFailedToInitTea          = "ERR_FAILED_TO_INIT_TEA"           // FailedTo.
	ErrFailedToRunTeaProgram    = "ERR_FAILED_TO_RUN_TEA_PROGRAM"    // FailedTo.
	ErrFailedToSetupLLM         = "ERR_FAILED_TO_SETUP_LLM"          // FailedTo.
	ErrFailedToStageFiles       = "ERR_FAILED_TO_STAGE_FILES"        // FailedTo.
	ErrInvalidProvider          = "ERR_INVALID_PROVIDER"             // Invalid.
	ErrNotGitRepo               = "ERR_NOT_GIT_REPO"                 // Required.
)

// errorCatalog is the error catalog for the CLI.
var errorCatalog = customerror.
	MustNewCatalog(shared.Name).
	MustSet(ErrEmptyCommitMessage, "commit message").
	MustSet(ErrFailedToCallLLM, "call LLM API").
	MustSet(ErrFailedToChunkDiff, "chunk diff").
	MustSet(ErrFailedToCreateHTTPClient, "create HTTP client").
	MustSet(ErrFailedToGitDiff, "obtain git diff").
	MustSet(ErrFailedToGitStats, "obtain git stats").
	MustSet(ErrFailedToInitChunker, "initialize chunker").
	MustSet(ErrFailedToInitTea, "initialize Tea application").
	MustSet(ErrFailedToRunTeaProgram, "run Tea program").
	MustSet(ErrFailedToSetupLLM, "setup LLM API").
	MustSet(ErrFailedToStageFiles, "stage files").
	MustSet(ErrInvalidProvider, "provider").
	MustSet(ErrNotGitRepo, "current directory is not a git repository")

//////
// Exported functionalities.
//////

// MustGet returns a custom error from the error catalog.
func MustGet(errorCode string, opts ...customerror.Option) *customerror.CustomError {
	return errorCatalog.MustGet(errorCode, opts...)
}
