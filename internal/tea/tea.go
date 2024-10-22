package tea

import (
	"github.com/charmbracelet/bubbles/textinput"
	bt "github.com/charmbracelet/bubbletea"
	"github.com/thalesfsp/committer/internal/shared"
	"github.com/thalesfsp/customerror"
)

//////
// UI Interactions
//////

// PromptYesNoTea displays a yes/no question using Bubble Tea.
//
// Parameters:
//   - question: The question to display.
//   - defaultChoice: The default choice (true for Yes, false for No).
//
// Returns:
//   - bool: True for Yes, false for No.
//
// NOTE: In case of an error, the function panics.
func PromptYesNoTea(question string, defaultChoice bool) bool {
	choices := []string{"Yes", "No"}

	defaultIndex := 1

	if defaultChoice {
		defaultIndex = 0
	}

	m := choicesModel{
		question:      question,
		choices:       choices,
		defaultChoice: defaultIndex,
		cursor:        defaultIndex,
	}

	p := bt.NewProgram(m)

	model, err := p.Run()
	if err != nil {
		panic(shared.ErrorCatalog.MustGet(
			shared.ErrFailedToInitTea,
		).NewFailedToError(customerror.WithError(err)))
	}

	if m, ok := model.(choicesModel); ok && m.choice != "" {
		return m.choice == "Yes"
	}

	return false
}

// PromptForInputTea displays a text input prompt using Bubble Tea.
//
// Parameters:
//   - prompt: The prompt to display.
//
// Returns:
//   - string: The user's input.
//
// NOTE: In case of an error, the function panics.
func PromptForInputTea(prompt string) string {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.Prompt = InputStyle.Render("> ")

	m := inputModel{
		textinput: ti,
		prompt:    prompt,
	}

	p := bt.NewProgram(m)

	model, err := p.Run()
	if err != nil {
		panic(shared.ErrorCatalog.MustGet(
			shared.ErrFailedToInitTea,
		).NewFailedToError(customerror.WithError(err)))
	}

	if m, ok := model.(inputModel); ok && m.input != "" {
		return m.input
	}

	return ""
}

// PromptWithChoices displays a multi-choice prompt using Bubble Tea.
//
// Parameters:
//   - question: The question to display.
//   - choices: Available choices.
//
// Returns:
//   - string: The selected choice.
//
// NOTE: In case of an error, the function panics.
func PromptWithChoices(question string, choices []string) string {
	m := choicesModel{
		question: question,
		choices:  choices,
	}

	p := bt.NewProgram(m)

	model, err := p.Run()
	if err != nil {
		panic(shared.ErrorCatalog.MustGet(
			shared.ErrFailedToInitTea,
		).NewFailedToError(customerror.WithError(err)))
	}

	if m, ok := model.(choicesModel); ok && m.choice != "" {
		return m.choice
	}

	return ""
}
