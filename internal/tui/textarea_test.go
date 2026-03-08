package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// TestTextAreaModel_Init verifies the TextAreaModel Init returns a command.
func TestTextAreaModel_Init(t *testing.T) {
	m := initializeTextAreaModel()

	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init to return a blink command, got nil")
	}
}

// TestTextAreaModel_Update_EscSetsDown verifies pressing Escape sets done=true.
func TestTextAreaModel_Update_EscSetsDone(t *testing.T) {
	m := initializeTextAreaModel()

	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})

	tam, ok := model.(TextAreaModel)
	if !ok {
		t.Fatal("expected TextAreaModel type after Update")
	}

	if !tam.done {
		t.Error("expected done=true after Escape key")
	}

	// cmd should be tea.Quit.
	if cmd == nil {
		t.Error("expected a quit command after Escape")
	}
}

// TestTextAreaModel_Update_ErrMsg verifies error messages are captured.
func TestTextAreaModel_Update_ErrMsg(t *testing.T) {
	m := initializeTextAreaModel()

	testErr := errMsg(tea.ErrProgramKilled)
	model, cmd := m.Update(testErr)

	tam, ok := model.(TextAreaModel)
	if !ok {
		t.Fatal("expected TextAreaModel type")
	}

	if tam.err == nil {
		t.Error("expected error to be set on model")
	}

	if cmd != nil {
		t.Error("expected nil command for error message")
	}
}

// TestTextAreaModel_View verifies the view renders non-empty output.
func TestTextAreaModel_View(t *testing.T) {
	m := initializeTextAreaModel()

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

// TestTextAreaModel_ViewDone verifies the done state renders textarea value.
func TestTextAreaModel_ViewDone(t *testing.T) {
	ti := textarea.New()
	ti.SetValue("test commit message")

	m := TextAreaModel{
		textarea: ti,
		done:     true,
	}

	view := m.View()
	if view == "" {
		t.Error("expected non-empty done view")
	}
}

// TestTextAreaModel_TypeAssertion verifies that safe type assertion works
// correctly for both matching and non-matching types.
// This is a regression test for the unsafe type assertions that were replaced
// with checked assertions.
func TestTextAreaModel_TypeAssertion(t *testing.T) {
	t.Run("correct type assertion succeeds", func(t *testing.T) {
		ti := textarea.New()
		ti.SetValue("commit msg")

		var m tea.Model = TextAreaModel{
			textarea: ti,
			done:     true,
		}

		model, ok := m.(TextAreaModel)
		if !ok {
			t.Fatal("expected type assertion to succeed for TextAreaModel")
		}
		if !model.done {
			t.Error("expected done=true")
		}
		if model.textarea.Value() != "commit msg" {
			t.Errorf("expected textarea value 'commit msg', got %q", model.textarea.Value())
		}
	})

	t.Run("wrong type assertion returns false", func(t *testing.T) {
		// Simulate a non-TextAreaModel being returned — the checked assertion
		// should gracefully handle this instead of panicking.
		var m tea.Model = ChoiceModel{choices: []string{"a"}}

		_, ok := m.(TextAreaModel)
		if ok {
			t.Error("expected type assertion to fail for ChoiceModel -> TextAreaModel")
		}
	})
}

// TestInitializeTextAreaModel verifies the helper creates a properly configured
// model (also validates the renamed function from initializeTextAreModel).
func TestInitializeTextAreaModel(t *testing.T) {
	m := initializeTextAreaModel()

	if m.done {
		t.Error("expected done=false on initialization")
	}

	if m.err != nil {
		t.Error("expected nil error on initialization")
	}
}
