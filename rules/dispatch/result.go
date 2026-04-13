package dispatch

import (
	"fmt"
	"strings"
	"time"

	rules "github.com/rhyselsmore/anyexpr/rules2"
)

// Result is the outcome of a dispatch run.
type Result[E any, A any] struct {
	// Plan is the name of the plan that was executed.
	Plan string

	// Evaluation is the original evaluation that was dispatched.
	Evaluation *rules.Evaluation[E, A]

	// Dispatched lists every handler that was invoked, in execution
	// order, with timing and error information.
	Dispatched []Dispatched

	// Duration is the total dispatch time (all handlers).
	Duration time.Duration

	// Gated is true if a gate expression was evaluated.
	Gated bool

	// GateExpr is the gate expression string, if Gated is true.
	GateExpr string

	// GatePassed is true if the gate expression passed. If false,
	// no handlers were invoked.
	GatePassed bool
}

// Errors returns all non-nil errors from dispatched handlers.
func (r *Result[E, A]) Errors() []error {
	var errs []error
	for _, d := range r.Dispatched {
		if d.Error != nil {
			errs = append(errs, d.Error)
		}
	}
	return errs
}

// OK returns true if no handler errors occurred.
func (r *Result[E, A]) OK() bool {
	for _, d := range r.Dispatched {
		if d.Error != nil {
			return false
		}
	}
	return true
}

// Dispatched records a single handler invocation.
type Dispatched struct {
	// Handler is the name of the handler.
	Handler string

	// MatchedExpr is the When expression that triggered this handler.
	MatchedExpr string

	// Duration is how long the handler took to execute.
	Duration time.Duration

	// Error is non-nil if the handler returned an error or panicked.
	Error error

	// Panicked is true if the handler panicked (Error will contain
	// the recovered value).
	Panicked bool
}

// Debug returns a human-readable summary of the dispatch result.
func (r *Result[E, A]) Debug() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Dispatch")
	if r.Plan != "" {
		fmt.Fprintf(&b, " (%s)", r.Plan)
	}
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "  Duration:    %s\n", r.Duration)
	fmt.Fprintf(&b, "  Handlers:    %d invoked\n", len(r.Dispatched))

	if r.Gated {
		fmt.Fprintf(&b, "  Gate:        %q → %v\n", r.GateExpr, r.GatePassed)
	}

	if !r.OK() {
		errs := r.Errors()
		fmt.Fprintf(&b, "  Errors:      %d\n", len(errs))
	}

	for _, d := range r.Dispatched {
		status := "ok"
		if d.Panicked {
			status = "PANIC"
		} else if d.Error != nil {
			status = "ERROR"
		}
		fmt.Fprintf(&b, "    %-20s %-6s %s", d.Handler, status, d.Duration)
		if d.MatchedExpr != "" {
			fmt.Fprintf(&b, "  when=%q", d.MatchedExpr)
		}
		if d.Error != nil {
			fmt.Fprintf(&b, "  err=%v", d.Error)
		}
		fmt.Fprintln(&b)
	}

	return b.String()
}
