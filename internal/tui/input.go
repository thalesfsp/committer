package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/thalesfsp/committer/internal/errorcatalog"
	"github.com/thalesfsp/committer/internal/shared"
	"github.com/thalesfsp/customerror"
)

//////
// Const, vars, types.
//////

// InputModel holds the state for text input prompts.
// It manages the text input field, any potential error state, the prompt message, and the final user input.
type InputModel struct {
	textinput textinput.Model // The model for the text input field.
	err       error           // Stores any error that may occur during input.
	prompt    string          // The prompt message to display to the user.
	input     string          // The actual input received from the user.
}

//////
// Exported methods.
//////

// Init initializes the input model for the Tea framework.
// Here, it sets up the text input blinking for user interaction.
func (m InputModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update serves as the state update function for the Tea framework.
// It processes incoming messages, which often stem from user input.
func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle different types of messages.
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// If the user presses Ctrl+C or Escape, exit the program.
		case tea.KeyCtrlC.String(), tea.KeyEsc.String():
			// Exit program if user presses Ctrl+C, or Esc.
			shared.NothingToDo()
		// If Enter is pressed, save the input and quit.
		case "enter":
			m.input = m.textinput.Value()

			return m, tea.Quit
		}
	}

	// Update the text input model with the new state.
	m.textinput, cmd = m.textinput.Update(msg)

	return m, cmd
}

// View returns the string that represents the current View of the program.
// It provides the prompt, the user input area, and a hint for submission.
func (m InputModel) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		QuestionStyle.Render(m.prompt),              // Render the prompt.
		InputStyle.Render(m.textinput.View()),       // Render the text input field.
		HintStyle.Render("(Press Enter to submit)"), // Display a hint for the user.
	)
}

//////
// Exported functionalities.
//////

// MustPromptForInputTea prompts the user for input using the BubbleTea framework.
// This function encapsulates the process of setting up, running, and managing the life cycle of the prompt.
func MustPromptForInputTea(prompt string) string {
	// Initialize the text input model with specific configurations.
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()                          // Set focus on the input field.
	ti.CharLimit = 256                  // Limit the input length.
	ti.Width = 50                       // Set a static width for the input.
	ti.Prompt = InputStyle.Render("> ") // Customize the input prompt appearance.

	// Create an instance of the InputModel with the given prompt.
	m := InputModel{
		textinput: ti,
		prompt:    prompt,
	}

	// Create a new Tea program with the input model.
	p := tea.NewProgram(m)

	// Run the program and handle any errors that may arise.
	model, err := p.Run()
	if err != nil {
		// Use a custom panic with error cataloging for initialization errors.
		panic(
			errorcatalog.MustGet(errorcatalog.ErrFailedToInitTea).
				NewFailedToError(customerror.WithError(err)),
		)
	}

	// Check if the final model is of type InputModel and if the input is non-empty.
	if m, ok := model.(InputModel); ok && m.input != "" {
		return m.input
	}

	// If input was empty or type assertion failed, return an empty string.
	return ""
}
