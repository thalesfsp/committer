/*
Token-based Text Splitter

This file implements a token-based text splitter that divides text into chunks
based on token count. It's particularly useful for processing large texts that
need to be sent to language models with token limits.

Flow:
graph TD
    A[Input Text] --> B[Initialize Tokenizer]
    B --> C[Encode Text to Tokens]
    C --> D[Split Tokens into Chunks]
    D --> E[Decode Chunks to Text]
    E --> F[Return Text Chunks]

    subgraph "Chunk Processing"
        D --> G[Calculate Chunk Size]
        G --> H[Apply Overlap]
        H --> I[Create Chunk]
    end

NOTE:
The splitter uses the tiktoken-go library for tokenization, which implements the
same tokenization schemes used by various language models. This ensures accurate
token counting and splitting.
*/

package textsplitter

import (
	"github.com/pkoukk/tiktoken-go"
	"github.com/thalesfsp/committer/internal/errorcatalog"
	"github.com/thalesfsp/customerror"
)

//////
// Constants
//////

// Default configuration values for token splitting.
const (
	// DefaultTokenChunkOverlap specifies the default number of overlapping tokens
	// between consecutive chunks.
	DefaultTokenChunkOverlap = 100

	// DefaultTokenChunkSize defines the default maximum number of tokens per
	// chunk.
	DefaultTokenChunkSize = 4096

	// DefaultTokenEncoding specifies the default encoding scheme for
	// tokenization.
	DefaultTokenEncoding = "cl100k_base"

	// DefaultTokenModelName defines the default model name used for
	// tokenization.
	DefaultTokenModelName = "gpt-3.5-turbo"
)

//////
// Types
//////

// TokenSplitter implements text splitting functionality based on token count.
// It provides configuration options for chunk size, overlap, model selection,
// and special token handling.
type TokenSplitter struct {
	// AllowedSpecial defines the list of allowed special tokens.
	AllowedSpecial []string

	// ChunkOverlap specifies the number of tokens to overlap between
	// consecutive chunks.
	ChunkOverlap int

	// ChunkSize defines the maximum number of tokens per chunk.
	ChunkSize int

	// DisallowedSpecial defines the list of disallowed special tokens.
	DisallowedSpecial []string

	// EncodingName specifies the encoding scheme to use for tokenization.
	EncodingName string

	// ModelName defines the model to determine token encoding.
	ModelName string
}

//////
// Helpers.
//////

// initializeTokenizer sets up the tokenizer based on the configuration.
func (s TokenSplitter) initializeTokenizer() (*tiktoken.Tiktoken, error) {
	if s.EncodingName != "" {
		return tiktoken.GetEncoding(s.EncodingName)
	}

	return tiktoken.EncodingForModel(s.ModelName)
}

// splitTextIntoChunks performs the actual text splitting operation.
func (s TokenSplitter) splitTextIntoChunks(
	text string,
	tk *tiktoken.Tiktoken,
) []string {
	// Initialize the results slice.
	chunks := make([]string, 0)

	// Encode the input text into token IDs.
	tokenIDs := tk.Encode(text, s.AllowedSpecial, s.DisallowedSpecial)

	// Process the text in chunks.
	startIdx := 0
	endIdx := s.calculateInitialEndIndex(len(tokenIDs))

	for startIdx < len(tokenIDs) {
		// Extract and decode the current chunk.
		chunk := s.processChunk(tokenIDs[startIdx:endIdx], tk)
		chunks = append(chunks, chunk)

		// Calculate indices for the next chunk.
		startIdx, endIdx = s.calculateNextIndices(startIdx, len(tokenIDs))
	}

	return chunks
}

// calculateInitialEndIndex determines the initial ending index for the first
// chunk.
func (s TokenSplitter) calculateInitialEndIndex(totalTokens int) int {
	if s.ChunkSize < totalTokens {
		return s.ChunkSize
	}

	return totalTokens
}

// calculateNextIndices computes the start and end indices for the next chunk.
func (s TokenSplitter) calculateNextIndices(
	currentStart, totalTokens int,
) (int, int) {
	// Calculate the next starting point, accounting for overlap.
	nextStart := currentStart + s.ChunkSize - s.ChunkOverlap

	// Calculate the next ending point.
	nextEnd := nextStart + s.ChunkSize
	if nextEnd > totalTokens {
		nextEnd = totalTokens
	}

	return nextStart, nextEnd
}

// processChunk converts a slice of token IDs back into text.
func (s TokenSplitter) processChunk(
	tokenIDs []int,
	tk *tiktoken.Tiktoken,
) string {
	return tk.Decode(tokenIDs)
}

//////
// Exported functionalities.
//////

// SplitText divides the input text into chunks based on token count.
func (s TokenSplitter) SplitText(text string) ([]string, error) {
	// Initialize the tokenizer based on configuration.
	tk, err := s.initializeTokenizer()
	if err != nil {
		return nil, errorcatalog.MustGet(
			errorcatalog.ErrFailedToInitChunker,
			customerror.WithError(err),
		).NewFailedToError()
	}

	// Perform the text splitting using the tokenizer.
	return s.splitTextIntoChunks(text, tk), nil
}

//////
// Factory.
//////

// NewTokenSplitter creates a new TokenSplitter with default configuration.
// It allows overriding the default chunk size through the chunkThreshold
// parameter.
func NewTokenSplitter(chunkThreshold int) TokenSplitter {
	// Initialize with default values.
	ts := TokenSplitter{
		ChunkOverlap: DefaultTokenChunkOverlap,
		ChunkSize:    DefaultTokenChunkSize,
		EncodingName: DefaultTokenEncoding,
		ModelName:    DefaultTokenModelName,
	}

	// Override default chunk size if a threshold is provided.
	if chunkThreshold > 0 {
		ts.ChunkSize = chunkThreshold
	}

	return ts
}
