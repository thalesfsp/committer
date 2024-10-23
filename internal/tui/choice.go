package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thalesfsp/committer/internal/errorcatalog"
	"github.com/thalesfsp/committer/internal/shared"
	"github.com/thalesfsp/customerror"
)

//////
// Const, vars, types.
//////

// ChoiceModel holds the state for multiple-choice prompts.
// It manages the current selection, the list of choices, and the specific
// question being asked.
type ChoiceModel struct {
	cursor        int      // Current position of the cursor for selection.
	choice        string   // The selected choice.
	question      string   // The question to be presented.
	choices       []string // List of possible choices.
	defaultChoice int      // Index for the default choice.
}

//////
// Exported methods.
//////

// Init initializes the model.
// Returns a command but does nothing initially.
func (m ChoiceModel) Init() tea.Cmd {
	return nil
}

// Update processes incoming messages and updates the model based on user input.
// Handles key presses to navigate and select from the choices.
func (m ChoiceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle key messages, like navigation and exiting.
		switch msg.String() {
		case tea.KeyCtrlC.String(), tea.KeyEsc.String(), "q":
			// Exit program if user presses Ctrl+C, Esc, or 'q'.
			shared.NothingToDo()
		case "enter":
			// Finalize choice and quit when Enter is pressed.
			m.choice = m.choices[m.cursor]

			return m, tea.Quit
		case "down", "j":
			// Move cursor down in the list of choices.
			m.cursor++
			// Wrap around if at the bottom.
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}
		case "up", "k":
			// Move cursor up in the list of choices.
			m.cursor--
			// Wrap around if at the top.
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
		}
	default:
		// Handle other types of messages if necessary.
		fmt.Printf("HERE %+v", msg)
	}

	// Return the updated model and no command.
	return m, nil
}

// View renders the UI components, showing the question and the list of choices.
func (m ChoiceModel) View() string {
	var s strings.Builder

	// Render the question.
	s.WriteString(QuestionStyle.Render(m.question))
	s.WriteString("\n\n")

	// Render each of the choices with an optional cursor and default indicator.
	for i := 0; i < len(m.choices); i++ {
		cursor := "  " // Default cursor is a space.

		// Highlight the current cursor position.
		if m.cursor == i {
			cursor = CursorStyle.Render("➤ ")
		}

		choice := m.choices[i]

		// Mark the default choice.
		if i == m.defaultChoice {
			choice += " (default)"
		}

		// Print the choice with styling.
		s.WriteString(cursor)
		s.WriteString(ChoiceStyle.Render(choice))
		s.WriteString("\n")
	}

	s.WriteString("\n")

	// Provide usage hints for navigation and actions.
	s.WriteString(HintStyle.Render(
		fmt.Sprintf(
			`(Use ↑/↓ to navigate, Enter to select, or %s, %s or "q" to quit)`,
			strings.ToUpper(tea.KeyCtrlC.String()),
			strings.ToUpper(tea.KeyEsc.String()),
		),
	))

	s.WriteString("\n\n")

	return s.String()
}

//////
// Exported functionalities.
//////

// MustPromptWithChoices prompts the user with multiple choices using Tea.
// Returns the selected choice as a string.
func MustPromptWithChoices(question string, choices []string) string {
	m := ChoiceModel{
		question: question,
		choices:  choices,
	}

	p := tea.NewProgram(m)

	// Runs the program and handles any initialization errors.
	model, err := p.Run()
	if err != nil {
		panic(errorcatalog.
			MustGet(errorcatalog.ErrFailedToInitTea).
			NewFailedToError(customerror.WithError(err)),
		)
	}

	// Extracts and returns the selected choice if available.
	if m, ok := model.(ChoiceModel); ok && m.choice != "" {
		return m.choice
	}

	return ""
}

// MustPromptYesNoTea prompts a yes/no question using Tea.
// Returns true for 'Yes' and false for 'No'.
func MustPromptYesNoTea(question string, defaultChoice bool) bool {
	choices := []string{"Yes", "No"}

	// Determine the default choice index based on the boolean input.
	defaultIndex := 1
	if defaultChoice {
		defaultIndex = 0
	}

	m := ChoiceModel{
		question:      question,
		choices:       choices,
		defaultChoice: defaultIndex,
		cursor:        defaultIndex, // Set initial cursor to default choice
	}

	p := tea.NewProgram(m)

	// Run the program and handle errors.
	model, err := p.Run()
	if err != nil {
		panic(errorcatalog.
			MustGet(errorcatalog.ErrFailedToInitTea).
			NewFailedToError(customerror.WithError(err)),
		)
	}

	// Return true if 'Yes' is chosen, otherwise false.
	if m, ok := model.(ChoiceModel); ok && m.choice != "" {
		return m.choice == "Yes"
	}

	return false
}
