package tea

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/bubbles/spinner"
	bt "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF875F"))

type spinnerModel struct {
	spinner spinner.Model
	text    string
}

func (m spinnerModel) Init() bt.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg bt.Msg) (bt.Model, bt.Cmd) {
	switch msg := msg.(type) {
	case bt.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, bt.Quit
		}
	case spinner.TickMsg:
		var cmd bt.Cmd

		m.spinner, cmd = m.spinner.Update(msg)

		return m, cmd
	}

	return m, nil
}

func (m spinnerModel) View() string {
	return ChoiceStyle.Render(fmt.Sprintf("%s %s\n", m.spinner.View(), m.text))
}

var (
	spinnerProgram *bt.Program
	spinnerMutex   sync.Mutex
)

// SpinnerStart starts the spinner with the given text.
func SpinnerStart(text string) {
	// if !isDebugMode() {
	// 	return
	// }

	spinnerMutex.Lock()
	defer spinnerMutex.Unlock()

	if spinnerProgram != nil {
		return // Spinner is already running
	}

	s := spinner.New()

	s.Spinner = spinner.Dot

	s.Style = spinnerStyle

	model := spinnerModel{
		spinner: s,
		text:    text,
	}

	spinnerProgram = bt.NewProgram(model)

	go func() {
		if _, err := spinnerProgram.Run(); err != nil {
			panic("Error running spinner:" + err.Error())
		}
	}()
}

// SpinnerStop stops the spinner.
func SpinnerStop() {
	spinnerMutex.Lock()
	defer spinnerMutex.Unlock()

	if spinnerProgram != nil {
		spinnerProgram.Quit()

		spinnerProgram = nil
	}
}
