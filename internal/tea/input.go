/*
Terminal UI.

This file implements a terminal-based user input interface using the Bubble Tea
framework. It provides a styled text input prompt with visual feedback and
keyboard interaction support.

NOTE:
This component uses the Bubble Tea TUI framework and Lip Gloss styling library
to create an interactive and visually appealing input interface.
*/

package tea

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	bt "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//////
// Style Constants
//////

// Color codes for UI elements.
const (
	ColorAccent = "#FF06B7"
	ColorCursor = "#00FF00"
	ColorHint   = "#767676"
	ColorInput  = "#00FF00"
	ColorText   = "#FFFFFF"
)

//////
// Styles
//////

var (
	// QuestionStyle defines the style for prompt questions.
	QuestionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorAccent)).
			Bold(true)

	// CursorStyle defines the style for the input cursor.
	CursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorCursor))

	// ChoiceStyle defines the style for selectable choices.
	ChoiceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText))

	// HintStyle defines the style for helper text and hints.
	HintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHint))

	// InputStyle defines the style for user input text.
	InputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorInput))
)

//////
// Types
//////

// inputModel represents the state and behavior of a text input prompt.
// It implements the bt.Model interface for use with Bubble Tea.
type inputModel struct {
	// err stores any error that occurred during input processing.
	err error

	// input stores the final submitted input value.
	input string

	// prompt stores the question or instruction shown to the user.
	prompt string

	// textinput handles the actual text input functionality.
	textinput textinput.Model
}

//////
// Tea Model Implementation
//////

// Init initializes the input model and starts the cursor blinking.
//
// Returns:
//   - bt.Cmd: Command to start the cursor blinking.
func (m inputModel) Init() bt.Cmd {
	return textinput.Blink
}

// Update handles input events and updates the model state.
//
// Parameters:
//   - msg: The message to process (usually a keyboard event).
//
// Returns:
//   - bt.Model: The updated model.
//   - bt.Cmd: The next command to execute, if any.
func (m inputModel) Update(msg bt.Msg) (bt.Model, bt.Cmd) {
	var cmd bt.Cmd

	// Handle different types of messages.
	switch msg := msg.(type) {
	case bt.KeyMsg:
		cmd = m.handleKeyPress(msg)
		if cmd != nil {
			return m, cmd
		}
	}

	// Update the text input component.
	m.textinput, cmd = m.textinput.Update(msg)

	return m, cmd
}

// View renders the current state of the input model.
//
// Returns:
//   - string: The rendered view as a string.
func (m inputModel) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		QuestionStyle.Render(m.prompt),
		InputStyle.Render(m.textinput.View()),
		HintStyle.Render("(Press Enter to submit)"),
	)
}

//////
// Helper Methods
//////

// handleKeyPress processes keyboard input events.
//
// Parameters:
//   - msg: The keyboard message to process.
//
// Returns:
//   - bt.Cmd: Command to execute based on the key press.
func (m *inputModel) handleKeyPress(msg bt.KeyMsg) bt.Cmd {
	switch msg.String() {
	case "ctrl+c", "esc":
		return bt.Quit
	case "enter":
		m.input = m.textinput.Value()

		return bt.Quit
	default:
		return nil
	}
}

//////
// Constructor
//////

// NewInputModel creates a new input model with the given prompt.
//
// Parameters:
//   - prompt: The prompt text to display.
//   - maxLength: Maximum allowed input length (use 0 for unlimited).
//   - defaultValue: Initial input value (can be empty).
//
// Returns:
//   - *inputModel: Initialized input model.
func NewInputModel(prompt string, maxLength int, defaultValue string) *inputModel {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()

	if maxLength > 0 {
		ti.CharLimit = maxLength
	}

	if defaultValue != "" {
		ti.SetValue(defaultValue)
	}

	return &inputModel{
		textinput: ti,
		prompt:    prompt,
	}
}

//////
// Utility Functions
//////

// RunInputPrompt creates and runs an input prompt with the given configuration.
//
// Parameters:
//   - prompt: The prompt text to display.
//   - maxLength: Maximum allowed input length (use 0 for unlimited).
//   - defaultValue: Initial input value (can be empty).
//
// Returns:
//   - string: The user's input.
//   - error: Any error that occurred during input.
func RunInputPrompt(
	prompt string,
	maxLength int,
	defaultValue string,
) (string, error) {
	model := NewInputModel(prompt, maxLength, defaultValue)
	p := bt.NewProgram(model)

	m, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run input prompt: %w", err)
	}

	if finalModel, ok := m.(inputModel); ok {
		return finalModel.input, nil
	}

	return "", fmt.Errorf("failed to get input value")
}
