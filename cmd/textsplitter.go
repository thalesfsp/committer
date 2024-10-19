package cmd

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
)

const (
	defaultTokenModelName    = "gpt-3.5-turbo"
	defaultTokenEncoding     = "cl100k_base"
	defaultTokenChunkSize    = 512
	defaultTokenChunkOverlap = 100
)

// TokenSplitter is a text splitter that will split texts by tokens.
type TokenSplitter struct {
	ChunkSize         int
	ChunkOverlap      int
	ModelName         string
	EncodingName      string
	AllowedSpecial    []string
	DisallowedSpecial []string
}

// NewTokenSplitter creates a new TokenSplitter with default options.
func NewTokenSplitter() TokenSplitter {
	return TokenSplitter{
		ChunkSize:    defaultTokenChunkSize,
		ChunkOverlap: defaultTokenChunkOverlap,
		ModelName:    defaultTokenModelName,
		EncodingName: defaultTokenEncoding,
	}
}

// SplitText splits a text into multiple chunks.
func (s TokenSplitter) SplitText(text string) ([]string, error) {
	// Get the tokenizer
	var tk *tiktoken.Tiktoken

	var err error

	if s.EncodingName != "" {
		tk, err = tiktoken.GetEncoding(s.EncodingName)
	} else {
		tk, err = tiktoken.EncodingForModel(s.ModelName)
	}

	if err != nil {
		return nil, fmt.Errorf("tiktoken.GetEncoding: %w", err)
	}

	return s.splitText(text, tk), nil
}

func (s TokenSplitter) splitText(text string, tk *tiktoken.Tiktoken) []string {
	splits := make([]string, 0)

	inputIDs := tk.Encode(text, s.AllowedSpecial, s.DisallowedSpecial)

	startIdx := 0

	curIdx := len(inputIDs)

	if startIdx+s.ChunkSize < curIdx {
		curIdx = startIdx + s.ChunkSize
	}

	for startIdx < len(inputIDs) {
		chunkIDs := inputIDs[startIdx:curIdx]

		splits = append(splits, tk.Decode(chunkIDs))

		startIdx += s.ChunkSize - s.ChunkOverlap

		curIdx = startIdx + s.ChunkSize

		if curIdx > len(inputIDs) {
			curIdx = len(inputIDs)
		}
	}

	return splits
}
