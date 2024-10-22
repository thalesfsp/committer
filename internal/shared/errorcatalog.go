package shared

import (
	"github.com/thalesfsp/customerror"
)

const (
	ErrEmptyCommitMessage        = "ERR_EMPTY_COMMIT_MESSAGE"         // Missing.
	ErrFailedToCallLLM           = "ERR_FAILED_TO_CALL_LLM"           // FailedTo.
	ErrFailedToChunkDiff         = "ERR_FAILED_TO_CHUNK_DIFF"         // FailedTo.
	ErrFailedToCreateHTTPClient  = "ERR_FAILED_TO_CREATE_HTTP_CLIENT" // FailedTo.
	ErrFailedToGitDiff           = "ERR_FAILED_TO_GIT_DIFF"           // FailedTo.
	ErrFailedToGitStats          = "ERR_FAILED_TO_GIT_STATS"          // FailedTo.
	ErrFailedToInitChunker       = "ERR_FAILED_TO_INIT_CHUNKER"       // FailedTo.
	ErrFailedToInitTea           = "ERR_FAILED_TO_INIT_TEA"           // FailedTo.
	ErrFailedToOpenFile          = "ERR_FAILED_TO_OPEN_FILE"          // FailedTo.
	ErrFailedToReadFile          = "ERR_FAILED_TO_READ_FILE"          // FailedTo.
	ErrFailedToSaveDocumentation = "ERR_FAILED_TO_SAVE_DOCUMENTATION" // FailedTo.
	ErrFailedToSetupLLM          = "ERR_FAILED_TO_SETUP_LLM"          // FailedTo.
	ErrFailedToStageFiles        = "ERR_FAILED_TO_STAGE_FILES"        // FailedTo.
	ErrInvalidProvider           = "ERR_INVALID_PROVIDER"             // Invalid.
	ErrNotGitRepo                = "ERR_NOT_GIT_REPO"                 // Required.
)

// ErrorCatalog is the error catalog for the CLI.
var ErrorCatalog = customerror.
	MustNewCatalog(Name).
	MustSet(ErrEmptyCommitMessage, "commit message").
	MustSet(ErrFailedToCallLLM, "call LLM API").
	MustSet(ErrFailedToChunkDiff, "chunk diff").
	MustSet(ErrFailedToCreateHTTPClient, "create HTTP client").
	MustSet(ErrFailedToGitDiff, "obtain git diff").
	MustSet(ErrFailedToGitStats, "obtain git stats").
	MustSet(ErrFailedToInitChunker, "initialize chunker").
	MustSet(ErrFailedToInitTea, "initialize Tea application").
	MustSet(ErrFailedToOpenFile, "open file").
	MustSet(ErrFailedToReadFile, "read file").
	MustSet(ErrFailedToSaveDocumentation, "save documentation").
	MustSet(ErrFailedToSetupLLM, "setup LLM API").
	MustSet(ErrFailedToStageFiles, "stage files").
	MustSet(ErrInvalidProvider, "provider").
	MustSet(ErrNotGitRepo, "current directory is not a git repository")
