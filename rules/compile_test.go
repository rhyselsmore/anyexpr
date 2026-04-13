package rules

import (
	"errors"
	"testing"

	"github.com/rhyselsmore/anyexpr"
	"github.com/rhyselsmore/anyexpr/rules/action"
)

func compileSetup(t *testing.T) (*Actions[testEnv, testActions[testEnv]], *anyexpr.Compiler[testEnv]) {
	t.Helper()
	actions := defineTestActions(t)
	compiler, err := anyexpr.NewCompiler[testEnv]()
	if err != nil {
		t.Fatalf("NewCompiler: %v", err)
	}
	return actions, compiler
}

// --- Compile: valid ---

func TestCompile_Valid(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	prog, err := Compile(compiler, actions, []Definition{
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
				{Name: "priority", Value: 5},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prog.IsZero() {
		t.Error("program should not be zero")
	}
}

// --- Compile: errors ---

func TestCompile_ActionsZero(t *testing.T) {
	t.Parallel()
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	_, err := Compile(compiler, &Actions[testEnv, testActions[testEnv]]{}, nil)
	if !errors.Is(err, ErrActionsZero) {
		t.Errorf("got %v, want ErrActionsZero", err)
	}
}

func TestCompile_NoDefinitions(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(compiler, actions, []Definition{})
	if !errors.Is(err, ErrNoDefinitions) {
		t.Errorf("got %v, want ErrNoDefinitions", err)
	}
}

func TestCompile_DuplicateRuleNames(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(compiler, actions, []Definition{
		{Name: "r1", When: `Active`},
		{Name: "r1", When: `Active`},
	})
	if !errors.Is(err, ErrDefinitionDuplicate) {
		t.Errorf("got %v, want ErrDefinitionDuplicate", err)
	}
}

func TestCompile_BadExpression(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(compiler, actions, []Definition{
		{Name: "r1", When: `invalid!!!`},
	})
	if !errors.Is(err, ErrCompile) {
		t.Errorf("got %v, want ErrCompile", err)
	}
}

func TestCompile_UnknownAction(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(compiler, actions, []Definition{
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

func TestCompile_ValueTypeMismatch(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(compiler, actions, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "read", Value: "not-a-bool"}},
		},
	})
	if !errors.Is(err, ErrActionValueType) {
		t.Errorf("got %v, want ErrActionValueType", err)
	}
}

func TestCompile_CardinalityViolation(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(compiler, actions, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{
				{Name: "move", Value: "a"},
				{Name: "move", Value: "b"}, // single used twice
			},
		},
	})
	if !errors.Is(err, ErrCardinalityViolation) {
		t.Errorf("got %v, want ErrCardinalityViolation", err)
	}
}

func TestCompile_MultiAllowed(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(compiler, actions, []Definition{
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

func TestCompile_MultipleTerminalsInRule(t *testing.T) {
	t.Parallel()
	// Need two terminal actions — use a custom struct.
	type twoTerminals[E any] struct {
		Del1 Action[action.NoArgs, E] `rule:"del1,terminal"`
		Del2 Action[action.NoArgs, E] `rule:"del2,terminal"`
	}
	actions, err := DefineActions[testEnv, twoTerminals[testEnv]]()
	if err != nil {
		t.Fatalf("DefineActions: %v", err)
	}
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	_, err = Compile(compiler, actions, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{
				{Name: "del1"},
				{Name: "del2"},
			},
		},
	})
	if !errors.Is(err, ErrMultipleTerminals) {
		t.Errorf("got %v, want ErrMultipleTerminals", err)
	}
}

func TestCompile_TerminalImpliesStop(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	prog, err := Compile(compiler, actions, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "delete"}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !prog.rules[0].stop {
		t.Error("expected stop for terminal action")
	}
}

func TestCompile_NoArgsAcceptsNil(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(compiler, actions, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "delete"}}, // Value is nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompile_NoArgsAcceptsExplicit(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := Compile(compiler, actions, []Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "delete", Value: action.NoArgs{}}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Program ---

func TestProgram_IsZero_Nil(t *testing.T) {
	t.Parallel()
	var p *Program[testEnv, testActions[testEnv]]
	if !p.IsZero() {
		t.Error("nil program should be zero")
	}
}

func TestProgram_IsZero_Uninitialised(t *testing.T) {
	t.Parallel()
	p := &Program[testEnv, testActions[testEnv]]{}
	if !p.IsZero() {
		t.Error("uninitialised program should be zero")
	}
}

// --- Definition order ---

func TestCompile_PreservesOrder(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	prog, err := Compile(compiler, actions, []Definition{
		{Name: "first", When: `Active`},
		{Name: "second", When: `Active`},
		{Name: "third", When: `Active`},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prog.rules) != 3 {
		t.Fatalf("got %d rules, want 3", len(prog.rules))
	}
	if prog.rules[0].name != "first" || prog.rules[1].name != "second" || prog.rules[2].name != "third" {
		t.Errorf("order: %s, %s, %s", prog.rules[0].name, prog.rules[1].name, prog.rules[2].name)
	}
}
