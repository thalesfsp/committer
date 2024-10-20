package cmd

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF875F"))

type spinnerModel struct {
	spinner spinner.Model
	text    string
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	return choiceStyle.Render(fmt.Sprintf("%s %s\n", m.spinner.View(), m.text))
}

var (
	spinnerProgram *tea.Program
	spinnerMutex   sync.Mutex
)

// spinnerStart starts the spinner with the given text
func spinnerStart(text string) {
	if !isDebugMode() {
		return
	}

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

	spinnerProgram = tea.NewProgram(model)

	go func() {
		if _, err := spinnerProgram.Run(); err != nil {
			fmt.Println("Error running spinner:", err)
		}
	}()
}

// spinnerStop stops the spinner
func spinnerStop() {
	spinnerMutex.Lock()
	defer spinnerMutex.Unlock()

	if spinnerProgram != nil {
		spinnerProgram.Quit()
		spinnerProgram = nil
	}
}
