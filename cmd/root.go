/*
Committer CLI Tool

This file implements the root command for a CLI tool that generates meaningful
commit messages using various LLM providers (OpenAI, Anthropic, Ollama). It
handles Git operations, LLM interactions, and user prompts to facilitate the
commit process.

Flow:
graph TD
    A[Start] --> B[Check Git Repository]
    B --> C[Initialize LLM Provider]
    C --> D[Check for Changes]
    D --> E[Get Git Diff & Stats]
    E --> F[Chunk Large Diffs]
    F --> G[Generate Commit Message]
    G --> H[Commit Changes]
    H --> I[Push Changes Option]
    I --> J[Tag Changes Option]
    J --> K[End]
*/

package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/thalesfsp/committer/internal/shared"
	"github.com/thalesfsp/committer/internal/tea"
	"github.com/thalesfsp/customerror"
	"github.com/thalesfsp/inference/anthropic"
	"github.com/thalesfsp/inference/ollama"
	"github.com/thalesfsp/inference/openai"
	"github.com/thalesfsp/inference/provider"
	"github.com/thalesfsp/sypl"
	"github.com/thalesfsp/sypl/level"
	"github.com/thalesfsp/sypl/processor"
)

//////
// Consts, vars, and types.
//////

// Choice constants define the available user interaction options.
const (
	Accept        = "Accept"
	AddAllFiles   = "Add all files"
	Edit          = "Edit"
	Exit          = "Exit"
	NoAndProceed  = "No and proceed"
	Proceed       = "Proceed"
	Regenerate    = "Regenerate"
	YesAndProceed = "Yes and proceed"
)

// Flag variables store command-line configuration options.
var (
	chunkThreshold    int
	llmAPICallTimeout time.Duration
	llmModel          string
	llmProvider       string
)

// Logger instance for CLI operations.
var cliLogger = sypl.NewDefault(shared.Name, level.Info, processor.Tagger(shared.Type))

//////
// Command Definition
//////

// rootCmd represents the base command of the CLI application.
var rootCmd = &cobra.Command{
	Use:   shared.Name,
	Short: "A CLI tool to generate meaningful commit messages",
	Run:   runRootCommand,
}

//////
// Command Functions
//////

// runRootCommand implements the main logic for the root command.
func runRootCommand(cmd *cobra.Command, args []string) {
	// Enable debug mode if requested.
	if isDebugMode() {
		cliLogger.Breakpoint(shared.Name)
	}

	// Validate Git repository.
	validateGitRepository()

	// Initialize LLM provider.
	provider, err := initializeLLMProvider(llmModel)
	if err != nil {
		cliLogger.Fatalln(err)
	}

	// Check for changes and handle staging.
	handleGitChanges()

	// Process Git diff and generate commit message.
	commitMessage := processGitDiffAndGenerateMessage(provider)

	// Perform Git operations.
	performGitOperations(commitMessage)

	os.Exit(0)
}

//////
// Helper Functions
//////

// validateGitRepository ensures the current directory is a Git repository.
func validateGitRepository() {
	if !isCurrentDirectoryGitRepo() {
		cliLogger.Fatalln(shared.ErrorCatalog.MustGet(shared.ErrNotGitRepo).New())
	}
}

// initializeLLMProvider sets up the requested LLM provider.
func initializeLLMProvider(model string) (provider.IProvider, error) {
	var providerInUse provider.IProvider

	var err error

	switch llmProvider {
	case openai.Name:
		providerInUse, err = openai.NewDefault(provider.WithDefaulModel(model))
	case anthropic.Name:
		providerInUse, err = anthropic.NewDefault(provider.WithDefaulModel(model))
	case ollama.Name:
		providerInUse, err = ollama.NewDefault(provider.WithDefaulModel(model))
	default:
		return nil, shared.ErrorCatalog.MustGet(shared.ErrInvalidProvider).New()
	}

	if err != nil {
		return nil,
			shared.ErrorCatalog.MustGet(
				shared.ErrFailedToSetupLLM,
				customerror.WithError(err),
			)
	}

	return providerInUse, nil
}

// handleGitChanges manages Git staging operations.
func handleGitChanges() {
	// Exit if no changes detected.
	if !isDirty() {
		fmt.Println("Nothing to do, exiting...")

		os.Exit(0)
	}

	// Handle unstaged changes.
	if !hasStagedChanges() {
		if !tea.PromptYesNoTea("Would you like to add all changes?", false) {
			fmt.Println("Nothing to do, exiting...")

			os.Exit(0)
		}

		tea.SpinnerStart("Adding files...")
		defer tea.SpinnerStop()

		if err := gitAddAll(); err != nil {
			cliLogger.Fatalln(
				shared.ErrorCatalog.MustGet(
					shared.ErrFailedToStageFiles,
					customerror.WithError(err),
				),
			)
		}
	}
}

// processGitDiffAndGenerateMessage handles diff processing and message
// generation.
func processGitDiffAndGenerateMessage(p provider.IProvider) string {
	// Get diff and stats.
	tea.SpinnerStart("Getting diff...")

	diff, err := getGitDiff()
	if err != nil {
		cliLogger.Fatalln(err)
	}

	tea.SpinnerStop()

	tea.SpinnerStart("Getting stats...")

	stats, err := getGitStats()
	if err != nil {
		cliLogger.Fatalln(err)
	}

	tea.SpinnerStop()

	// Process chunks if needed.
	tea.SpinnerStart("Generating chunks...")

	chunks, err := chunkDiff(chunkThreshold, diff)
	if err != nil {
		cliLogger.Fatalln(err)
	}

	tea.SpinnerStop()

	// Generate commit message.
	commitMessage, err := generateCommitMessageLoop(p, stats, chunks)
	if err != nil {
		cliLogger.Fatalln(err)
	}

	if commitMessage == "" {
		cliLogger.Fatalln(
			shared.ErrorCatalog.MustGet(shared.ErrEmptyCommitMessage).NewMissingError(),
		)
	}

	return commitMessage
}

// performGitOperations handles commit, push, and tag operations.
func performGitOperations(commitMessage string) {
	// Commit changes.
	tea.SpinnerStart("Committing changes...")

	if err := gitCommit(commitMessage); err != nil {
		cliLogger.Fatalln(err)
	}

	tea.SpinnerStop()

	// Push changes if requested.
	tea.SpinnerStart("Pushing changes...")

	if tea.PromptYesNoTea("Would you like to push the commits?", true) {
		if err := gitPush(); err != nil {
			cliLogger.Fatalln(err)
		}
	}

	tea.SpinnerStop()

	// Tag changes if requested.
	tea.SpinnerStart("Tagging changes...")

	if tea.PromptYesNoTea("Would you like to tag the commit?", false) {
		tag := tea.PromptForInputTea("Enter the tag name:")
		if err := gitTag(tag); err != nil {
			cliLogger.Fatalln(err)
		}

		if err := gitPushTags(); err != nil {
			cliLogger.Fatalln(err)
		}
	}

	tea.SpinnerStop()
}

//////
// Exported Functions
//////

// Execute adds all child commands to the root command.
func Execute() error {
	return rootCmd.Execute()
}

// init initializes the command flags.
func init() {
	// Define allowed providers for help message.
	allowedProviders := []string{
		openai.Name,
		anthropic.Name,
		ollama.Name,
	}

	llmProviderMsg := fmt.Sprintf(
		"LLM providers, allowed: %s",
		strings.Join(allowedProviders, ", "),
	)

	// Add command flags.
	rootCmd.PersistentFlags().IntVarP(
		&chunkThreshold,
		"chunk-threshold",
		"c",
		128000,
		"Max text size. Above this threshold, the text will be chunked",
	)

	rootCmd.PersistentFlags().DurationVarP(
		&llmAPICallTimeout,
		"llm-api-call-timeout",
		"t",
		15*time.Second,
		"LLM API call timeout",
	)

	rootCmd.PersistentFlags().StringVarP(
		&llmModel,
		"model",
		"m",
		"gpt-4o",
		"Model to be used by the provider for generating commit messages",
	)

	rootCmd.PersistentFlags().StringVarP(
		&llmProvider,
		"provider",
		"p",
		openai.Name,
		llmProviderMsg,
	)
}
