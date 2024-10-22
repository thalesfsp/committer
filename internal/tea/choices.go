package tea

import (
	"fmt"
	"os"
	"strings"

	bt "github.com/charmbracelet/bubbletea"
)

// choicesModel holds the state for multiple-choice prompts.
type choicesModel struct {
	cursor        int
	choice        string
	question      string
	choices       []string
	defaultChoice int
}

// Init initializes the model.
func (m choicesModel) Init() bt.Cmd {
	return nil
}

func (m choicesModel) Update(msg bt.Msg) (bt.Model, bt.Cmd) {
	switch msg := msg.(type) {
	case bt.KeyMsg:
		switch msg.String() {
		case bt.KeyCtrlC.String(), bt.KeyEsc.String(), "q":
			fmt.Println("Nothing to do, exiting...")

			os.Exit(0)
		case "enter":
			// Set the choice and exit.
			m.choice = m.choices[m.cursor]

			return m, bt.Quit

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

func (m choicesModel) View() string {
	var s strings.Builder

	s.WriteString(QuestionStyle.Render(m.question))

	s.WriteString("\n\n")

	for i := 0; i < len(m.choices); i++ {
		cursor := "  "

		if m.cursor == i {
			cursor = CursorStyle.Render("➤ ")
		}

		choice := m.choices[i]

		if i == m.defaultChoice {
			choice += " (default)"
		}

		s.WriteString(cursor)

		s.WriteString(ChoiceStyle.Render(choice))

		s.WriteString("\n")
	}

	s.WriteString("\n")

	s.WriteString(HintStyle.Render(
		fmt.Sprintf(
			`(Use ↑/↓ to navigate, Enter to select, or %s, %s or "q" to quit)`,
			strings.ToUpper(bt.KeyCtrlC.String()),
			strings.ToUpper(bt.KeyEsc.String()),
		),
	))

	s.WriteString("\n\n")

	return s.String()
}
