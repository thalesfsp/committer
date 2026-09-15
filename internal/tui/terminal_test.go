package tui

import (
	"bytes"
	"os"
	"testing"

	"github.com/thalesfsp/committer/internal/shared"
)

// TestIsInteractive_FalseUnderGoTest verifies terminal detection reports no
// terminal where there is none: go test always runs the test binary with
// stdout piped, so this holds locally and in CI.
func TestIsInteractive_FalseUnderGoTest(t *testing.T) {
	if isTerminal(os.Stdout) {
		t.Skip("stdout is a terminal, which go test normally never provides")
	}

	if IsInteractive() {
		t.Error("expected IsInteractive to be false when stdout is not a terminal")
	}
}

// TestIsTerminal_PipeIsNotATerminal verifies both ends of a pipe are
// reported as non-terminals.
func TestIsTerminal_PipeIsNotATerminal(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	defer reader.Close()
	defer writer.Close()

	if isTerminal(reader) || isTerminal(writer) {
		t.Error("expected a pipe not to be detected as a terminal")
	}
}

// TestSpinnerStart_NonInteractiveFallsBackToPlainText verifies that without a
// terminal the spinner prints the status text as a plain line and starts no
// Bubble Tea program, so SpinnerStop has nothing to do.
//
// This is a regression test for a panic in headless runs (git hooks, CI,
// redirected stdin or stdout): Bubble Tea could not open /dev/tty and took
// the whole process down from the spinner goroutine.
func TestSpinnerStart_NonInteractiveFallsBackToPlainText(t *testing.T) {
	if shared.IsDebugMode() {
		t.Skip("spinner is disabled in debug mode")
	}

	previousInteractive := interactive
	interactive = func() bool { return false }

	t.Cleanup(func() { interactive = previousInteractive })

	var output bytes.Buffer

	previousOutput := spinnerOutput
	spinnerOutput = &output

	t.Cleanup(func() { spinnerOutput = previousOutput })

	SpinnerStart("Getting diff...")

	if spinnerProgram != nil {
		t.Fatal("expected no Bubble Tea program to be started without a terminal")
	}

	SpinnerStop()

	SpinnerStart("Committing changes...")
	SpinnerStop()

	if got, want := output.String(), "Getting diff...\nCommitting changes...\n"; got != want {
		t.Errorf("expected plain status lines %q, got %q", want, got)
	}
}
