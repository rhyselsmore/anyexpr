package rules

import (
	"context"
	"errors"
	"testing"
)

func evalSetup(t *testing.T, defs []Definition, opts ...EvaluatorOpt) *Evaluator[testEnv, testActions[testEnv]] {
	t.Helper()
	actions, compiler := compileSetup(t)
	prog, err := Compile(compiler, actions, defs)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	ev, err := NewEvaluator(prog, opts...)
	if err != nil {
		t.Fatalf("NewEvaluator: %v", err)
	}
	return ev
}

// --- Basic evaluation ---

func TestEvaluator_SingleMatch(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "greet",
			When: `has(Name, "alice")`,
			Then: []ActionEntry{{Name: "label", Value: "friend"}},
		},
	})

	eval, err := ev.Run(context.Background(), testEnv{Name: "alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eval.Matched) != 1 || eval.Matched[0] != "greet" {
		t.Errorf("Matched = %v, want [greet]", eval.Matched)
	}
	if !eval.Result.Label.Triggered {
		t.Error("Label should be triggered")
	}
	if eval.Result.Label.Value != "friend" {
		t.Errorf("Label.Value = %q, want friend", eval.Result.Label.Value)
	}
}

func TestEvaluator_NoMatch(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "greet",
			When: `has(Name, "alice")`,
			Then: []ActionEntry{{Name: "label", Value: "friend"}},
		},
	})

	eval, err := ev.Run(context.Background(), testEnv{Name: "bob"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eval.Matched) != 0 {
		t.Errorf("Matched = %v, want []", eval.Matched)
	}
	if eval.Result.Label.Triggered {
		t.Error("Label should not be triggered")
	}
}

func TestEvaluator_MultipleMatches(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "label", Value: "a"}},
		},
		{
			Name: "r2",
			When: `Active`,
			Then: []ActionEntry{
				{Name: "label", Value: "b"},
				{Name: "move", Value: "inbox"},
			},
		},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eval.Matched) != 2 {
		t.Fatalf("got %d matched, want 2", len(eval.Matched))
	}

	labels := eval.Result.Label.Values
	if len(labels) != 2 || labels[0] != "a" || labels[1] != "b" {
		t.Errorf("Label.Values = %v, want [a b]", labels)
	}

	if eval.Result.Move.Value != "inbox" {
		t.Errorf("Move.Value = %q, want inbox", eval.Result.Move.Value)
	}
}

// --- All value types ---

func TestEvaluator_AllTypes(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "all",
			When: `Active`,
			Then: []ActionEntry{
				{Name: "label", Value: "tag"},
				{Name: "move", Value: "archive"},
				{Name: "read", Value: true},
				{Name: "priority", Value: 5},
				{Name: "score", Value: 0.95},
			},
		},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if eval.Result.Move.Value != "archive" {
		t.Errorf("Move = %q", eval.Result.Move.Value)
	}
	if eval.Result.Read.Value != true {
		t.Error("Read = false")
	}
	if eval.Result.Priority.Value != 5 {
		t.Errorf("Priority = %d", eval.Result.Priority.Value)
	}
	if eval.Result.Score.Value != 0.95 {
		t.Errorf("Score = %f", eval.Result.Score.Value)
	}
}

func TestEvaluator_NoArgs(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "cleanup",
			When: `Active`,
			Then: []ActionEntry{{Name: "delete"}},
		},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !eval.Result.Delete.Triggered {
		t.Error("Delete should be triggered")
	}
}

// --- Multi dedup ---

func TestEvaluator_MultiDedup(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{
				{Name: "label", Value: "dup"},
				{Name: "label", Value: "unique"},
			},
		},
		{
			Name: "r2",
			When: `Active`,
			Then: []ActionEntry{{Name: "label", Value: "dup"}},
		},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	labels := eval.Result.Label.Values
	if len(labels) != 2 || labels[0] != "dup" || labels[1] != "unique" {
		t.Errorf("Label.Values = %v, want [dup unique]", labels)
	}
}

// --- Single last-wins ---

func TestEvaluator_SingleLastWins(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "move", Value: "first"}},
		},
		{
			Name: "r2",
			When: `Active`,
			Then: []ActionEntry{{Name: "move", Value: "second"}},
		},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if eval.Result.Move.Value != "second" {
		t.Errorf("Move = %q, want second", eval.Result.Move.Value)
	}
}

// --- Stop / terminal ---

func TestEvaluator_StopHalts(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "stopper",
			When: `Active`,
			Stop: true,
			Then: []ActionEntry{{Name: "label", Value: "a"}},
		},
		{
			Name: "after",
			When: `Active`,
			Then: []ActionEntry{{Name: "label", Value: "b"}},
		},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !eval.Stopped {
		t.Error("expected stopped")
	}
	if eval.StoppedBy != "stopper" {
		t.Errorf("StoppedBy = %q, want stopper", eval.StoppedBy)
	}
	if len(eval.Matched) != 1 {
		t.Errorf("got %d matched, want 1", len(eval.Matched))
	}
}

func TestEvaluator_TerminalImpliesStop(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "cleanup",
			When: `Active`,
			Then: []ActionEntry{
				{Name: "label", Value: "before"},
				{Name: "delete"},
			},
		},
		{
			Name: "after",
			When: `Active`,
			Then: []ActionEntry{{Name: "label", Value: "after"}},
		},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !eval.Stopped {
		t.Error("expected stopped")
	}
	if !eval.Result.Delete.Triggered {
		t.Error("Delete should be triggered")
	}
	labels := eval.Result.Label.Values
	if len(labels) != 1 || labels[0] != "before" {
		t.Errorf("Label.Values = %v, want [before]", labels)
	}
}

// --- Provenance ---

func TestEvaluator_Triggers(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "r1",
			Tags: []string{"billing"},
			When: `Active`,
			Then: []ActionEntry{{Name: "label", Value: "from-r1"}},
		},
		{
			Name: "r2",
			Tags: []string{"shipping"},
			When: `Active`,
			Then: []ActionEntry{{Name: "label", Value: "from-r2"}},
		},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	triggers := eval.Result.Label.Triggers
	if len(triggers) != 2 {
		t.Fatalf("got %d triggers, want 2", len(triggers))
	}
	if triggers[0].Rule != "r1" || triggers[0].Value != "from-r1" {
		t.Errorf("trigger[0] = %+v", triggers[0])
	}
	if triggers[1].Rule != "r2" || triggers[1].Value != "from-r2" {
		t.Errorf("trigger[1] = %+v", triggers[1])
	}
	if len(triggers[0].Tags) != 1 || triggers[0].Tags[0] != "billing" {
		t.Errorf("trigger[0].Tags = %v, want [billing]", triggers[0].Tags)
	}
}

// --- Disabled rules ---

func TestEvaluator_DisabledSkipped(t *testing.T) {
	t.Parallel()
	f := false
	ev := evalSetup(t, []Definition{
		{
			Name:    "disabled",
			When:    `Active`,
			Enabled: &f,
			Then:    []ActionEntry{{Name: "label", Value: "nope"}},
		},
		{
			Name: "enabled",
			When: `Active`,
			Then: []ActionEntry{{Name: "label", Value: "yes"}},
		},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eval.Matched) != 1 || eval.Matched[0] != "enabled" {
		t.Errorf("Matched = %v, want [enabled]", eval.Matched)
	}
}

// --- Selectors ---

func TestEvaluator_WithTags(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{Name: "r1", Tags: []string{"billing"}, When: `Active`, Then: []ActionEntry{{Name: "label", Value: "a"}}},
		{Name: "r2", Tags: []string{"shipping"}, When: `Active`, Then: []ActionEntry{{Name: "label", Value: "b"}}},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true}, WithTags("billing"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	labels := eval.Result.Label.Values
	if len(labels) != 1 || labels[0] != "a" {
		t.Errorf("Label.Values = %v, want [a]", labels)
	}
}

func TestEvaluator_ExcludeTags(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{Name: "r1", Tags: []string{"billing"}, When: `Active`, Then: []ActionEntry{{Name: "label", Value: "a"}}},
		{Name: "r2", Tags: []string{"shipping"}, When: `Active`, Then: []ActionEntry{{Name: "label", Value: "b"}}},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true}, ExcludeTags("billing"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	labels := eval.Result.Label.Values
	if len(labels) != 1 || labels[0] != "b" {
		t.Errorf("Label.Values = %v, want [b]", labels)
	}
}

func TestEvaluator_WithNames(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{Name: "r1", When: `Active`, Then: []ActionEntry{{Name: "label", Value: "a"}}},
		{Name: "r2", When: `Active`, Then: []ActionEntry{{Name: "label", Value: "b"}}},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true}, WithNames("r2"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	labels := eval.Result.Label.Values
	if len(labels) != 1 || labels[0] != "b" {
		t.Errorf("Label.Values = %v, want [b]", labels)
	}
}

func TestEvaluator_ExcludeNames(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{Name: "r1", When: `Active`, Then: []ActionEntry{{Name: "label", Value: "a"}}},
		{Name: "r2", When: `Active`, Then: []ActionEntry{{Name: "label", Value: "b"}}},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true}, ExcludeNames("r1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	labels := eval.Result.Label.Values
	if len(labels) != 1 || labels[0] != "b" {
		t.Errorf("Label.Values = %v, want [b]", labels)
	}
}

func TestEvaluator_OnEvaluation_Defaults(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{Name: "r1", Tags: []string{"billing"}, When: `Active`, Then: []ActionEntry{{Name: "label", Value: "a"}}},
		{Name: "r2", Tags: []string{"shipping"}, When: `Active`, Then: []ActionEntry{{Name: "label", Value: "b"}}},
	}, OnEvaluation(WithTags("billing")))

	eval, err := ev.Run(context.Background(), testEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	labels := eval.Result.Label.Values
	if len(labels) != 1 || labels[0] != "a" {
		t.Errorf("Label.Values = %v, want [a]", labels)
	}
}

// --- Tracing ---

func TestEvaluator_Trace(t *testing.T) {
	t.Parallel()
	f := false
	ev := evalSetup(t, []Definition{
		{Name: "r1", When: `Active`, Then: []ActionEntry{{Name: "label", Value: "a"}}},
		{Name: "disabled", When: `Active`, Enabled: &f},
		{Name: "r3", When: `!Active`},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true}, WithTrace(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !eval.Traced {
		t.Error("expected traced")
	}
	if len(eval.Trace) != 3 {
		t.Fatalf("got %d steps, want 3", len(eval.Trace))
	}

	if eval.Trace[0].Outcome != OutcomeMatched {
		t.Errorf("step 0 outcome = %v, want Matched", eval.Trace[0].Outcome)
	}
	if eval.Trace[1].Outcome != OutcomeDisabled {
		t.Errorf("step 1 outcome = %v, want Disabled", eval.Trace[1].Outcome)
	}
	if eval.Trace[2].Outcome != OutcomeSkipped {
		t.Errorf("step 2 outcome = %v, want Skipped", eval.Trace[2].Outcome)
	}
}

func TestEvaluator_TraceOff(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{Name: "r1", When: `Active`, Then: []ActionEntry{{Name: "label", Value: "a"}}},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eval.Traced {
		t.Error("should not be traced by default")
	}
	if eval.Trace != nil {
		t.Error("Trace should be nil when tracing is off")
	}
}

func TestEvaluator_TraceExcluded(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{Name: "r1", Tags: []string{"billing"}, When: `Active`},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true},
		WithTrace(true), ExcludeTags("billing"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eval.Trace) != 1 {
		t.Fatalf("got %d steps, want 1", len(eval.Trace))
	}
	if eval.Trace[0].Outcome != OutcomeExcluded {
		t.Errorf("outcome = %v, want Excluded", eval.Trace[0].Outcome)
	}
}

// --- Timing ---

func TestEvaluator_Timing(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{Name: "r1", When: `Active`, Then: []ActionEntry{{Name: "label", Value: "a"}}},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eval.StartedAt.IsZero() {
		t.Error("StartedAt should be set")
	}
	if eval.Duration == 0 {
		t.Error("Duration should be > 0")
	}
}

// --- Context ---

func TestEvaluator_ContextCancelled(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{Name: "r1", When: `Active`, Then: []ActionEntry{{Name: "label", Value: "a"}}},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ev.Run(ctx, testEnv{Active: true})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("got %v, want context.Canceled", err)
	}
}

// --- Concurrency ---

func TestEvaluator_Concurrent(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{
				{Name: "label", Value: "concurrent"},
				{Name: "read", Value: true},
			},
		},
	})

	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			eval, err := ev.Run(context.Background(), testEnv{Active: true})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if eval.Result.Label.Value != "concurrent" {
				t.Errorf("Label.Value = %q, want concurrent", eval.Result.Label.Value)
			}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

// --- NewEvaluator validation ---

func TestNewEvaluator_ProgramZero(t *testing.T) {
	t.Parallel()
	_, err := NewEvaluator[testEnv, testActions[testEnv]](nil)
	if !errors.Is(err, ErrProgramZero) {
		t.Errorf("got %v, want ErrProgramZero", err)
	}
}

// --- Unfired actions ---

func TestEvaluator_UnfiredActions(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "label", Value: "a"}},
		},
	})

	eval, err := ev.Run(context.Background(), testEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Label was triggered.
	if !eval.Result.Label.Triggered {
		t.Error("Label should be triggered")
	}

	// Everything else was not.
	if eval.Result.Move.Triggered {
		t.Error("Move should not be triggered")
	}
	if eval.Result.Read.Triggered {
		t.Error("Read should not be triggered")
	}
	if eval.Result.Priority.Triggered {
		t.Error("Priority should not be triggered")
	}
	if eval.Result.Score.Triggered {
		t.Error("Score should not be triggered")
	}
	if eval.Result.Delete.Triggered {
		t.Error("Delete should not be triggered")
	}
}

// --- Outcome.String ---

func TestOutcome_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		o    Outcome
		want string
	}{
		{OutcomeMatched, "matched"},
		{OutcomeSkipped, "skipped"},
		{OutcomeDisabled, "disabled"},
		{OutcomeExcluded, "excluded"},
		{Outcome(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.o.String(); got != tt.want {
			t.Errorf("Outcome(%d).String() = %q, want %q", tt.o, got, tt.want)
		}
	}
}

// --- Selector ---

func TestSelector_NoFilters(t *testing.T) {
	t.Parallel()
	s := selector{}
	if !s.includes("anything", nil) {
		t.Error("no filters should include everything")
	}
}

func TestSelector_ExcludeOverridesInclude(t *testing.T) {
	t.Parallel()
	s := selector{
		onlyTags:     map[string]bool{"billing": true},
		excludeNames: map[string]bool{"r1": true},
	}
	if s.includes("r1", []string{"billing"}) {
		t.Error("exclude should override include")
	}
}

// --- NoArgs nil handling ---

func TestEvaluator_NoArgsNilValue(t *testing.T) {
	t.Parallel()
	// Compile with nil Value for NoArgs action.
	actions, compiler := compileSetup(t)
	prog, err := Compile(compiler, actions, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "delete"}}, // nil value
		},
	})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	ev, _ := NewEvaluator(prog)
	eval, err := ev.Run(context.Background(), testEnv{Active: true})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !eval.Result.Delete.Triggered {
		t.Error("Delete should be triggered with nil value")
	}
}

func TestCompile_NilValueForNonNoArgs(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(compiler, actions, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "label"}}, // nil value for string action
		},
	})
	if !errors.Is(err, ErrActionValueType) {
		t.Errorf("got %v, want ErrActionValueType", err)
	}
}
