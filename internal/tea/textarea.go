package tea

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	bt "github.com/charmbracelet/bubbletea"
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

func (m textAreModel) Init() bt.Cmd {
	return textarea.Blink
}

func (m textAreModel) Update(msg bt.Msg) (bt.Model, bt.Cmd) {
	var cmds []bt.Cmd

	var cmd bt.Cmd

	switch msg := msg.(type) {
	case bt.KeyMsg:
		switch msg.Type {
		case bt.KeyCtrlC:
			fmt.Println("Nothing to do, exiting...")

			os.Exit(0)
		case bt.KeyEscape:
			m.done = true

			return m, bt.Quit
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

	return m, bt.Batch(cmds...)
}

func (m textAreModel) View() string {
	if m.done {
		return m.textarea.Value() + "\n\n"
	}

	return fmt.Sprintf(
		"%s\n\n%s",
		m.textarea.View(),
		HintStyle.Render(
			fmt.Sprintf(
				`(Press %s when you are done)`,
				strings.ToUpper(bt.KeyEsc.String()),
			),
		),
	) + "\n\n"
}

// NewMessageTextArea provides a text area for manual commit message entry.
//
// Returns:
//   - string: The entered commit message.
//   - error: Error if text area initialization fails.
//
//nolint:forcetypeassert
func NewMessageTextArea() (string, error) {
	p := bt.NewProgram(initializeTextAreModel())

	m, err := p.Run()
	if err != nil {
		return "", err
	}

	if m.(textAreModel).done {
		return m.(textAreModel).textarea.Value(), nil
	}

	return "", nil
}
