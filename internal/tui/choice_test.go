package tui

import (
	"bytes"
	"io"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestChoiceModel_Update_NoDebugOutput verifies that the Update method does not
// produce any debug output to stdout for unhandled message types.
// This is a regression test for a removed fmt.Printf("HERE %+v", msg) call.
func TestChoiceModel_Update_NoDebugOutput(t *testing.T) {
	m := ChoiceModel{
		choices:  []string{"Option A", "Option B"},
		question: "Pick one",
	}

	// Capture stdout to verify no debug output is printed.
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Send a non-KeyMsg message (e.g., a custom message type) to trigger the
	// default case in the switch statement.
	type customMsg struct{}
	updatedModel, cmd := m.Update(customMsg{})

	// Also send a tea.WindowSizeMsg which is another non-KeyMsg type.
	updatedModel, _ = updatedModel.(ChoiceModel).Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Restore stdout and read captured output.
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()

	captured := buf.String()
	if captured != "" {
		t.Errorf("expected no stdout output for unhandled messages, but got: %q", captured)
	}

	// Verify the model is returned unchanged and no command is issued.
	if cmd != nil {
		t.Errorf("expected nil command, got %v", cmd)
	}

	choiceModel, ok := updatedModel.(ChoiceModel)
	if !ok {
		t.Fatal("expected ChoiceModel type")
	}

	if choiceModel.choice != "" {
		t.Errorf("expected empty choice, got %q", choiceModel.choice)
	}
}

// TestChoiceModel_Update_Navigation verifies cursor navigation works correctly.
func TestChoiceModel_Update_Navigation(t *testing.T) {
	m := ChoiceModel{
		choices:  []string{"A", "B", "C"},
		question: "Pick",
		cursor:   0,
	}

	// Move down.
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	cm := model.(ChoiceModel)
	if cm.cursor != 1 {
		t.Errorf("expected cursor=1 after 'j', got %d", cm.cursor)
	}

	// Move down again.
	model, _ = cm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	cm = model.(ChoiceModel)
	if cm.cursor != 2 {
		t.Errorf("expected cursor=2, got %d", cm.cursor)
	}

	// Wrap around at bottom.
	model, _ = cm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	cm = model.(ChoiceModel)
	if cm.cursor != 0 {
		t.Errorf("expected cursor=0 (wrap), got %d", cm.cursor)
	}

	// Move up wraps to bottom.
	model, _ = cm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	cm = model.(ChoiceModel)
	if cm.cursor != 2 {
		t.Errorf("expected cursor=2 (wrap up), got %d", cm.cursor)
	}
}
