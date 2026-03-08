package provider

import (
	"strings"
	"testing"
)

// TestCommitPrompt_GrammarFix verifies the commit prompt does not contain the
// old grammar error "No more change 240 characters" and instead contains the
// corrected text "No more than 240 characters".
func TestCommitPrompt_GrammarFix(t *testing.T) {
	if strings.Contains(commitPrompt, "No more change 240 characters") {
		t.Error("commit prompt still contains grammar error: 'No more change 240 characters'")
	}

	if !strings.Contains(commitPrompt, "No more than 240 characters") {
		t.Error("commit prompt missing corrected text: 'No more than 240 characters'")
	}
}

// TestCommitPrompt_HasRequiredSections verifies the prompt contains expected
// structural sections.
func TestCommitPrompt_HasRequiredSections(t *testing.T) {
	required := []string{
		"## Task",
		"## Commit Message Template",
		"### Template Fields",
		"### Examples",
		"### Best Practices",
		"Change Statistics:",
	}

	for _, section := range required {
		if !strings.Contains(commitPrompt, section) {
			t.Errorf("commit prompt missing required section: %q", section)
		}
	}
}
