package rules

import (
	"context"
	"fmt"

	"github.com/rhyselsmore/anyexpr"
)

// Check validates a when-expression against the environment type
// without compiling a full program. Useful for iterating on
// expressions before committing them to rule definitions.
func Check[E any](compiler *anyexpr.Compiler[E], when string) error {
	_, err := compiler.Compile(anyexpr.NewSource("check", when))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCompile, err)
	}
	return nil
}

// TestRule compiles and evaluates a single rule definition against
// a single environment value. Returns the evaluation result for
// that rule only, with tracing enabled. Useful for testing rules
// in isolation.
func TestRule[E any, A any](
	compiler *anyexpr.Compiler[E],
	actions *Actions[E, A],
	def Definition,
	env E,
) (*Evaluation[E, A], error) {
	if actions.IsZero() {
		return nil, ErrActionsZero
	}

	prog, err := Compile(compiler, actions, []Definition{def})
	if err != nil {
		return nil, err
	}

	evaluator, err := NewEvaluator(prog)
	if err != nil {
		return nil, err
	}

	return evaluator.Run(context.Background(), env, WithTrace(true))
}

// Assertion is an expression evaluated against the actions struct A
// after rule evaluation. Written in the same expression language as
// rules, but targeting the result instead of the environment.
//
// Example assertions (given actions struct with Label, Priority):
//
//	"Label.Triggered"
//	"Label.Value == \"billing\""
//	"len(Label.Values) == 3"
//	"Priority.Value > 2"
//	"!Delete.Triggered"
type Assertion[A any] struct {
	prog *anyexpr.Program[A]
	expr string
}

// NewAssertion compiles an assertion expression against the actions
// struct type A. The expression has access to all exported fields on
// A and on each Action field (Triggered, Value, Values, Triggers).
func NewAssertion[A any](expr string) (*Assertion[A], error) {
	compiler, err := anyexpr.NewCompiler[A]()
	if err != nil {
		return nil, fmt.Errorf("%w: assertion compiler: %w", ErrCompile, err)
	}

	prog, err := compiler.Compile(anyexpr.NewSource("assertion", expr))
	if err != nil {
		return nil, fmt.Errorf("%w: assertion %q: %w", ErrAssert, expr, err)
	}

	return &Assertion[A]{prog: prog, expr: expr}, nil
}

// Assert evaluates the assertion against an evaluation result.
// Returns nil if the assertion passes (expression returns true),
// or ErrAssertFailed with context if it returns false.
func (a *Assertion[A]) Assert(eval *Evaluation[any, A]) error {
	matched, err := a.prog.Match(eval.Result)
	if err != nil {
		return fmt.Errorf("%w: assertion %q: %w", ErrAssert, a.expr, err)
	}
	if !matched {
		return fmt.Errorf("%w: %q", ErrAssertFailed, a.expr)
	}
	return nil
}

// AssertResult evaluates the assertion against an actions struct
// directly.
func (a *Assertion[A]) AssertResult(result A) error {
	matched, err := a.prog.Match(result)
	if err != nil {
		return fmt.Errorf("%w: assertion %q: %w", ErrAssert, a.expr, err)
	}
	if !matched {
		return fmt.Errorf("%w: %q", ErrAssertFailed, a.expr)
	}
	return nil
}

// TestCase bundles a rule definition, a test environment, and
// assertions to run against the result. Use with RunTestCase.
type TestCase[E any, A any] struct {
	// Name identifies this test case.
	Name string

	// Rule is the definition to test.
	Rule Definition

	// Env is the environment to evaluate against.
	Env E

	// Assertions are expressions evaluated against the actions struct
	// after evaluation. All must pass.
	Assertions []string
}

// TestResult is the outcome of running a TestCase.
type TestResult[E any, A any] struct {
	// Name is the test case name.
	Name string

	// Evaluation is the full evaluation result.
	Evaluation *Evaluation[E, A]

	// Passed is true if all assertions passed.
	Passed bool

	// Failures lists assertion expressions that failed.
	Failures []string

	// Error is set if compilation or evaluation failed before
	// assertions could run.
	Error error
}

// RunTestCase compiles and evaluates a single rule, then runs all
// assertions against the result. Returns a TestResult with pass/fail
// status and any failures.
func RunTestCase[E any, A any](
	compiler *anyexpr.Compiler[E],
	actions *Actions[E, A],
	tc TestCase[E, A],
) TestResult[E, A] {
	result := TestResult[E, A]{Name: tc.Name}

	eval, err := TestRule(compiler, actions, tc.Rule, tc.Env)
	if err != nil {
		result.Error = err
		return result
	}
	result.Evaluation = eval

	for _, expr := range tc.Assertions {
		assertion, err := NewAssertion[A](expr)
		if err != nil {
			result.Failures = append(result.Failures, fmt.Sprintf("%s (compile error: %v)", expr, err))
			continue
		}

		if err := assertion.AssertResult(eval.Result); err != nil {
			result.Failures = append(result.Failures, expr)
		}
	}

	result.Passed = len(result.Failures) == 0 && result.Error == nil
	return result
}
