package tui

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thalesfsp/committer/internal/errorcatalog"
	"github.com/thalesfsp/committer/internal/shared"
	"github.com/thalesfsp/customerror"
)

//////
// Const, vars, types.
//////

// Define a lipgloss style for the spinner's foreground color.
var spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF875F"))

var (
	spinnerProgram *tea.Program // Represents the spinner's program instance
	spinnerMutex   sync.Mutex   // Ensures thread-safe operation for the spinner
)

// SpinnerModel contains the state and behavior of the spinner.
type SpinnerModel struct {
	spinner spinner.Model // Instance of the spinner model
	text    string        // Text to be displayed alongside the spinner
}

//////
// Exported methods.
//////

// Init initializes the spinner, returning its initial command for starting the tick loop.
func (m SpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages for spinner updates and potential quit requests.
func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle different types of messages received.
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Check for quit keystrokes and request to exit the spinner.
		switch msg.String() {
		case tea.KeyCtrlC.String(), tea.KeyEsc.String(), "q":
			// Exit program if user presses Ctrl+C, Esc, or 'q'.
			shared.NothingToDo()
		}
	case spinner.TickMsg:
		// Update the spinner based on tick messages.
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)

		return m, cmd
	}

	return m, nil
}

// View renders the spinner and associated text to be displayed.
func (m SpinnerModel) View() string {
	return ChoiceStyle.Render(fmt.Sprintf("%s %s\n", m.spinner.View(), m.text))
}

//////
// Exported functionalities.
//////

// SprinnerStart starts the spinner with the given text.
func SprinnerStart(text string) {
	// Functionality only proceeds if not in debug mode.
	if !shared.IsDebugMode() {
		return
	}

	spinnerMutex.Lock() // Ensure exclusive access to shared resources.
	defer spinnerMutex.Unlock()

	// Do nothing if a spinner is already running.
	if spinnerProgram != nil {
		return
	}

	// Initialize a new spinner with desired style and type.
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	// Create a new spinner model with the provided text.
	model := SpinnerModel{
		spinner: s,
		text:    text,
	}

	// Start a new Bubble Tea program for the spinner.
	spinnerProgram = tea.NewProgram(model)

	// Run the spinner program asynchronously.
	go func() {
		if _, err := spinnerProgram.Run(); err != nil {
			panic(
				errorcatalog.
					MustGet(errorcatalog.ErrFailedToRunTeaProgram).
					NewFailedToError(customerror.WithError(err)),
			)
		}
	}()
}

// SprinnerStop stops the running spinner.
func SprinnerStop() {
	spinnerMutex.Lock() // Ensure exclusive access to shared resources.
	defer spinnerMutex.Unlock()

	// If a spinner program is running, quit and clear the reference.
	if spinnerProgram != nil {
		spinnerProgram.Quit()

		spinnerProgram = nil
	}
}
