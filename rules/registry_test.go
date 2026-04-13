package rules

import (
	"errors"
	"testing"

	"github.com/rhyselsmore/anyexpr"
)

func registrySetup(t *testing.T) *Registry[testEnv, testActions[testEnv]] {
	t.Helper()
	actions := defineTestActions(t)
	compiler, err := anyexpr.NewCompiler[testEnv]()
	if err != nil {
		t.Fatalf("NewCompiler: %v", err)
	}
	reg, err := NewRegistry(compiler, actions)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	return reg
}

// --- NewRegistry ---

func TestNewRegistry_Valid(t *testing.T) {
	t.Parallel()
	reg := registrySetup(t)
	if reg.Len() != 0 {
		t.Errorf("Len() = %d, want 0", reg.Len())
	}
}

func TestNewRegistry_ActionsZero(t *testing.T) {
	t.Parallel()
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	_, err := NewRegistry(compiler, &Actions[testEnv, testActions[testEnv]]{})
	if !errors.Is(err, ErrActionsZero) {
		t.Errorf("got %v, want ErrActionsZero", err)
	}
}

// --- Add ---

func TestRegistry_Add(t *testing.T) {
	t.Parallel()
	reg := registrySetup(t)
	err := reg.Add(Definition{Name: "r1", When: `Active`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg.Len() != 1 {
		t.Errorf("Len() = %d, want 1", reg.Len())
	}
}

func TestRegistry_Add_Duplicate(t *testing.T) {
	t.Parallel()
	reg := registrySetup(t)
	reg.Add(Definition{Name: "r1", When: `Active`})
	err := reg.Add(Definition{Name: "r1", When: `Active`})
	if !errors.Is(err, ErrDefinitionDuplicate) {
		t.Errorf("got %v, want ErrDefinitionDuplicate", err)
	}
}

// --- Update ---

func TestRegistry_Update(t *testing.T) {
	t.Parallel()
	reg := registrySetup(t)
	reg.Add(Definition{Name: "r1", When: `Active`})
	err := reg.Update(Definition{Name: "r1", When: `!Active`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegistry_Update_Unknown(t *testing.T) {
	t.Parallel()
	reg := registrySetup(t)
	err := reg.Update(Definition{Name: "nope", When: `Active`})
	if !errors.Is(err, ErrUnknownDefinition) {
		t.Errorf("got %v, want ErrUnknownDefinition", err)
	}
}

// --- Upsert ---

func TestRegistry_Upsert_Add(t *testing.T) {
	t.Parallel()
	reg := registrySetup(t)
	reg.Upsert(Definition{Name: "r1", When: `Active`})
	if reg.Len() != 1 {
		t.Errorf("Len() = %d, want 1", reg.Len())
	}
}

func TestRegistry_Upsert_Update(t *testing.T) {
	t.Parallel()
	reg := registrySetup(t)
	reg.Upsert(Definition{Name: "r1", When: `Active`})
	reg.Upsert(Definition{Name: "r1", When: `!Active`})
	if reg.Len() != 1 {
		t.Errorf("Len() = %d, want 1", reg.Len())
	}
}

// --- Remove ---

func TestRegistry_Remove(t *testing.T) {
	t.Parallel()
	reg := registrySetup(t)
	reg.Add(Definition{Name: "r1", When: `Active`})
	reg.Add(Definition{Name: "r2", When: `Active`})
	reg.Remove("r1")
	if reg.Len() != 1 {
		t.Errorf("Len() = %d, want 1", reg.Len())
	}
}

func TestRegistry_Remove_Unknown(t *testing.T) {
	t.Parallel()
	reg := registrySetup(t)
	reg.Remove("nope") // should not panic
	if reg.Len() != 0 {
		t.Errorf("Len() = %d, want 0", reg.Len())
	}
}

// --- Definitions ---

func TestRegistry_Definitions_Order(t *testing.T) {
	t.Parallel()
	reg := registrySetup(t)
	reg.Add(
		Definition{Name: "a", When: `Active`},
		Definition{Name: "b", When: `Active`},
		Definition{Name: "c", When: `Active`},
	)
	defs := reg.Definitions()
	if len(defs) != 3 {
		t.Fatalf("got %d defs, want 3", len(defs))
	}
	if defs[0].Name != "a" || defs[1].Name != "b" || defs[2].Name != "c" {
		t.Errorf("order: %s, %s, %s", defs[0].Name, defs[1].Name, defs[2].Name)
	}
}

func TestRegistry_Definitions_AfterRemove(t *testing.T) {
	t.Parallel()
	reg := registrySetup(t)
	reg.Add(
		Definition{Name: "a", When: `Active`},
		Definition{Name: "b", When: `Active`},
		Definition{Name: "c", When: `Active`},
	)
	reg.Remove("b")
	defs := reg.Definitions()
	if len(defs) != 2 {
		t.Fatalf("got %d defs, want 2", len(defs))
	}
	if defs[0].Name != "a" || defs[1].Name != "c" {
		t.Errorf("order: %s, %s", defs[0].Name, defs[1].Name)
	}
}

// --- Compile ---

func TestRegistry_Compile(t *testing.T) {
	t.Parallel()
	reg := registrySetup(t)
	reg.Add(
		Definition{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "label", Value: "a"}},
		},
	)
	prog, err := reg.Compile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prog.IsZero() {
		t.Error("program should not be zero")
	}
}

func TestRegistry_Compile_Empty(t *testing.T) {
	t.Parallel()
	reg := registrySetup(t)
	_, err := reg.Compile()
	if !errors.Is(err, ErrNoDefinitions) {
		t.Errorf("got %v, want ErrNoDefinitions", err)
	}
}
