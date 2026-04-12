package rules

import (
	"testing"
)

// action_test uses struct{} as E since the env type is irrelevant
// for unit-testing action accessors.

func stringAction(name string, c Cardinality) *Action[string, struct{}] {
	a := &Action[string, struct{}]{}
	a.configure(name, c, false, 0)
	return a
}

func boolAction(name string) *Action[bool, struct{}] {
	a := &Action[bool, struct{}]{}
	a.configure(name, Single, false, 0)
	return a
}

func intAction(name string) *Action[int, struct{}] {
	a := &Action[int, struct{}]{}
	a.configure(name, Single, false, 0)
	return a
}

func float64Action(name string) *Action[float64, struct{}] {
	a := &Action[float64, struct{}]{}
	a.configure(name, Single, false, 0)
	return a
}

func noargsAction(name string) *Action[NoArgs, struct{}] {
	a := &Action[NoArgs, struct{}]{}
	a.configure(name, Single, true, 0)
	return a
}

// --- String Multi ---

func TestAction_Values_Multi_Multiple(t *testing.T) {
	t.Parallel()
	a := stringAction("tag", Multi)
	a.entries = append(a.entries, entry[string]{value: "a", ruleName: "r1"})
	a.entries = append(a.entries, entry[string]{value: "b", ruleName: "r2"})
	a.entries = append(a.entries, entry[string]{value: "a", ruleName: "r3"}) // duplicate
	a.resolve()

	vals := a.Values()
	if len(vals) != 2 {
		t.Fatalf("got %d values, want 2", len(vals))
	}
	if vals[0] != "a" || vals[1] != "b" {
		t.Errorf("got %v, want [a b]", vals)
	}
}

func TestAction_Values_Multi_Empty(t *testing.T) {
	t.Parallel()
	a := stringAction("tag", Multi)
	vals := a.Values()
	if vals == nil {
		t.Fatal("expected non-nil slice")
	}
	if len(vals) != 0 {
		t.Errorf("got %d values, want 0", len(vals))
	}
}

// --- String Single ---

func TestAction_Value_Single_Set(t *testing.T) {
	t.Parallel()
	a := stringAction("cat", Single)
	a.entries = append(a.entries, entry[string]{value: "billing", ruleName: "r1"})
	a.resolve()

	v, ok := a.Value()
	if !ok {
		t.Fatal("expected ok")
	}
	if v != "billing" {
		t.Errorf("got %q, want %q", v, "billing")
	}
}

func TestAction_Value_Single_NotSet(t *testing.T) {
	t.Parallel()
	a := stringAction("cat", Single)
	_, ok := a.Value()
	if ok {
		t.Error("expected not ok")
	}
}

func TestAction_Value_Single_LastWins(t *testing.T) {
	t.Parallel()
	a := stringAction("cat", Single)
	a.entries = append(a.entries, entry[string]{value: "first", ruleName: "r1"})
	a.entries = append(a.entries, entry[string]{value: "second", ruleName: "r2"})
	a.resolve()

	v, ok := a.Value()
	if !ok {
		t.Fatal("expected ok")
	}
	if v != "second" {
		t.Errorf("got %q, want %q", v, "second")
	}
}

// --- Bool ---

func TestAction_Value_Bool_True(t *testing.T) {
	t.Parallel()
	a := boolAction("read")
	a.entries = append(a.entries, entry[bool]{value: true, ruleName: "r1"})
	a.resolve()

	v, ok := a.Value()
	if !ok {
		t.Fatal("expected ok")
	}
	if !v {
		t.Error("expected true")
	}
}

func TestAction_Value_Bool_False(t *testing.T) {
	t.Parallel()
	a := boolAction("read")
	a.entries = append(a.entries, entry[bool]{value: false, ruleName: "r1"})
	a.resolve()

	v, ok := a.Value()
	if !ok {
		t.Fatal("expected ok")
	}
	if v {
		t.Error("expected false")
	}
}

func TestAction_Value_Bool_LastWins(t *testing.T) {
	t.Parallel()
	a := boolAction("read")
	a.entries = append(a.entries, entry[bool]{value: true, ruleName: "r1"})
	a.entries = append(a.entries, entry[bool]{value: false, ruleName: "r2"})
	a.resolve()

	v, _ := a.Value()
	if v {
		t.Error("expected false (last wins)")
	}
}

func TestAction_Value_Bool_NotSet(t *testing.T) {
	t.Parallel()
	a := boolAction("read")
	_, ok := a.Value()
	if ok {
		t.Error("expected not ok")
	}
}

// --- Int ---

func TestAction_Value_Int(t *testing.T) {
	t.Parallel()
	a := intAction("priority")
	a.entries = append(a.entries, entry[int]{value: 42, ruleName: "r1"})
	a.resolve()

	v, ok := a.Value()
	if !ok {
		t.Fatal("expected ok")
	}
	if v != 42 {
		t.Errorf("got %d, want 42", v)
	}
}

// --- Float64 ---

func TestAction_Value_Float64(t *testing.T) {
	t.Parallel()
	a := float64Action("score")
	a.entries = append(a.entries, entry[float64]{value: 0.95, ruleName: "r1"})
	a.resolve()

	v, ok := a.Value()
	if !ok {
		t.Fatal("expected ok")
	}
	if v != 0.95 {
		t.Errorf("got %f, want 0.95", v)
	}
}

// --- NoArgs ---

func TestAction_Fired_NoArgs_Triggered(t *testing.T) {
	t.Parallel()
	a := noargsAction("delete")
	a.entries = append(a.entries, entry[NoArgs]{value: NoArgs{}, ruleName: "r1"})
	a.resolve()

	if !a.Fired() {
		t.Error("expected fired")
	}
}

func TestAction_Fired_NoArgs_NotTriggered(t *testing.T) {
	t.Parallel()
	a := noargsAction("delete")
	if a.Fired() {
		t.Error("expected not fired")
	}
}

// --- Fired (all types) ---

func TestAction_Fired_String_Present(t *testing.T) {
	t.Parallel()
	a := stringAction("label", Multi)
	a.entries = append(a.entries, entry[string]{value: "x", ruleName: "r1"})
	if !a.Fired() {
		t.Error("expected fired")
	}
}

func TestAction_Fired_String_Absent(t *testing.T) {
	t.Parallel()
	a := stringAction("label", Multi)
	if a.Fired() {
		t.Error("expected not fired")
	}
}

// --- ByRule ---

func TestAction_ByRule_Match(t *testing.T) {
	t.Parallel()
	a := stringAction("label", Multi)
	a.entries = append(a.entries, entry[string]{value: "a", ruleName: "r1"})
	a.entries = append(a.entries, entry[string]{value: "b", ruleName: "r2"})
	a.entries = append(a.entries, entry[string]{value: "c", ruleName: "r1"})
	a.resolve()

	vals := a.ByRule("r1")
	if len(vals) != 2 {
		t.Fatalf("got %d values, want 2", len(vals))
	}
	if vals[0] != "a" || vals[1] != "c" {
		t.Errorf("got %v, want [a c]", vals)
	}
}

func TestAction_ByRule_NoMatch(t *testing.T) {
	t.Parallel()
	a := stringAction("label", Multi)
	a.entries = append(a.entries, entry[string]{value: "a", ruleName: "r1"})
	a.resolve()

	vals := a.ByRule("r2")
	if vals == nil {
		t.Fatal("expected non-nil slice")
	}
	if len(vals) != 0 {
		t.Errorf("got %d values, want 0", len(vals))
	}
}

// --- ByTag ---

func TestAction_ByTag_Match(t *testing.T) {
	t.Parallel()
	a := stringAction("label", Multi)
	a.entries = append(a.entries, entry[string]{value: "a", ruleName: "r1", ruleTags: []string{"billing", "auto"}})
	a.entries = append(a.entries, entry[string]{value: "b", ruleName: "r2", ruleTags: []string{"manual"}})
	a.resolve()

	vals := a.ByTag("billing")
	if len(vals) != 1 || vals[0] != "a" {
		t.Errorf("got %v, want [a]", vals)
	}
}

func TestAction_ByTag_NoMatch(t *testing.T) {
	t.Parallel()
	a := stringAction("label", Multi)
	a.entries = append(a.entries, entry[string]{value: "a", ruleName: "r1", ruleTags: []string{"billing"}})
	a.resolve()

	vals := a.ByTag("nope")
	if vals == nil {
		t.Fatal("expected non-nil slice")
	}
	if len(vals) != 0 {
		t.Errorf("got %d values, want 0", len(vals))
	}
}

// --- Rules ---

func TestAction_Rules(t *testing.T) {
	t.Parallel()
	a := stringAction("label", Multi)
	a.entries = append(a.entries, entry[string]{value: "a", ruleName: "r1"})
	a.entries = append(a.entries, entry[string]{value: "b", ruleName: "r2"})
	a.entries = append(a.entries, entry[string]{value: "c", ruleName: "r1"}) // duplicate rule
	a.resolve()

	rules := a.Rules()
	if len(rules) != 2 {
		t.Fatalf("got %d rules, want 2", len(rules))
	}
	if rules[0] != "r1" || rules[1] != "r2" {
		t.Errorf("got %v, want [r1 r2]", rules)
	}
}

func TestAction_Rules_Empty(t *testing.T) {
	t.Parallel()
	a := stringAction("label", Multi)
	rules := a.Rules()
	if rules == nil {
		t.Fatal("expected non-nil slice")
	}
	if len(rules) != 0 {
		t.Errorf("got %d rules, want 0", len(rules))
	}
}
