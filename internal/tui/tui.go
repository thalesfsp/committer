package tui

import (
	"github.com/charmbracelet/lipgloss"
)

//////
// Const, vars, types.
//////

// Style definitions.
var (
	ChoiceStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	CursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	HintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#767676"))
	InputStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	QuestionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF06B7")).Bold(true)
)
