package tui

import (
	"os"
	"runtime"

	"github.com/mattn/go-isatty"
)

//////
// Const, vars, types.
//////

// interactive reports whether the TUI can run. It is a variable so tests can
// force the non-interactive path.
var interactive = IsInteractive

//////
// Exported functionalities.
//////

// IsInteractive reports whether the process is attached to a terminal on both
// ends, which is what Bubble Tea needs: one to render to and one to read
// keystrokes from. Bubble Tea falls back to the controlling terminal when
// stdin is redirected, so that counts too. When it returns false the spinner
// degrades to plain status lines and the interactive prompts cannot run.
func IsInteractive() bool {
	return isTerminal(os.Stdout) && (isTerminal(os.Stdin) || canOpenInputTTY())
}

//////
// Internal functionalities.
//////

// isTerminal reports whether the file is a terminal, including the pseudo
// terminals used by Cygwin and MSYS on Windows.
func isTerminal(file *os.File) bool {
	fd := file.Fd()

	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

// canOpenInputTTY reports whether the controlling terminal can be opened for
// input, which is what Bubble Tea does when stdin is not a terminal.
func canOpenInputTTY() bool {
	name := "/dev/tty"

	if runtime.GOOS == "windows" {
		name = "CONIN$"
	}

	tty, err := os.Open(name)
	if err != nil {
		return false
	}

	_ = tty.Close()

	return true
}
