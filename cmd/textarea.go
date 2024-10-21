package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type errMsg error

type textAreModel struct {
	textarea textarea.Model
	err      error
	done     bool
}

func initializeTextAreModel() textAreModel {
	ti := textarea.New()
	ti.Placeholder = "Start typing..."
	ti.SetWidth(80)
	ti.SetHeight(10)
	ti.Focus()

	return textAreModel{
		textarea: ti,
		err:      nil,
		done:     false,
	}
}

func (m textAreModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m textAreModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			fmt.Println("Nothing to do, exiting...")

			os.Exit(0)
		case tea.KeyEscape:
			m.done = true

			return m, tea.Quit
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()

				cmds = append(cmds, cmd)
			}
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m textAreModel) View() string {
	if m.done {
		return m.textarea.Value()
	}

	return fmt.Sprintf(
		"%s\n\n%s",
		m.textarea.View(),
		hintStyle.Render(
			fmt.Sprintf(
				`(Press %s when you are done)`,
				strings.ToUpper(tea.KeyEsc.String()),
			),
		),
	) + "\n\n"
}
