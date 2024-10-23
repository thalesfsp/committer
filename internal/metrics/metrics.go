package metrics

import (
	"expvar"
)

// NewInt creates and initializes a new expvar.Int.
func NewInt(name string) *expvar.Int {
	counter := expvar.NewInt(name)

	// Initialize the counter to zero.
	counter.Set(0)

	return counter
}
