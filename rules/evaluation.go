package rules

import (
	"fmt"
	"strings"
	"time"
)

// Evaluation is the result of evaluating rules against an environment.
//
//   - E is the environment type (e.g. Email).
//   - A is the actions struct (e.g. EmailActions).
type Evaluation[E any, A any] struct {
	// Env is the environment value that was evaluated.
	Env E

	// Result holds the actions struct with triggered values populated.
	Result A

	// Matched lists the names of rules that matched, in evaluation order.
	Matched []string

	// Stopped is true if evaluation was halted by a stop or terminal.
	Stopped bool

	// StoppedBy is the name of the rule that halted evaluation.
	StoppedBy string

	// StartedAt is when the evaluation began.
	StartedAt time.Time

	// Duration is the total evaluation time.
	Duration time.Duration

	// Traced is true if tracing was enabled for this evaluation.
	Traced bool

	// Trace holds per-rule evaluation steps. Only populated when
	// tracing is enabled via WithTrace.
	Trace []Step
}

// Step records the evaluation of a single rule.
type Step struct {
	// Rule is the name of the rule.
	Rule string

	// Outcome is what happened — Matched, Skipped, Disabled, Excluded,
	// or SkipExpr.
	Outcome Outcome

	// Duration is the expression evaluation time for this rule.
	Duration time.Duration

	// Actions lists which action names this rule referenced.
	// Nil if the rule did not match.
	Actions []string

	// Mode is the evaluation mode used for this rule.
	Mode EvalMode

	// Selector is the expression that excluded the rule, if the
	// outcome was Excluded and an expression selector was active.
	Selector string

	// Skip is the skip expression that fired, if the outcome was
	// SkipExpr.
	Skip string
}

// Outcome describes what happened when a rule was evaluated.
type Outcome int

const (
	// OutcomeMatched means the rule's expression evaluated to true.
	OutcomeMatched Outcome = iota
	// OutcomeSkipped means the When expression evaluated to false.
	OutcomeSkipped
	// OutcomeDisabled means the rule's Enabled field was false.
	OutcomeDisabled
	// OutcomeExcluded means the rule was filtered by a selector.
	OutcomeExcluded
	// OutcomeSkipExpr means the rule's Skip expression evaluated to
	// true, causing the rule to be skipped.
	OutcomeSkipExpr
)

func (o Outcome) String() string {
	switch o {
	case OutcomeMatched:
		return "matched"
	case OutcomeSkipped:
		return "skipped"
	case OutcomeDisabled:
		return "disabled"
	case OutcomeExcluded:
		return "excluded"
	case OutcomeSkipExpr:
		return "skip-expr"
	default:
		return "unknown"
	}
}

// Debug returns a human-readable summary of the evaluation, suitable
// for logging or printing. Includes matched rules, timing, action
// results, and trace (if enabled).
func (e *Evaluation[E, A]) Debug() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Evaluation\n")
	fmt.Fprintf(&b, "  Started:   %s\n", e.StartedAt.Format("2006-01-02 15:04:05.000"))
	fmt.Fprintf(&b, "  Duration:  %s\n", e.Duration)
	fmt.Fprintf(&b, "  Matched:   %d rules %v\n", len(e.Matched), e.Matched)

	if e.Stopped {
		fmt.Fprintf(&b, "  Stopped:   yes (by %s)\n", e.StoppedBy)
	}

	fmt.Fprintf(&b, "  Env:       %+v\n", e.Env)
	fmt.Fprintf(&b, "  Result:    %+v\n", e.Result)

	if e.Traced && len(e.Trace) > 0 {
		fmt.Fprintf(&b, "  Trace:\n")
		for _, step := range e.Trace {
			fmt.Fprintf(&b, "    %-20s %-12s %s", step.Rule, step.Outcome, step.Duration)
			if step.Mode != WhenThenSkip {
				fmt.Fprintf(&b, "  mode=%s", step.Mode)
			}
			if len(step.Actions) > 0 {
				fmt.Fprintf(&b, "  actions=%v", step.Actions)
			}
			if step.Selector != "" {
				fmt.Fprintf(&b, "  selector=%q", step.Selector)
			}
			if step.Skip != "" {
				fmt.Fprintf(&b, "  skip=%q", step.Skip)
			}
			fmt.Fprintln(&b)
		}
	}

	return b.String()
}
