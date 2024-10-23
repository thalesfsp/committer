// root.go

package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/thalesfsp/committer/internal/errorcatalog"
	"github.com/thalesfsp/committer/internal/git"
	"github.com/thalesfsp/committer/internal/provider"
	"github.com/thalesfsp/committer/internal/shared"
	"github.com/thalesfsp/committer/internal/tui"
	"github.com/thalesfsp/customerror"
	"github.com/thalesfsp/inference/anthropic"
	"github.com/thalesfsp/inference/ollama"
	"github.com/thalesfsp/inference/openai"
	"github.com/thalesfsp/sypl"
	"github.com/thalesfsp/sypl/level"
	"github.com/thalesfsp/sypl/processor"
)

// Flags.
var (
	chunkThreshold    int
	llmAPICallTimeout time.Duration
	llmModel          string
	llmProvider       string
)

// CLI logger.
var cliLogger = sypl.NewDefault(shared.Name, level.Info, processor.Tagger(shared.Type))

// rootCmd represents the base command.
var rootCmd = &cobra.Command{
	Use:   shared.Name,
	Short: "A CLI tool to generate meaningful commit messages",
	Long: `Overview:
  Committer is a nimble and powerful CLI that streamline the process
  of generating meaningful commit messages. It leverages large language
  models (LLMs) to automatically create concise and descriptive commit
  messages based on the changes staged in a Git repository.

Providers:
  Each provider has their own requirements. OpenAI requires the
  OPENAI_API_KEY env var to be set while Claude (Anthropic)
  requires the ANTHROPIC_API_KEY env var. For the Ollama provider
  you can set its endpoint by setting the OLLAMA_ENDPOINT env var.`,
	Example: `  Use Anthropic provider with their most capable model.
  $ committer -p anthropic -m claude-3-5-sonnet-20240620`,
	Run: func(_ *cobra.Command, _ []string) {
		// For debug purposes.
		if shared.IsDebugMode() {
			cliLogger.Breakpoint(shared.Name)
		}

		// Check if the current directory is a Git repository.
		if !git.IsCurrentDirectoryGitRepo() {
			cliLogger.Fatalln(errorcatalog.MustGet(errorcatalog.ErrNotGitRepo).New())
		}

		providerInUse, err := provider.InitializeLLMProvider(
			llmProvider,
			llmModel,
		)
		if err != nil {
			cliLogger.Fatalln(err)
		}

		// Should exit if there are no changes and git isn't dirty.
		if !git.HasStagedChanges() && !git.IsDirty() {
			shared.NothingToDo()
		}

		// Check if there are staged changes.
		if !git.HasStagedChanges() {
			if tui.MustPromptYesNoTea("Would you like to add all changes?", false) {
				tui.SprinnerStart("Adding files...")

				if err := git.GitAddAll(); err != nil {
					tui.SprinnerStop()

					cliLogger.Fatalln(
						errorcatalog.MustGet(
							errorcatalog.ErrFailedToStageFiles,
							customerror.WithError(err),
						),
					)
				}

				tui.SprinnerStop()
			} else {
				shared.NothingToDo()
			}
		}

		// Get the diff and stats.
		tui.SprinnerStart("Getting diff...")

		diff, err := git.GetGitDiff()
		if err != nil {
			cliLogger.Fatalln(err)
		}

		tui.SprinnerStop()

		tui.SprinnerStart("Getting stats...")

		stats, err := git.GetGitStats()
		if err != nil {
			cliLogger.Fatalln(err)
		}

		tui.SprinnerStop()

		// Chunking if diff is too big.
		tui.SprinnerStart("Generating chunks...")

		chunks, err := provider.ChunkDiff(chunkThreshold, diff)
		if err != nil {
			cliLogger.Fatalln(err)
		}

		tui.SprinnerStop()

		commitMessage, err := provider.GenerateCommitMessageLoop(
			providerInUse,
			llmAPICallTimeout,
			stats, chunks)
		if err != nil {
			cliLogger.Fatalln(err)
		}

		if commitMessage == "" {
			cliLogger.Fatalln(errorcatalog.MustGet(errorcatalog.ErrEmptyCommitMessage).NewMissingError())
		}

		// Commit changes.
		tui.SprinnerStart("Committing changes...")

		if err := git.GitCommit(commitMessage); err != nil {
			cliLogger.Fatalln(err)
		}

		tui.SprinnerStop()

		// Push changes.
		tui.SprinnerStart("Pushing changes...")

		if tui.MustPromptYesNoTea("Would you like to push the commits?", true) {
			if err := git.GitPush(); err != nil {
				cliLogger.Fatalln(err)
			}
		}

		tui.SprinnerStop()

		// Tag changes.
		tui.SprinnerStart("Tagging changes...")

		if tui.MustPromptYesNoTea("Would you like to tag the commit?", false) {
			tag := tui.MustPromptForInputTea("Enter the tag name:")
			if err := git.GitTag(tag); err != nil {
				cliLogger.Fatalln(err)
			}

			if err := git.GitPushTags(); err != nil {
				cliLogger.Fatalln(err)
			}
		}

		tui.SprinnerStop()

		os.Exit(0)
	},
}

// Execute adds all child commands to the root command.
func Execute() error {
	return rootCmd.Execute()
}

//nolint:lll
func init() {
	// Add the flags to your command
	rootCmd.Flags().IntVarP(&chunkThreshold, "chunk-threshold", "c", 128000, "Chunk threshold in characters")
	rootCmd.Flags().DurationVarP(&llmAPICallTimeout, "llm-api-call-timeout", "t", 30*time.Second, "LLM API call timeout")
	rootCmd.Flags().StringVarP(&llmModel, "model", "m", "gpt-4o", "Model to be used by the provider for generating commit messages")

	llmProviderMsg := fmt.Sprintf(
		"LLM providers, allowed: %s",
		strings.Join([]string{
			openai.Name,
			anthropic.Name,
			ollama.Name,
		}, ", "),
	)

	rootCmd.Flags().StringVarP(&llmProvider, "provider", "p", openai.Name, llmProviderMsg)
}
