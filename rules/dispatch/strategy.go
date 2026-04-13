package dispatch

import "fmt"

// Strategy controls how the dispatcher executes matching handlers.
type Strategy int

const (
	// AllContinue runs all matching handlers. Errors are collected
	// and returned together. All handlers run regardless of failures.
	AllContinue Strategy = iota

	// AllHaltOnError runs matching handlers in registration order.
	// Stops on the first error.
	AllHaltOnError

	// FirstMatch runs only the first matching handler and stops.
	FirstMatch
)

func (s Strategy) String() string {
	switch s {
	case AllContinue:
		return "all-continue"
	case AllHaltOnError:
		return "all-halt-on-error"
	case FirstMatch:
		return "first-match"
	default:
		return fmt.Sprintf("Strategy(%d)", int(s))
	}
}
