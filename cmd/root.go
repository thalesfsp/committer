// root.go

package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/thalesfsp/customerror"
	"github.com/thalesfsp/inference/anthropic"
	"github.com/thalesfsp/inference/ollama"
	"github.com/thalesfsp/inference/openai"
	"github.com/thalesfsp/inference/provider"
	"github.com/thalesfsp/sypl"
	"github.com/thalesfsp/sypl/level"
	"github.com/thalesfsp/sypl/processor"
)

// Constants for the choices.
const (
	Accept        = "Accept"
	Edit          = "Edit"
	Exit          = "Exit"
	Regenerate    = "Regenerate"
	YesAndProceed = "Yes and proceed"
	NoAndProceed  = "No and proceed"
	AddAllFiles   = "Add all files"
	Proceed       = "Proceed"
)

// Name and Type of the CLI.
const (
	Name = "committer"
	Type = "cli"
)

// Flags.
var (
	chunkThreshold    int
	llmAPICallTimeout time.Duration
	llmProvider       string
	model             string
)

// CLI logger.
var cliLogger = sypl.NewDefault(Name, level.Info, processor.Tagger(Type))

// rootCmd represents the base command.
var rootCmd = &cobra.Command{
	Use:   Name,
	Short: "A CLI tool to generate meaningful commit messages",
	Run: func(cmd *cobra.Command, args []string) {
		// For debug purposes.
		if isDebugMode() {
			cliLogger.Breakpoint(Name)
		}

		// Check if the current directory is a Git repository.
		if !isCurrentDirectoryGitRepo() {
			cliLogger.Fatalln(ErrorCatalog.MustGet(ErrNotGitRepo).New())
		}

		// Provider definition.
		var providerInUse provider.IProvider

		switch llmProvider {
		case openai.Name:
			oai, err := openai.NewDefault()
			if err != nil {
				cliLogger.Fatalln(ErrorCatalog.MustGet(ErrFailedToSetupLLM).New())
			}

			providerInUse = oai
		case anthropic.Name:
			anth, err := anthropic.NewDefault()
			if err != nil {
				cliLogger.Fatalln(ErrorCatalog.MustGet(ErrFailedToSetupLLM).New())
			}

			providerInUse = anth
		case ollama.Name:
			oll, err := ollama.NewDefault()
			if err != nil {
				cliLogger.Fatalln(ErrorCatalog.MustGet(ErrFailedToSetupLLM).New())
			}

			providerInUse = oll
		default:
			cliLogger.Fatalln(ErrorCatalog.MustGet(ErrInvalidProvider).New())
		}

		if !isDirty() {
			fmt.Println("Nothing to do, exiting...")

			os.Exit(0)
		}

		// Check if there are staged changes.
		if !hasStagedChanges() {
			if promptYesNoTea("Would you like to add all changes?", false) {
				spinnerStart("Adding files...")
				if err := gitAddAll(); err != nil {
					spinnerStop()

					cliLogger.Fatalln(
						ErrorCatalog.MustGet(
							ErrFailedToStageFiles,
							customerror.WithError(err),
						),
					)
				}

				spinnerStop()
			} else {
				fmt.Println("Nothing to do, exiting...")

				os.Exit(0)
			}
		}

		// Get the diff and stats.
		spinnerStart("Getting diff...")
		diff, err := getGitDiff()
		if err != nil {
			cliLogger.Fatalln(err)
		}
		spinnerStop()

		spinnerStart("Getting stats...")
		stats, err := getGitStats()
		if err != nil {
			cliLogger.Fatalln(err)
		}
		spinnerStop()

		// Chunking if diff is too big.
		spinnerStart("Generating chunks...")
		chunks, err := chunkDiff(chunkThreshold, diff)
		if err != nil {
			cliLogger.Fatalln(err)
		}
		spinnerStop()

		// Generate commit message.
		commitMessage, err := generateCommitMessageLoop(providerInUse, stats, chunks)
		if err != nil {
			cliLogger.Fatalln(err)
		}

		if commitMessage == "" {
			cliLogger.Fatalln(ErrorCatalog.MustGet(ErrEmptyCommitMessage).NewMissingError())
		}

		// Commit changes.
		spinnerStart("Committing changes...")
		if err := gitCommit(commitMessage); err != nil {
			cliLogger.Fatalln(err)
		}
		spinnerStop()

		// Push changes.
		spinnerStart("Pushing changes...")
		if promptYesNoTea("Would you like to push the commits?", true) {
			if err := gitPush(); err != nil {
				cliLogger.Fatalln(err)
			}
		}
		spinnerStop()

		// Tag changes.
		spinnerStart("Tagging changes...")
		if promptYesNoTea("Would you like to tag the commit?", false) {
			tag := promptForInputTea("Enter the tag name:")
			if err := gitTag(tag); err != nil {
				cliLogger.Fatalln(err)
			}

			if err := gitPushTags(); err != nil {
				cliLogger.Fatalln(err)
			}
		}
		spinnerStop()

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
	rootCmd.Flags().DurationVarP(&llmAPICallTimeout, "llm-api-call-timeout", "t", 15*time.Second, "LLM API call timeout")
	rootCmd.Flags().StringVarP(&model, "model", "m", "gpt-4o", "Model to be used by the provider for generating commit messages")

	llmProviderMsg := fmt.Sprintf(
		"LLM providers, allowed: %s",
		strings.Join([]string{
			openai.Name,
			anthropic.Name,
			ollama.Name,
		}, ","),
	)

	rootCmd.Flags().StringVarP(&llmProvider, "provider", "p", openai.Name, llmProviderMsg)
}
