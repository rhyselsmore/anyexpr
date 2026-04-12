package rules

import (
	"errors"
	"testing"

	"github.com/rhyselsmore/anyexpr"
)

type compileEnv struct {
	Name   string
	Amount float64
	Active bool
}

type compileActions[E any] struct {
	Label    Action[string, E]  `rule:"label,multi"`
	Category Action[string, E]  `rule:"category"`
	Read     Action[bool, E]    `rule:"read"`
	Priority Action[int, E]     `rule:"priority"`
	Score    Action[float64, E] `rule:"score"`
	Delete   Action[NoArgs, E]  `rule:"delete,terminal"`
}

func compileSetup(t *testing.T) (*Actions[compileActions[compileEnv], compileEnv], *anyexpr.Compiler[compileEnv]) {
	t.Helper()
	actions, err := DefineActions[compileActions[compileEnv], compileEnv]()
	if err != nil {
		t.Fatalf("DefineActions: %v", err)
	}
	compiler, err := anyexpr.NewCompiler[compileEnv]()
	if err != nil {
		t.Fatalf("NewCompiler: %v", err)
	}
	return actions, compiler
}

func TestCompile_Valid(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	rs, err := Compile(actions, compiler, []Definition{
		{
			Name: "r1",
			When: `has(Name, "alice")`,
			Then: []ActionEntry{
				{Name: "label", Value: "friend"},
				{Name: "read", Value: true},
			},
		},
		{
			Name: "r2",
			When: `Amount > 100`,
			Then: []ActionEntry{
				{Name: "category", Value: "large"},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rs.Len() != 2 {
		t.Errorf("got %d rules, want 2", rs.Len())
	}
}

func TestCompile_DuplicateRuleNames(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(actions, compiler, []Definition{
		{Name: "r1", When: `Active`},
		{Name: "r1", When: `Active`},
	})
	if !errors.Is(err, ErrDuplicateRule) {
		t.Errorf("got %v, want ErrDuplicateRule", err)
	}
}

func TestCompile_UnknownAction(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(actions, compiler, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "nope", Value: "x"}},
		},
	})
	if !errors.Is(err, ErrUnknownAction) {
		t.Errorf("got %v, want ErrUnknownAction", err)
	}
}

func TestCompile_BadExpression(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(actions, compiler, []Definition{
		{Name: "r1", When: `invalid!!!`},
	})
	if !errors.Is(err, ErrCompile) {
		t.Errorf("got %v, want ErrCompile", err)
	}
}

func TestCompile_BoolValueValid(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(actions, compiler, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "read", Value: true}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompile_BoolValueInvalid(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(actions, compiler, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "read", Value: "banana"}},
		},
	})
	if !errors.Is(err, ErrValueType) {
		t.Errorf("got %v, want ErrValueType", err)
	}
}

func TestCompile_IntValueValid(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(actions, compiler, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "priority", Value: 42}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompile_IntValueInvalid(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(actions, compiler, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "priority", Value: "notint"}},
		},
	})
	if !errors.Is(err, ErrValueType) {
		t.Errorf("got %v, want ErrValueType", err)
	}
}

func TestCompile_Float64ValueValid(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(actions, compiler, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "score", Value: 0.95}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompile_Float64ValueInvalid(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(actions, compiler, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "score", Value: "nope"}},
		},
	})
	if !errors.Is(err, ErrValueType) {
		t.Errorf("got %v, want ErrValueType", err)
	}
}

func TestCompile_NoArgsWithValue(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(actions, compiler, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "delete", Value: "oops"}},
		},
	})
	if !errors.Is(err, ErrValueType) {
		t.Errorf("got %v, want ErrValueType", err)
	}
}

func TestCompile_NoArgsEmpty(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(actions, compiler, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "delete", Value: NoArgs{}}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompile_CardinalityViolation(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(actions, compiler, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{
				{Name: "category", Value: "a"},
				{Name: "category", Value: "b"}, // single used twice
			},
		},
	})
	if !errors.Is(err, ErrCardinalityViolation) {
		t.Errorf("got %v, want ErrCardinalityViolation", err)
	}
}

func TestCompile_MultiAllowedMultipleTimes(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(actions, compiler, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{
				{Name: "label", Value: "a"},
				{Name: "label", Value: "b"},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompile_TerminalImpliesStop(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	rs, err := Compile(actions, compiler, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "delete", Value: NoArgs{}}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rs.rules[0].stop {
		t.Error("expected stop to be true for terminal action")
	}
}

func TestCompile_EmptyDefinitions(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	rs, err := Compile(actions, compiler, []Definition{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rs.Len() != 0 {
		t.Errorf("got %d rules, want 0", rs.Len())
	}
}

func TestCompile_NotDefined(t *testing.T) {
	t.Parallel()
	compiler, _ := anyexpr.NewCompiler[compileEnv]()
	actions := &Actions[compileActions[compileEnv], compileEnv]{}
	_, err := Compile(actions, compiler, nil)
	if !errors.Is(err, ErrNotDefined) {
		t.Errorf("got %v, want ErrNotDefined", err)
	}
}

// --- Names / Len ---

func TestRuleset_Names(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	rs, _ := Compile(actions, compiler, []Definition{
		{Name: "a", When: `Active`},
		{Name: "b", When: `Active`},
	})
	names := rs.Names()
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Errorf("got %v, want [a b]", names)
	}
}

// --- Merge ---

func TestRuleset_Merge_NoCollision(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	a, _ := Compile(actions, compiler, []Definition{{Name: "a", When: `Active`}})
	b, _ := Compile(actions, compiler, []Definition{{Name: "b", When: `Active`}})

	merged, err := a.Merge(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merged.Len() != 2 {
		t.Errorf("got %d rules, want 2", merged.Len())
	}
}

func TestRuleset_Merge_Collision(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	a, _ := Compile(actions, compiler, []Definition{{Name: "x", When: `Active`}})
	b, _ := Compile(actions, compiler, []Definition{{Name: "x", When: `Active`}})

	_, err := a.Merge(b)
	if !errors.Is(err, ErrNameCollision) {
		t.Errorf("got %v, want ErrNameCollision", err)
	}
}

func TestRuleset_Merge_AllowOverride(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	a, _ := Compile(actions, compiler, []Definition{{Name: "x", When: `Active`}})
	b, _ := Compile(actions, compiler, []Definition{{Name: "x", When: `!Active`}})

	merged, err := a.Merge(b, AllowOverride)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merged.Len() != 1 {
		t.Errorf("got %d rules, want 1", merged.Len())
	}
}
