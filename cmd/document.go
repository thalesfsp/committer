/*
Documentation Generator Command

This file implements a Cobra command for processing files and directories to
generate documentation using LLM providers. It supports multiple file inputs,
directory processing, and chunked processing for large files.

Flow:
graph TD
    A[Start] --> B[Initialize LLM Provider]
    B --> C[Process Input Files]
    C --> D[Process Directories]
    D --> E[Chunk Content]
    E --> F[Generate Documentation]
    F --> G[Save Output]
    G --> H[End]

    subgraph "File Processing"
        C --> C1[Read Files]
        C1 --> C2[Collect Content]
    end

    subgraph "Documentation Generation"
        F --> F1[Process Chunks]
        F1 --> F2[Combine Results]
    end
*/

package cmd

import (
	"context"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thalesfsp/committer/internal/shared"
	"github.com/thalesfsp/committer/internal/tea"
	"github.com/thalesfsp/concurrentloop"
	"github.com/thalesfsp/customerror"
	"github.com/thalesfsp/inference/provider"
)

//////
// Variables
//////

var (
	// Directories holds the list of directories to process.
	Directories []string

	// filePaths holds the list of files to process.
	filePaths []string
)

//////
// Constants
//////

// Default output file name for generated documentation.
const defaultOutputFile = "documentation.md"

// Default file permissions for output file.
const defaultFilePerms = 0o644

//////
// Command Definition
//////

// documentCmd represents the document command for processing files and generating
// documentation.
var documentCmd = &cobra.Command{
	Use:   "document",
	Short: "Process files outputting documentation",
	Run:   runDocumentCommand,
}

//////
// Command Functions
//////

// runDocumentCommand implements the main logic for the document command.
func runDocumentCommand(_ *cobra.Command, _ []string) {
	// Enable debug mode if requested.
	if isDebugMode() {
		cliLogger.Breakpoint(shared.Name)
	}

	// Initialize LLM provider and process files.
	provider, err := initializeLLMProvider(llmModel)
	if err != nil {
		cliLogger.Fatalln(err)
	}

	if err := processAndGenerateDocumentation(provider); err != nil {
		cliLogger.Fatalln(err)
	}
}

// processAndGenerateDocumentation handles file processing and documentation
// generation.
//
// Parameters:
//   - providerInUse: The LLM provider for generating documentation.
func processAndGenerateDocumentation(providerInUse provider.IProvider) error {
	// Log processing targets.
	logProcessingTargets()

	// Process files and collect content.
	content, err := processFiles()
	if err != nil {
		return err
	}

	// Process content and generate documentation.
	documentation, err := generateContentDocumentation(providerInUse, content)
	if err != nil {
		return err
	}

	// Save the generated documentation.
	if err := saveDocumentation(documentation); err != nil {
		return err
	}

	return nil
}

// logProcessingTargets logs the files and directories to be processed.
func logProcessingTargets() {
	cliLogger.Infolnf("Processing files: %v", filePaths)

	cliLogger.Infolnf("Processing directories: %v", Directories)
}

// processFiles reads and collects content from all specified files.
//
// Returns:
//   - []string: Collected content from all files.
func processFiles() ([]string, error) {
	var contentOfAllFiles []string

	tea.SpinnerStart("Processing specified files...")
	defer tea.SpinnerStop()

	// Process files concurrently.
	if _, errs := concurrentloop.Map(
		context.Background(),
		filePaths,
		func(_ context.Context, filePath string) (bool, error) {
			content, err := processFile(filePath)
			if err != nil {
				return false, err
			}

			contentOfAllFiles = append(contentOfAllFiles, content)

			return true, nil
		},
	); len(errs) > 0 {
		return nil, errs
	}

	return contentOfAllFiles, nil
}

// processFile reads the content of a single file.
//
// Parameters:
//   - filePath: Path to the file to process.
//
// Returns:
//   - string: Content of the file.
//   - error: Error if reading fails.
func processFile(filePath string) (string, error) {
	file, err := OpenFile(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return ReadFile(file)
}

// generateContentDocumentation processes content and generates documentation.
//
// Parameters:
//   - providerInUse: The LLM provider for generating documentation.
//   - content: Content to process.
//
// Returns:
//   - []string: Generated documentation chunks.
func generateContentDocumentation(
	providerInUse provider.IProvider,
	content []string,
) ([]string, error) {
	tea.SpinnerStart("Chunking content...")

	chunks, err := chunkDiff(
		chunkThreshold,
		strings.Join(content, "\n"),
	)
	if err != nil {
		tea.SpinnerStop()

		return nil, err
	}

	tea.SpinnerStop()

	totalChunks := len(chunks)

	cliLogger.Tracelnf(
		"Threshold: %d Total chunks: %d",
		chunkThreshold,
		totalChunks,
	)

	return processChunks(providerInUse, chunks, totalChunks)
}

// processChunks generates documentation for each content chunk.
//
// Parameters:
//   - providerInUse: The LLM provider for generating documentation.
//   - chunks: Content chunks to process.
//   - totalChunks: Total number of chunks.
//
// Returns:
//   - []string: Generated documentation for all chunks.
func processChunks(
	providerInUse provider.IProvider,
	chunks []string,
	totalChunks int,
) ([]string, error) {
	finalDocumentation := make([]string, 0, totalChunks)

	tea.SpinnerStart("Generating documentation...")
	defer tea.SpinnerStop()

	for i, chunk := range chunks {
		content, err := generateDocumentation(
			context.Background(),
			providerInUse,
			chunk,
			i+1,
			totalChunks,
		)
		if err != nil {
			return nil, err
		}

		finalDocumentation = append(finalDocumentation, content)
	}

	return finalDocumentation, nil
}

// saveDocumentation writes the generated documentation to a file.
//
// Parameters:
//   - documentation: Documentation chunks to save.
func saveDocumentation(documentation []string) error {
	tea.SpinnerStart("Saving documentation...")
	defer tea.SpinnerStop()

	if err := os.WriteFile(
		defaultOutputFile,
		[]byte(strings.Join(documentation, "\n")),
		defaultFilePerms,
	); err != nil {
		return shared.ErrorCatalog.MustGet(
			shared.ErrFailedToSaveDocumentation,
			customerror.WithError(err),
		)
	}

	return nil
}

//////
// Initialization
//////

func init() {
	rootCmd.AddCommand(documentCmd)

	// Add command flags.
	documentCmd.Flags().StringSliceVarP(
		&filePaths,
		"file-paths",
		"f",
		nil,
		"File paths to be processed. Both relative and absolute paths are accepted",
	)

	documentCmd.Flags().StringSliceVarP(
		&Directories,
		"directories",
		"d",
		nil,
		"Directories paths to be processed",
	)
}
