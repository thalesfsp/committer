package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thalesfsp/committer/internal/provider"
)

// providersCmd lists the supported LLM providers and how to configure them.
var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "Lists the supported LLM providers and how to configure them",
	Long: `Lists every provider accepted by --provider (-p): the environment variable
holding its API key, the environment variable overriding its endpoint, and
the model used when --model (-m) is not set.`,
	Run: func(cmd *cobra.Command, _ []string) {
		printProviders(cmd.OutOrStdout())
	},
}

// printProviders writes one block per provider, followed by the aliases.
func printProviders(w io.Writer) {
	for _, spec := range provider.Specs() {
		fmt.Fprintf(w, "%-18s %s\n", spec.Name, spec.Description)
		fmt.Fprintf(w, "  %-16s %s\n", "default model", orRequired(spec.DefaultModel, "--model (-m)"))
		fmt.Fprintf(w, "  %-16s %s\n", "api key", describeAPIKey(spec))
		fmt.Fprintf(w, "  %-16s %s\n\n", "base url", describeBaseURL(spec))
	}

	fmt.Fprintf(w, "Aliases: %s\n", strings.Join(aliasList(), ", "))
	fmt.Fprintln(w, "Defaults: COMMITTER_PROVIDER and COMMITTER_MODEL set --provider and --model.")
}

// describeAPIKey explains where the API key comes from.
func describeAPIKey(spec provider.Spec) string {
	envs := strings.Join(spec.APIKeyEnvs, " or ")

	if spec.APIKeyRequired {
		return envs + " (required)"
	}

	return envs + " (optional)"
}

// describeBaseURL explains where the endpoint comes from.
func describeBaseURL(spec provider.Spec) string {
	sources := make([]string, 0, 1+len(spec.BaseURLEnvs))
	sources = append(sources, "--base-url (-u)")
	sources = append(sources, spec.BaseURLEnvs...)

	description := strings.Join(sources, " or ")

	if spec.BaseURLRequired {
		return description + " (required)"
	}

	if spec.DefaultBaseURL != "" {
		return description + " (default " + spec.DefaultBaseURL + ")"
	}

	return description + " (optional)"
}

// orRequired returns the value, or a hint naming the flag that must set it.
func orRequired(value, flag string) string {
	if value != "" {
		return value
	}

	return "none, set " + flag
}

// aliasList renders the provider aliases as "alias=provider", sorted.
func aliasList() []string {
	entries := make([]string, 0, len(provider.Aliases()))

	for alias, name := range provider.Aliases() {
		entries = append(entries, alias+"="+name)
	}

	sort.Strings(entries)

	return entries
}

func init() {
	rootCmd.AddCommand(providersCmd)
}
