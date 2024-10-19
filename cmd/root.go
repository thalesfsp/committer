package cmd

import (
	"context"
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

//////
// Const, vars, and types.
//////

const (
	// Name of entity.
	Name = "committer"

	// Type of entity.
	Type = "cli"
)

//////
// CLI flags content.
//////

// Flags.
var (
	chunkThreshold    int
	llmAPICallTimeout time.Duration
	llmProvider       string
	model             string
)

//////
// CLI logger.
//////

var cliLogger = sypl.NewDefault(Name, level.Info, processor.Tagger(Type))

//////
// Command definition.
//////

// rootCmd represents the base command.
var rootCmd = &cobra.Command{
	Use:   Name,
	Short: "A CLI tool to generate meaningful commit messages",
	Run: func(cmd *cobra.Command, args []string) {
		//////
		// For debug purposes.
		//////

		if isDebugMode() {
			cliLogger.Breakpoint(Name)
		}

		//////
		// Check if the current directory is a Git repository.
		//////

		if !isCurrentDirectoryGitRepo() {
			cliLogger.Fatalln(ErrorCatalog.MustGet(ErrNotGitRepo).New())
		}

		//////
		// Provider definition.
		//////

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

		//////
		// Check if there are staged changes.
		//////

		// TODO: If partially staged, should ask to stage all changes.
		if !hasStagedChanges() {
			cliLogger.Debugln("No staged changes detected.")

			if promptYesNo("Would you like to add all changes?") {
				if err := gitAddAll(); err != nil {
					cliLogger.Fatalln(
						ErrorCatalog.MustGet(
							ErrFailedToStageFiles,
							customerror.WithError(err),
						),
					)
				}
			} else {
				cliLogger.Debugln("Nothing to do, exiting...")
				os.Exit(0)
			}
		}

		//////
		// Get the diff and stats.
		//////

		diff, err := getGitDiff()
		if err != nil {
			cliLogger.Fatalln(err)
		}

		//////
		// Generate commit message.
		//////

		stats, err := getGitStats()
		if err != nil {
			cliLogger.Fatalln(err)
		}

		//////
		// Chunking if diff is too big.
		//////

		chunks, err := chunkDiff(chunkThreshold, diff)
		if err != nil {
			cliLogger.Fatalln(err)
		}

		//////
		// Generate commit message using LLM.
		//////

		commitMessage := ""

		totalChunks := len(chunks) // Get the total number of chunks

		ctxWithTimeout, cancel := context.WithTimeout(
			context.Background(),
			llmAPICallTimeout,
		)
		defer cancel()

		for i, chunk := range chunks {
			message, err := generateCommitMessage(
				ctxWithTimeout,
				providerInUse,
				stats, chunk,
				i+1, totalChunks,
			)
			if err != nil {
				cliLogger.Fatalln(err)
			}

			// TODO: Should not add a new line here if there was no message
			// before.
			fmt.Printf("\nGenerated Commit Message:\n%s\n", message)

			// TODO: Add default value.
			if promptYesNo("Do you approve this commit message?") {
				commitMessage = message

				break
			} else if promptYesNo("Would you like to try again?") {
				continue
			} else if promptYesNo("Would you like to write the commit message yourself?") {
				commitMessage = promptForInput("Enter your commit message:")

				break
			}

			os.Exit(0)
		}

		//////
		// Commit changes.
		//////

		if err := gitCommit(commitMessage); err != nil {
			cliLogger.Fatalln(err)
		}

		//////
		// Push changes if user approves.
		//////

		if promptYesNo("Would you like to push the commits?") {
			if err := gitPush(); err != nil {
				cliLogger.Fatalln(err)
			}
		}

		//////
		// Tag commit if user approves.
		//////

		if promptYesNo("Would you like to tag the commit?") {
			tag := promptForInput("Enter the tag name:")
			if err := gitTag(tag); err != nil {
				cliLogger.Fatalln(err)
			}
			if err := gitPushTags(); err != nil {
				cliLogger.Fatalln(err)
			}
		}

		os.Exit(0)
	},
}

// Execute adds all child commands to the root command.
func Execute() error {
	return rootCmd.Execute()
}

//nolint:lll,gomnd,mnd,gochecknoinits
func init() {
	// Add the flag to your command
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
