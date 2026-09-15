package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/thalesfsp/committer/internal/errorcatalog"
	"github.com/thalesfsp/committer/internal/git"
	"github.com/thalesfsp/committer/internal/provider"
	"github.com/thalesfsp/committer/internal/shared"
	"github.com/thalesfsp/committer/internal/tui"
	"github.com/thalesfsp/customerror"
	"github.com/thalesfsp/sypl/v2"
	"github.com/thalesfsp/sypl/v2/level"
	"github.com/thalesfsp/sypl/v2/processor"
)

// CLI tool configuration flags.
var (
	// Auto-accept mode: add all, approve generated message, push, skip tag.
	autoAccept bool

	// Threshold for how large a diff chunk can be before splitting.
	chunkThreshold int

	// Timeout duration for LLM API calls.
	llmAPICallTimeout time.Duration

	// Override for the provider's API base URL.
	llmBaseURL string

	// Maximum number of output tokens per LLM call.
	llmMaxOutputTokens int64

	// The model to be used for LLM.
	llmModel string

	// The provider for the LLM service.
	llmProvider string
)

// Logger setup for the CLI with default settings.
var cliLogger = sypl.NewDefault(
	shared.Name,
	level.Info,
	processor.Tagger(shared.Type),
)

// rootCmd is the primary entry point for the CLI tool, representing the base
// command.
var rootCmd = &cobra.Command{
	Use:   shared.Name,
	Short: "A CLI tool to generate meaningful commit messages",
	Long: `Overview:
  Committer is a nimble and powerful CLI that streamline the process
  of generating meaningful commit messages. It leverages large language
  models (LLMs) to automatically create concise and descriptive commit
  messages based on the changes staged in a Git repository.

Providers:
  Committer talks to OpenAI, Anthropic (Claude), Google (Gemini),
  xAI (Grok), Groq, DeepSeek, Mistral, OpenRouter, Hugging Face,
  Ollama, Azure OpenAI, Amazon Bedrock and any OpenAI-compatible
  endpoint. Each provider reads its API key from its usual
  environment variable (OPENAI_API_KEY, ANTHROPIC_API_KEY,
  XAI_API_KEY, ...) and has a sensible default model. Run
  "committer providers" for the full list.

  The provider and model can also be set once through the
  COMMITTER_PROVIDER and COMMITTER_MODEL environment variables.`,
	Example: `  Use Anthropic (Claude) with the default model.
  $ committer -p anthropic

  Use xAI (Grok) with a specific model.
  $ committer -p xai -m grok-4.5

  Use a local Ollama model.
  $ committer -p ollama -m qwen3

  Use any OpenAI-compatible server, e.g. LM Studio.
  $ committer -p openai-compatible --base-url http://localhost:1234/v1 -m my-model`,
	Run: func(cmd *cobra.Command, _ []string) {
		// Check if debug mode is enabled and set a breakpoint if so.
		if shared.IsDebugMode() {
			cliLogger.Breakpoint(shared.Name)
		}

		// Exit if the current directory is not a Git repository.
		if !git.IsCurrentDirectoryGitRepo() {
			cliLogger.Fatalln(errorcatalog.MustGet(
				errorcatalog.ErrNotGitRepo).New())
		}

		// Initialize the LLM provider using configuration provided by the user.
		llm, err := provider.InitializeLLMProvider(cmd.Context(), provider.Config{
			Provider:        llmProvider,
			Model:           llmModel,
			BaseURL:         llmBaseURL,
			MaxOutputTokens: llmMaxOutputTokens,
		})
		if err != nil {
			cliLogger.Fatalln(err)
		}

		fmt.Println(tui.HintStyle.Render(fmt.Sprintf("Using %s (%s)", llm.Provider(), llm.Model())))

		// If there are no changes to be committed, exit the process.
		if !git.HasStagedChanges() && !git.IsDirty() {
			shared.NothingToDo()
		}

		// Stage changes: auto-add in auto-accept mode, otherwise prompt.
		if !git.HasStagedChanges() {
			if autoAccept || tui.MustPromptYesNoTea(
				"Would you like to add all changes?", false) {
				tui.SpinnerStart("Adding files...")

				if err := git.GitAddAll(); err != nil {
					tui.SpinnerStop()

					cliLogger.Fatalln(
						errorcatalog.MustGet(
							errorcatalog.ErrFailedToStageFiles,
							customerror.WithError(err),
						),
					)
				}

				tui.SpinnerStop()
			} else {
				shared.NothingToDo()
			}
		}

		// Retrieve and process the Git diff and stats.
		tui.SpinnerStart("Getting diff...")

		diff, err := git.GetGitDiff()
		if err != nil {
			cliLogger.Fatalln(err)
		}

		tui.SpinnerStop()

		tui.SpinnerStart("Getting stats...")

		stats, err := git.GetGitStats()
		if err != nil {
			cliLogger.Fatalln(err)
		}

		tui.SpinnerStop()

		// If needed, chunk the Git diff based on the defined threshold.
		tui.SpinnerStart("Generating chunks...")

		chunks, err := provider.ChunkDiff(chunkThreshold, diff)
		if err != nil {
			cliLogger.Fatalln(err)
		}

		tui.SpinnerStop()

		// Generate the commit message by communicating with the LLM.
		commitMessage, err := provider.GenerateCommitMessageLoop(
			llm,
			llmAPICallTimeout,
			stats, chunks,
			autoAccept)
		if err != nil {
			cliLogger.Fatalln(err)
		}

		// Handle the scenario of an empty commit message.
		if commitMessage == "" {
			cliLogger.Fatalln(errorcatalog.MustGet(
				errorcatalog.ErrEmptyCommitMessage).NewMissingError())
		}

		// Commit the changes using the generated commit message.
		tui.SpinnerStart("Committing changes...")

		if err := git.GitCommit(commitMessage); err != nil {
			cliLogger.Fatalln(err)
		}

		tui.SpinnerStop()

		// Push: auto-push in auto-accept mode, otherwise prompt.
		tui.SpinnerStart("Pushing changes...")

		if autoAccept || tui.MustPromptYesNoTea("Would you like to push the commits?", true) {
			if err := git.GitPush(); err != nil {
				cliLogger.Fatalln(err)
			}
		}

		tui.SpinnerStop()

		// Skip tagging in auto-accept mode. Otherwise, offer smart tagging.
		if !autoAccept {
			if tui.MustPromptYesNoTea("Would you like to tag the commit?", false) {
				handleTagging()
			}
		}

		// Gracefully exit the application.
		os.Exit(0)
	},
}

// bumpPatch takes a semver tag like "v1.2.3" and returns "v1.2.4".
// Returns empty string if the tag doesn't match a recognized semver pattern.
func bumpPatch(tag string) string {
	raw := tag
	prefix := ""

	if strings.HasPrefix(raw, "v") {
		prefix = "v"
		raw = raw[1:]
	}

	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return ""
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s%s.%s.%d", prefix, parts[0], parts[1], patch+1)
}

// handleTagging implements the smart tagging flow: fetches remote tags,
// displays the latest 3, suggests next patch version, and lets user
// accept or enter a custom tag.
func handleTagging() {
	tui.SpinnerStart("Fetching tags...")

	if err := git.GitFetchTags(); err != nil {
		tui.SpinnerStop()
		cliLogger.Warnln("Failed to fetch remote tags, proceeding with local tags")
	}

	tags, err := git.GitGetLatestTags(3)

	tui.SpinnerStop()

	if err != nil || len(tags) == 0 {
		// No existing tags — fall back to manual input.
		fmt.Println(tui.HintStyle.Render("No existing tags found."))

		tag := tui.MustPromptForInputTea("Enter the tag name:")
		if tag == "" {
			return
		}

		if err := git.GitTag(tag); err != nil {
			cliLogger.Fatalln(err)
		}

		if err := git.GitPushTags(); err != nil {
			cliLogger.Fatalln(err)
		}

		return
	}

	// Display latest tags.
	fmt.Printf("\n%s\n", tui.QuestionStyle.Render("Latest tags:"))

	for _, t := range tags {
		fmt.Printf("  %s\n", t)
	}

	fmt.Println()

	// Suggest next patch version based on the latest tag.
	suggested := bumpPatch(tags[0])

	choices := []string{}

	if suggested != "" {
		choices = append(choices, suggested+" (suggested)")
	}

	choices = append(choices, "Enter custom tag")

	choice := tui.MustPromptWithChoices("Which tag would you like to use?", choices)

	var tag string

	switch {
	case strings.HasSuffix(choice, "(suggested)"):
		tag = suggested
	case choice == "Enter custom tag":
		tag = tui.MustPromptForInputTea("Enter the tag name:")
	}

	if tag == "" {
		return
	}

	if err := git.GitTag(tag); err != nil {
		cliLogger.Fatalln(err)
	}

	if err := git.GitPushTags(); err != nil {
		cliLogger.Fatalln(err)
	}
}

// Execute is called by main to run the root command and setup the CLI.
func Execute() error {
	return rootCmd.Execute()
}

// envOr returns the value of the environment variable, or the fallback when
// it is unset or empty.
func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}

// init is used to initialize the command and attach flags to it.
func init() {
	// Configure flags for chunk threshold, API call timeout, model, and provider.
	rootCmd.Flags().BoolVarP(&autoAccept, "auto-accept", "a", false,
		"Automatically add all files, approve the generated commit message, and push (skip tagging)")
	rootCmd.Flags().IntVarP(&chunkThreshold, "chunk-threshold", "c", 128000,
		"Chunk threshold in characters")
	rootCmd.Flags().DurationVarP(&llmAPICallTimeout,
		"llm-api-call-timeout", "t", 60*time.Second, "LLM API call timeout, including retries")
	rootCmd.Flags().StringVarP(&llmModel, "model", "m",
		envOr("COMMITTER_MODEL", ""),
		"Model to be used by the provider, defaults to the provider's default model (env: COMMITTER_MODEL)")
	rootCmd.Flags().StringVarP(&llmBaseURL, "base-url", "u", "",
		"Override the provider's API base URL (env: e.g. OPENAI_BASE_URL, XAI_BASE_URL, OLLAMA_HOST)")
	rootCmd.Flags().Int64Var(&llmMaxOutputTokens, "max-tokens", provider.DefaultMaxOutputTokens,
		"Maximum number of output tokens per LLM call, reasoning tokens count towards it")

	// Construct the message detailing which providers are allowed.
	llmProviderMsg := fmt.Sprintf(
		"LLM provider, allowed: %s (env: COMMITTER_PROVIDER)",
		strings.Join(provider.Names(), ", "),
	)

	// Assign the provider flag, enabling selection of the desired LLM provider.
	rootCmd.Flags().StringVarP(&llmProvider, "provider", "p",
		envOr("COMMITTER_PROVIDER", provider.OpenAI), llmProviderMsg)
}
