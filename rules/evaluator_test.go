package rules

import (
	"context"
	"errors"
	"testing"

	"github.com/rhyselsmore/anyexpr"
)

type evalEnv struct {
	Name   string
	Amount float64
	Active bool
}

type evalActions[E any] struct {
	Label    Action[string, E]  `rule:"label,multi"`
	Category Action[string, E]  `rule:"category"`
	Read     Action[bool, E]    `rule:"read"`
	Priority Action[int, E]     `rule:"priority"`
	Score    Action[float64, E] `rule:"score"`
	Delete   Action[NoArgs, E]  `rule:"delete,terminal"`
}

func evalSetup(t *testing.T, defs []Definition, evalOpts ...EvaluatorOpt) *Evaluator[evalActions[evalEnv], evalEnv] {
	t.Helper()
	actions, err := DefineActions[evalActions[evalEnv], evalEnv]()
	if err != nil {
		t.Fatalf("DefineActions: %v", err)
	}
	compiler, err := anyexpr.NewCompiler[evalEnv]()
	if err != nil {
		t.Fatalf("NewCompiler: %v", err)
	}
	rs, err := Compile(actions, compiler, defs)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	ev, err := NewEvaluator(actions, rs, evalOpts...)
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

	result, err := ev.Run(context.Background(), evalEnv{Name: "alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Matched) != 1 {
		t.Fatalf("got %d matched, want 1", len(result.Matched))
	}
	if result.Matched[0].Name != "greet" {
		t.Errorf("got %q, want %q", result.Matched[0].Name, "greet")
	}

	vals := result.Actions.Label.Values()
	if len(vals) != 1 || vals[0] != "friend" {
		t.Errorf("Label.Values() = %v, want [friend]", vals)
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

	result, err := ev.Run(context.Background(), evalEnv{Name: "bob"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Matched) != 0 {
		t.Errorf("got %d matched, want 0", len(result.Matched))
	}
	if result.Actions.Label.Fired() {
		t.Error("label should not be fired")
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
				{Name: "category", Value: "active"},
			},
		},
	})

	result, err := ev.Run(context.Background(), evalEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Matched) != 2 {
		t.Fatalf("got %d matched, want 2", len(result.Matched))
	}

	labels := result.Actions.Label.Values()
	if len(labels) != 2 || labels[0] != "a" || labels[1] != "b" {
		t.Errorf("Label.Values() = %v, want [a b]", labels)
	}

	cat, ok := result.Actions.Category.Value()
	if !ok || cat != "active" {
		t.Errorf("Category.Value() = (%q, %v), want (active, true)", cat, ok)
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
				{Name: "label", Value: "tag1"},
				{Name: "category", Value: "main"},
				{Name: "read", Value: true},
				{Name: "priority", Value: 5},
				{Name: "score", Value: 0.95},
			},
		},
	})

	result, err := ev.Run(context.Background(), evalEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v, _ := result.Actions.Category.Value(); v != "main" {
		t.Errorf("Category = %q, want main", v)
	}
	if v, _ := result.Actions.Read.Value(); !v {
		t.Error("Read = false, want true")
	}
	if v, _ := result.Actions.Priority.Value(); v != 5 {
		t.Errorf("Priority = %d, want 5", v)
	}
	if v, _ := result.Actions.Score.Value(); v != 0.95 {
		t.Errorf("Score = %f, want 0.95", v)
	}
}

func TestEvaluator_NoArgs(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "cleanup",
			When: `Active`,
			Then: []ActionEntry{{Name: "delete", Value: NoArgs{}}},
		},
	})

	result, err := ev.Run(context.Background(), evalEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Actions.Delete.Fired() {
		t.Error("Delete should be fired")
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

	result, err := ev.Run(context.Background(), evalEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	labels := result.Actions.Label.Values()
	if len(labels) != 2 || labels[0] != "dup" || labels[1] != "unique" {
		t.Errorf("Label.Values() = %v, want [dup unique]", labels)
	}
}

// --- Single last-wins ---

func TestEvaluator_SingleLastWins(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "category", Value: "first"}},
		},
		{
			Name: "r2",
			When: `Active`,
			Then: []ActionEntry{{Name: "category", Value: "second"}},
		},
	})

	result, err := ev.Run(context.Background(), evalEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cat, _ := result.Actions.Category.Value()
	if cat != "second" {
		t.Errorf("Category = %q, want second", cat)
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

	result, err := ev.Run(context.Background(), evalEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Stopped {
		t.Error("expected stopped")
	}
	if result.StoppedBy != "stopper" {
		t.Errorf("StoppedBy = %q, want stopper", result.StoppedBy)
	}
	if len(result.Matched) != 1 {
		t.Errorf("got %d matched, want 1", len(result.Matched))
	}
}

func TestEvaluator_TerminalImpliesStop(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "cleanup",
			When: `Active`,
			Then: []ActionEntry{
				{Name: "label", Value: "before-delete"},
				{Name: "delete", Value: NoArgs{}},
			},
		},
		{
			Name: "after",
			When: `Active`,
			Then: []ActionEntry{{Name: "label", Value: "should-not-appear"}},
		},
	})

	result, err := ev.Run(context.Background(), evalEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Stopped {
		t.Error("expected stopped")
	}
	if !result.Actions.Delete.Fired() {
		t.Error("Delete should be fired")
	}
	labels := result.Actions.Label.Values()
	if len(labels) != 1 || labels[0] != "before-delete" {
		t.Errorf("Label.Values() = %v, want [before-delete]", labels)
	}
}

// --- Provenance ---

func TestEvaluator_Provenance_ByRule(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "label", Value: "from-r1"}},
		},
		{
			Name: "r2",
			When: `Active`,
			Then: []ActionEntry{{Name: "label", Value: "from-r2"}},
		},
	})

	result, err := ev.Run(context.Background(), evalEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r1 := result.Actions.Label.ByRule("r1")
	if len(r1) != 1 || r1[0] != "from-r1" {
		t.Errorf("ByRule(r1) = %v, want [from-r1]", r1)
	}

	rules := result.Actions.Label.Rules()
	if len(rules) != 2 || rules[0] != "r1" || rules[1] != "r2" {
		t.Errorf("Rules() = %v, want [r1 r2]", rules)
	}
}

func TestEvaluator_Provenance_ByTag(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{
			Name: "r1",
			Tags: []string{"billing"},
			When: `Active`,
			Then: []ActionEntry{{Name: "label", Value: "billed"}},
		},
		{
			Name: "r2",
			Tags: []string{"shipping"},
			When: `Active`,
			Then: []ActionEntry{{Name: "label", Value: "shipped"}},
		},
	})

	result, err := ev.Run(context.Background(), evalEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	billing := result.Actions.Label.ByTag("billing")
	if len(billing) != 1 || billing[0] != "billed" {
		t.Errorf("ByTag(billing) = %v, want [billed]", billing)
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

	result, err := ev.Run(context.Background(), evalEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Matched) != 1 {
		t.Fatalf("got %d matched, want 1", len(result.Matched))
	}
	labels := result.Actions.Label.Values()
	if len(labels) != 1 || labels[0] != "yes" {
		t.Errorf("Label.Values() = %v, want [yes]", labels)
	}
}

// --- Selectors ---

func TestEvaluator_WithTags(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{Name: "r1", Tags: []string{"billing"}, When: `Active`, Then: []ActionEntry{{Name: "label", Value: "billed"}}},
		{Name: "r2", Tags: []string{"shipping"}, When: `Active`, Then: []ActionEntry{{Name: "label", Value: "shipped"}}},
	})

	result, err := ev.Run(context.Background(), evalEnv{Active: true}, WithTags("billing"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	labels := result.Actions.Label.Values()
	if len(labels) != 1 || labels[0] != "billed" {
		t.Errorf("Label.Values() = %v, want [billed]", labels)
	}
}

func TestEvaluator_ExcludeTags(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{Name: "r1", Tags: []string{"billing"}, When: `Active`, Then: []ActionEntry{{Name: "label", Value: "billed"}}},
		{Name: "r2", Tags: []string{"shipping"}, When: `Active`, Then: []ActionEntry{{Name: "label", Value: "shipped"}}},
	})

	result, err := ev.Run(context.Background(), evalEnv{Active: true}, ExcludeTags("billing"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	labels := result.Actions.Label.Values()
	if len(labels) != 1 || labels[0] != "shipped" {
		t.Errorf("Label.Values() = %v, want [shipped]", labels)
	}
}

func TestEvaluator_WithNames(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{Name: "r1", When: `Active`, Then: []ActionEntry{{Name: "label", Value: "a"}}},
		{Name: "r2", When: `Active`, Then: []ActionEntry{{Name: "label", Value: "b"}}},
	})

	result, err := ev.Run(context.Background(), evalEnv{Active: true}, WithNames("r2"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	labels := result.Actions.Label.Values()
	if len(labels) != 1 || labels[0] != "b" {
		t.Errorf("Label.Values() = %v, want [b]", labels)
	}
}

func TestEvaluator_ExcludeNames(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{Name: "r1", When: `Active`, Then: []ActionEntry{{Name: "label", Value: "a"}}},
		{Name: "r2", When: `Active`, Then: []ActionEntry{{Name: "label", Value: "b"}}},
	})

	result, err := ev.Run(context.Background(), evalEnv{Active: true}, ExcludeNames("r1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	labels := result.Actions.Label.Values()
	if len(labels) != 1 || labels[0] != "b" {
		t.Errorf("Label.Values() = %v, want [b]", labels)
	}
}

// --- Evaluator-level defaults ---

func TestEvaluator_OnEvaluation_Defaults(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{Name: "r1", Tags: []string{"billing"}, When: `Active`, Then: []ActionEntry{{Name: "label", Value: "a"}}},
		{Name: "r2", Tags: []string{"shipping"}, When: `Active`, Then: []ActionEntry{{Name: "label", Value: "b"}}},
	}, OnEvaluation(WithTags("billing")))

	result, err := ev.Run(context.Background(), evalEnv{Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	labels := result.Actions.Label.Values()
	if len(labels) != 1 || labels[0] != "a" {
		t.Errorf("Label.Values() = %v, want [a]", labels)
	}
}

func TestEvaluator_PerCallAdditive(t *testing.T) {
	t.Parallel()
	ev := evalSetup(t, []Definition{
		{Name: "r1", Tags: []string{"billing"}, When: `Active`, Then: []ActionEntry{{Name: "label", Value: "a"}}},
		{Name: "r2", Tags: []string{"billing", "archived"}, When: `Active`, Then: []ActionEntry{{Name: "label", Value: "b"}}},
	}, OnEvaluation(WithTags("billing")))

	result, err := ev.Run(context.Background(), evalEnv{Active: true}, ExcludeTags("archived"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	labels := result.Actions.Label.Values()
	if len(labels) != 1 || labels[0] != "a" {
		t.Errorf("Label.Values() = %v, want [a]", labels)
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

	_, err := ev.Run(ctx, evalEnv{Active: true})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("got %v, want context.Canceled", err)
	}
}

// --- Concurrent safety ---

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
			result, err := ev.Run(context.Background(), evalEnv{Active: true})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			labels := result.Actions.Label.Values()
			if len(labels) != 1 || labels[0] != "concurrent" {
				t.Errorf("Label.Values() = %v, want [concurrent]", labels)
			}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

// --- NewEvaluator validation ---

func TestNewEvaluator_NotDefined(t *testing.T) {
	t.Parallel()
	actions := &Actions[evalActions[evalEnv], evalEnv]{}
	rs := &Ruleset[evalEnv]{}
	_, err := NewEvaluator(actions, rs)
	if !errors.Is(err, ErrNotDefined) {
		t.Errorf("got %v, want ErrNotDefined", err)
	}
}
