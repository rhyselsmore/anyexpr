package dispatch

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/rhyselsmore/anyexpr"
	rules "github.com/rhyselsmore/anyexpr/rules"
	"github.com/rhyselsmore/anyexpr/rules/action"
)

// --- Test types ---

type testEnv struct {
	Name   string
	Active bool
}

type testActions[E any] struct {
	Label  rules.Action[string, E]       `rule:"label,multi"`
	Move   rules.Action[string, E]       `rule:"move"`
	Delete rules.Action[action.NoArgs, E] `rule:"delete,terminal"`
}

func setup(t *testing.T) (*rules.Evaluation[testEnv, testActions[testEnv]], *Dispatcher[testEnv, testActions[testEnv]]) {
	t.Helper()

	actions, err := rules.DefineActions[testEnv, testActions[testEnv]]()
	if err != nil {
		t.Fatal(err)
	}
	compiler, err := anyexpr.NewCompiler[testEnv]()
	if err != nil {
		t.Fatal(err)
	}
	prog, err := rules.Compile(compiler, actions, []rules.Definition{
		{
			Name: "r1",
			Tags: []string{"billing"},
			When: `Active`,
			Then: []rules.ActionEntry{
				{Name: "label", Value: "billing"},
				{Name: "move", Value: "archive"},
			},
		},
		{
			Name: "r2",
			When: `has(Name, "delete")`,
			Then: []rules.ActionEntry{{Name: "delete"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	evaluator, err := rules.NewEvaluator(prog)
	if err != nil {
		t.Fatal(err)
	}
	eval, err := evaluator.Run(context.Background(), testEnv{Name: "test", Active: true})
	if err != nil {
		t.Fatal(err)
	}

	noop := func(ctx context.Context, eval *rules.Evaluation[testEnv, testActions[testEnv]]) error {
		return nil
	}
	d, err := New(
		Handle("handler-a", noop),
		Handle("handler-b", noop),
		Handle("handler-c", noop),
	)
	if err != nil {
		t.Fatal(err)
	}

	return eval, d
}

// --- New ---

func TestNew_Valid(t *testing.T) {
	t.Parallel()
	noop := func(ctx context.Context, eval *rules.Evaluation[testEnv, testActions[testEnv]]) error {
		return nil
	}
	_, err := New(Handle("a", noop))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNew_DuplicateHandler(t *testing.T) {
	t.Parallel()
	noop := func(ctx context.Context, eval *rules.Evaluation[testEnv, testActions[testEnv]]) error {
		return nil
	}
	_, err := New(Handle("a", noop), Handle("a", noop))
	if !errors.Is(err, ErrDuplicateHandler) {
		t.Errorf("got %v, want ErrDuplicateHandler", err)
	}
}

func TestNew_NoHandlers(t *testing.T) {
	t.Parallel()
	d, err := New[testEnv, testActions[testEnv]]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = d
}

// --- Plan ---

func TestPlan_Valid(t *testing.T) {
	t.Parallel()
	_, d := setup(t)
	_, err := d.Plan(
		Run[testEnv, testActions[testEnv]]("handler-a",
			When[testEnv, testActions[testEnv]](`Result.Label.Triggered`),
		),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPlan_UnknownHandler(t *testing.T) {
	t.Parallel()
	_, d := setup(t)
	_, err := d.Plan(
		Run[testEnv, testActions[testEnv]]("nope"),
	)
	if !errors.Is(err, ErrUnknownHandler) {
		t.Errorf("got %v, want ErrUnknownHandler", err)
	}
}

func TestPlan_BadExpression(t *testing.T) {
	t.Parallel()
	_, d := setup(t)
	_, err := d.Plan(
		Run[testEnv, testActions[testEnv]]("handler-a",
			When[testEnv, testActions[testEnv]](`invalid!!!`),
		),
	)
	if err == nil {
		t.Error("expected error for bad expression")
	}
}

// --- Execute ---

func TestExecute_HandlerCalled(t *testing.T) {
	t.Parallel()
	eval, d := setup(t)

	called := false
	d2, _ := New(Handle("test", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
		called = true
		return nil
	}))
	plan, _ := d2.Plan(Run[testEnv, testActions[testEnv]]("test"))
	_ = d

	result := plan.Execute(context.Background(), eval)
	if !called {
		t.Error("handler was not called")
	}
	if len(result.Dispatched) != 1 {
		t.Errorf("got %d dispatched, want 1", len(result.Dispatched))
	}
	if result.Dispatched[0].Handler != "test" {
		t.Errorf("handler = %q, want test", result.Dispatched[0].Handler)
	}
}

func TestExecute_WhenGates(t *testing.T) {
	t.Parallel()
	eval, d := setup(t)

	called := false
	d2, _ := New(Handle("test", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
		called = true
		return nil
	}))
	_ = d

	plan, _ := d2.Plan(
		Run[testEnv, testActions[testEnv]]("test",
			When[testEnv, testActions[testEnv]](`Result.Delete.Triggered`), // not triggered
		),
	)

	result := plan.Execute(context.Background(), eval)
	if called {
		t.Error("handler should not have been called")
	}
	if len(result.Dispatched) != 0 {
		t.Errorf("got %d dispatched, want 0", len(result.Dispatched))
	}
}

func TestExecute_WhenPasses(t *testing.T) {
	t.Parallel()
	eval, _ := setup(t)

	d, _ := New(Handle("test", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
		return nil
	}))

	plan, _ := d.Plan(
		Run[testEnv, testActions[testEnv]]("test",
			When[testEnv, testActions[testEnv]](`Result.Label.Triggered`),
		),
	)

	result := plan.Execute(context.Background(), eval)
	if len(result.Dispatched) != 1 {
		t.Errorf("got %d dispatched, want 1", len(result.Dispatched))
	}
	if result.Dispatched[0].MatchedExpr != "Result.Label.Triggered" {
		t.Errorf("matched expr = %q", result.Dispatched[0].MatchedExpr)
	}
}

func TestExecute_NoWhenAlwaysRuns(t *testing.T) {
	t.Parallel()
	eval, _ := setup(t)

	called := false
	d, _ := New(Handle("always", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
		called = true
		return nil
	}))

	plan, _ := d.Plan(Run[testEnv, testActions[testEnv]]("always"))
	plan.Execute(context.Background(), eval)

	if !called {
		t.Error("handler with no When should always run")
	}
}

// --- Gate ---

func TestExecute_GatePasses(t *testing.T) {
	t.Parallel()
	eval, _ := setup(t)

	d, _ := New(Handle("test", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
		return nil
	}))

	plan, _ := d.Plan(
		Gate[testEnv, testActions[testEnv]](`len(Matched) > 0`),
		Run[testEnv, testActions[testEnv]]("test"),
	)

	result := plan.Execute(context.Background(), eval)
	if !result.Gated {
		t.Error("expected gated")
	}
	if !result.GatePassed {
		t.Error("gate should pass")
	}
	if len(result.Dispatched) != 1 {
		t.Errorf("got %d dispatched, want 1", len(result.Dispatched))
	}
}

func TestExecute_GateBlocks(t *testing.T) {
	t.Parallel()
	eval, _ := setup(t)

	called := false
	d, _ := New(Handle("test", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
		called = true
		return nil
	}))

	plan, _ := d.Plan(
		Gate[testEnv, testActions[testEnv]](`Result.Delete.Triggered`),
		Run[testEnv, testActions[testEnv]]("test"),
	)

	result := plan.Execute(context.Background(), eval)
	if !result.Gated {
		t.Error("expected gated")
	}
	if result.GatePassed {
		t.Error("gate should block")
	}
	if called {
		t.Error("handler should not run when gate blocks")
	}
	if len(result.Dispatched) != 0 {
		t.Errorf("got %d dispatched, want 0", len(result.Dispatched))
	}
}

// --- Strategy ---

func TestExecute_AllContinue(t *testing.T) {
	t.Parallel()
	eval, _ := setup(t)

	d, _ := New(
		Handle("fail", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
			return fmt.Errorf("boom")
		}),
		Handle("after", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
			return nil
		}),
	)

	plan, _ := d.Plan(
		WithStrategy[testEnv, testActions[testEnv]](AllContinue),
		Run[testEnv, testActions[testEnv]]("fail"),
		Run[testEnv, testActions[testEnv]]("after"),
	)

	result := plan.Execute(context.Background(), eval)
	if len(result.Dispatched) != 2 {
		t.Errorf("AllContinue: got %d dispatched, want 2", len(result.Dispatched))
	}
	if result.OK() {
		t.Error("should have errors")
	}
}

func TestExecute_AllHaltOnError(t *testing.T) {
	t.Parallel()
	eval, _ := setup(t)

	d, _ := New(
		Handle("fail", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
			return fmt.Errorf("boom")
		}),
		Handle("after", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
			return nil
		}),
	)

	plan, _ := d.Plan(
		WithStrategy[testEnv, testActions[testEnv]](AllHaltOnError),
		Run[testEnv, testActions[testEnv]]("fail"),
		Run[testEnv, testActions[testEnv]]("after"),
	)

	result := plan.Execute(context.Background(), eval)
	if len(result.Dispatched) != 1 {
		t.Errorf("AllHaltOnError: got %d dispatched, want 1", len(result.Dispatched))
	}
}

func TestExecute_FirstMatch(t *testing.T) {
	t.Parallel()
	eval, _ := setup(t)

	count := 0
	d, _ := New(
		Handle("a", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
			count++
			return nil
		}),
		Handle("b", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
			count++
			return nil
		}),
	)

	plan, _ := d.Plan(
		WithStrategy[testEnv, testActions[testEnv]](FirstMatch),
		Run[testEnv, testActions[testEnv]]("a"),
		Run[testEnv, testActions[testEnv]]("b"),
	)

	result := plan.Execute(context.Background(), eval)
	if count != 1 {
		t.Errorf("FirstMatch: handler called %d times, want 1", count)
	}
	if len(result.Dispatched) != 1 {
		t.Errorf("got %d dispatched, want 1", len(result.Dispatched))
	}
}

// --- Panic recovery ---

func TestExecute_PanicRecovery(t *testing.T) {
	t.Parallel()
	eval, _ := setup(t)

	d, _ := New(Handle("panicker", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
		panic("oh no")
	}))

	plan, _ := d.Plan(Run[testEnv, testActions[testEnv]]("panicker"))

	result := plan.Execute(context.Background(), eval)
	if len(result.Dispatched) != 1 {
		t.Fatalf("got %d dispatched, want 1", len(result.Dispatched))
	}
	if !result.Dispatched[0].Panicked {
		t.Error("expected panicked")
	}
	if result.Dispatched[0].Error == nil {
		t.Error("expected error from panic")
	}
	if result.OK() {
		t.Error("should not be OK after panic")
	}
}

// --- Context cancellation ---

func TestExecute_ContextCancelled(t *testing.T) {
	t.Parallel()
	eval, _ := setup(t)

	d, _ := New(
		Handle("slow", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
			return nil
		}),
		Handle("after", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
			return nil
		}),
	)

	plan, _ := d.Plan(
		Run[testEnv, testActions[testEnv]]("slow"),
		Run[testEnv, testActions[testEnv]]("after"),
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := plan.Execute(ctx, eval)
	// With a cancelled context, dispatch should stop.
	if len(result.Dispatched) > 1 {
		t.Errorf("expected at most 1 dispatched after cancel, got %d", len(result.Dispatched))
	}
}

// --- Result ---

func TestResult_Evaluation(t *testing.T) {
	t.Parallel()
	eval, _ := setup(t)

	d, _ := New(Handle("test", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
		return nil
	}))
	plan, _ := d.Plan(Run[testEnv, testActions[testEnv]]("test"))

	result := plan.Execute(context.Background(), eval)
	if result.Evaluation != eval {
		t.Error("result should carry the original evaluation")
	}
}

func TestResult_Duration(t *testing.T) {
	t.Parallel()
	eval, _ := setup(t)

	d, _ := New(Handle("test", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
		time.Sleep(time.Millisecond)
		return nil
	}))
	plan, _ := d.Plan(Run[testEnv, testActions[testEnv]]("test"))

	result := plan.Execute(context.Background(), eval)
	if result.Duration < time.Millisecond {
		t.Errorf("duration = %v, expected >= 1ms", result.Duration)
	}
}

func TestResult_Debug(t *testing.T) {
	t.Parallel()
	eval, _ := setup(t)

	d, _ := New(Handle("test", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
		return nil
	}))
	plan, _ := d.Plan(Run[testEnv, testActions[testEnv]]("test"))

	result := plan.Execute(context.Background(), eval)
	debug := result.Debug()
	if debug == "" {
		t.Error("Debug() should return non-empty string")
	}
}

// --- Multiple When expressions ---

func TestExecute_MultipleWhens(t *testing.T) {
	t.Parallel()
	eval, _ := setup(t)

	called := false
	d, _ := New(Handle("test", func(ctx context.Context, e *rules.Evaluation[testEnv, testActions[testEnv]]) error {
		called = true
		return nil
	}))

	// First When fails, second passes — handler should run.
	plan, _ := d.Plan(
		Run[testEnv, testActions[testEnv]]("test",
			When[testEnv, testActions[testEnv]](`Result.Delete.Triggered`),       // false
			When[testEnv, testActions[testEnv]](`Result.Label.Triggered`), // true
		),
	)

	result := plan.Execute(context.Background(), eval)
	if !called {
		t.Error("handler should run when any When matches")
	}
	if result.Dispatched[0].MatchedExpr != "Result.Label.Triggered" {
		t.Errorf("matched = %q, want Result.Label.Triggered", result.Dispatched[0].MatchedExpr)
	}
}
