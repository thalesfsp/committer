package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Style definitions.
var (
	questionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF06B7")).Bold(true)
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	choiceStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	hintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#767676"))
	inputStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
)

// cliModel holds the state for multiple-choice prompts.
type cliModel struct {
	cursor        int
	choice        string
	question      string
	choices       []string
	defaultChoice int
}

// Init initializes the model.
func (m cliModel) Init() tea.Cmd {
	return nil
}

func (m cliModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyCtrlC.String(), tea.KeyEsc.String(), "q":
			fmt.Println("Nothing to do, exiting...")

			os.Exit(0)
		case "enter":
			// Set the choice and exit.
			m.choice = m.choices[m.cursor]

			return m, tea.Quit

		case "down", "j":
			m.cursor++

			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}

		case "up", "k":
			m.cursor--

			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
		}
	default:
		fmt.Printf("HERE %+v", msg)
	}

	return m, nil
}

func (m cliModel) View() string {
	var s strings.Builder

	s.WriteString(questionStyle.Render(m.question))

	s.WriteString("\n\n")

	for i := 0; i < len(m.choices); i++ {
		cursor := "  "

		if m.cursor == i {
			cursor = cursorStyle.Render("➤ ")
		}

		choice := m.choices[i]

		if i == m.defaultChoice {
			choice += " (default)"
		}

		s.WriteString(cursor)

		s.WriteString(choiceStyle.Render(choice))

		s.WriteString("\n")
	}

	s.WriteString("\n")

	s.WriteString(hintStyle.Render(
		fmt.Sprintf(
			`(Use ↑/↓ to navigate, Enter to select, or %s, %s or "q" to quit)`,
			strings.ToUpper(tea.KeyCtrlC.String()),
			strings.ToUpper(tea.KeyEsc.String()),
		),
	))

	s.WriteString("\n\n")

	return s.String()
}

// inputModel holds the state for text input prompts.
type inputModel struct {
	textinput textinput.Model
	err       error
	prompt    string
	input     string
}

// Init initializes the input model.
func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			m.input = m.textinput.Value()

			return m, tea.Quit
		}
	}

	m.textinput, cmd = m.textinput.Update(msg)

	return m, cmd
}

func (m inputModel) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		questionStyle.Render(m.prompt),
		inputStyle.Render(m.textinput.View()),
		hintStyle.Render("(Press Enter to submit)"),
	)
}
