package cmd

import (
	"github.com/thalesfsp/customerror"
)

const (
	ErrFailedToCallLLM          = "ERR_FAILED_TO_CALL_LLM"           // FailedTo.
	ErrFailedToChunkDiff        = "ERR_FAILED_TO_CHUNK_DIFF"         // FailedTo.
	ErrFailedToCreateHTTPClient = "ERR_FAILED_TO_CREATE_HTTP_CLIENT" // FailedTo.
	ErrFailedToGitDiff          = "ERR_FAILED_TO_GIT_DIFF"           // FailedTo.
	ErrFailedToGitStats         = "ERR_FAILED_TO_GIT_STATS"          // FailedTo.
	ErrFailedToSetupLLM         = "ERR_FAILED_TO_SETUP_LLM"          // FailedTo.
	ErrFailedToStageFiles       = "ERR_FAILED_TO_STAGE_FILES"        // FailedTo.
	ErrInvalidProvider          = "ERR_INVALID_PROVIDER"             // Invalid.
	ErrNotGitRepo               = "ERR_NOT_GIT_REPO"                 // Required.
)

// ErrorCatalog is the error catalog for the CLI.
var ErrorCatalog = customerror.
	MustNewCatalog(Name).
	MustSet(ErrFailedToCallLLM, "call LLM API").
	MustSet(ErrFailedToChunkDiff, "chunk diff").
	MustSet(ErrFailedToCreateHTTPClient, "create HTTP client").
	MustSet(ErrFailedToGitDiff, "obtain git diff").
	MustSet(ErrFailedToGitStats, "obtain git stats").
	MustSet(ErrFailedToSetupLLM, "setup LLM API").
	MustSet(ErrFailedToStageFiles, "stage files").
	MustSet(ErrInvalidProvider, "provider").
	MustSet(ErrNotGitRepo, "current directory is not a git repository")
