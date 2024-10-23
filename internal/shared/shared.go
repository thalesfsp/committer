package shared

import (
	"fmt"
	"os"
)

//////
// Const, vars, types.
//////

var syplLevel = os.Getenv("SYPL_LEVEL")

//////
// Exported functionalities.
//////

// IsDebugMode checks if the CLI is running in debug mode.
func IsDebugMode() bool {
	return syplLevel == "debug"
}

// NothingToDo prints a message and exits the program.
func NothingToDo() {
	fmt.Println("Nothing to do, exiting...")

	os.Exit(0)
}
