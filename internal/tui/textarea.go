package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/thalesfsp/committer/internal/shared"
)

//////
// Const, vars, types.
//////

// Define a type for handling errors as messages within the TUI.
type errMsg error

// Define a model for managing the state of the text area.
type TextAreModel struct {
	textarea textarea.Model
	err      error
	done     bool
}

//////
// Exported methods.
//////

// Init initializes the text area model for the Tea framework.
func (m TextAreModel) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages for updating the text area model.
func (m TextAreModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd // Collect asynchronous commands to be executed together later.

	var cmd tea.Cmd

	// Handle different types of messages to update model state or take actions.
	switch msg := msg.(type) {
	case tea.KeyMsg: // Handle keyboard messages.
		switch msg.Type {
		case tea.KeyCtrlC:
			// Exit program if user presses Ctrl+C.
			shared.NothingToDo()
		case tea.KeyEscape:
			m.done = true

			// Trigger program quit when Escape is pressed.
			return m, tea.Quit
		default:
			// Ensure the textarea has focus to capture user inputs.
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

	// Handle errors as part of the model processing.
	case errMsg:
		m.err = msg

		return m, nil
	}

	// Update the text area with the message and accumulate resulting commands.
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...) // Batch commands and return.
}

// View method renders the current state of the model into a viewable string format.
func (m TextAreModel) View() string {
	if m.done {
		return m.textarea.Value() + "\n\n" // Display the text area's content when done.
	}

	// Render the hint for quitting and display the text area's current state.
	return fmt.Sprintf(
		"%s\n\n%s",
		m.textarea.View(),
		HintStyle.Render(
			fmt.Sprintf(
				`(Press %s when you are done)`,
				strings.ToUpper(tea.KeyEsc.String()),
			),
		),
	) + "\n\n"
}

//////
// Helpers.
//////

// Initialize the text area model with specific properties like placeholder and focus.
func initializeTextAreModel() TextAreModel {
	ti := textarea.New()
	ti.Placeholder = "Start typing..."
	ti.SetWidth(80)
	ti.SetHeight(10)
	ti.Focus() // Set the component to be focused initially for immediate user input.

	return TextAreModel{
		textarea: ti,
		err:      nil,
		done:     false, // Track completion status for user input (used to signal readiness to quit).
	}
}

//////
// Exported functionalities.
//////

// CommitMessageTextArea is an entry function to interact with the text area model.
//
//nolint:forcetypeassert
func CommitMessageTextArea() (string, error) {
	p := tea.NewProgram(initializeTextAreModel())

	// Run the TUI program and capture the final model and potential errors.
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	// When the model signals it's done, retrieve and return the textarea content.
	if m.(TextAreModel).done {
		return m.(TextAreModel).textarea.Value(), nil
	}

	return "", nil
}
