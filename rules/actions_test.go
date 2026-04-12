package rules

import (
	"errors"
	"testing"
)

// Test env type — used as E parameter.
type testEnv struct{ Name string }

// Test actions struct — generic on E.
type testActions[E any] struct {
	Label    Action[string, E]  `rule:"label,multi"`
	Read     Action[bool, E]    `rule:"read"`
	Move     Action[string, E]  `rule:"move"`
	Priority Action[int, E]     `rule:"priority"`
	Score    Action[float64, E] `rule:"score"`
	Delete   Action[NoArgs, E]  `rule:"delete,terminal"`
}

func TestDefineActions_Valid(t *testing.T) {
	t.Parallel()
	actions, err := DefineActions[testActions[testEnv], testEnv]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !actions.defined {
		t.Fatal("expected defined to be true")
	}
	if len(actions.fields) != 6 {
		t.Fatalf("expected 6 fields, got %d", len(actions.fields))
	}
}

func TestDefineActions_FieldsConfigured(t *testing.T) {
	t.Parallel()
	actions, err := DefineActions[testActions[testEnv], testEnv]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schema := actions.schema
	checks := []struct {
		name        string
		cardinality Cardinality
		terminal    bool
	}{
		{"label", Multi, false},
		{"read", Single, false},
		{"move", Single, false},
		{"priority", Single, false},
		{"score", Single, false},
		{"delete", Single, true},
	}

	_ = schema
	for _, tt := range checks {
		af, ok := actions.compilers[tt.name]
		if !ok {
			t.Errorf("action %q not found in compilers", tt.name)
			continue
		}
		// Compile with a dummy value to get metadata via the actionValuer.
		var av actionValuer[testEnv]
		var err error
		switch tt.name {
		case "label", "move":
			av, err = af.compile("test")
		case "read":
			av, err = af.compile(true)
		case "priority":
			av, err = af.compile(1)
		case "score":
			av, err = af.compile(1.0)
		case "delete":
			av, err = af.compile(NoArgs{})
		}
		if err != nil {
			t.Errorf("action %q: compile error: %v", tt.name, err)
			continue
		}
		if av.actionCardinality() != tt.cardinality {
			t.Errorf("action %q: cardinality = %v, want %v", tt.name, av.actionCardinality(), tt.cardinality)
		}
		if av.actionTerminal() != tt.terminal {
			t.Errorf("action %q: terminal = %v, want %v", tt.name, av.actionTerminal(), tt.terminal)
		}
	}
}

func TestDefineActions_SchemaConfigured(t *testing.T) {
	t.Parallel()
	actions, err := DefineActions[testActions[testEnv], testEnv]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if actions.schema.Label.Name() != "label" {
		t.Errorf("Label.Name() = %q, want %q", actions.schema.Label.Name(), "label")
	}
	if actions.schema.Delete.Name() != "delete" {
		t.Errorf("Delete.Name() = %q, want %q", actions.schema.Delete.Name(), "delete")
	}
}

func TestDefineActions_DuplicateNames(t *testing.T) {
	t.Parallel()
	type bad[E any] struct {
		A Action[string, E] `rule:"same"`
		B Action[string, E] `rule:"same"`
	}
	_, err := DefineActions[bad[testEnv], testEnv]()
	if !errors.Is(err, ErrDuplicateRegistration) {
		t.Errorf("got %v, want ErrDuplicateRegistration", err)
	}
}

func TestDefineActions_MissingTag(t *testing.T) {
	t.Parallel()
	type bad[E any] struct {
		A Action[string, E] `rule:"ok"`
		B Action[string, E] // no tag
	}
	_, err := DefineActions[bad[testEnv], testEnv]()
	if !errors.Is(err, ErrDefine) {
		t.Errorf("got %v, want ErrDefine", err)
	}
}

func TestDefineActions_InvalidName(t *testing.T) {
	t.Parallel()
	type bad[E any] struct {
		A Action[string, E] `rule:"123invalid"`
	}
	_, err := DefineActions[bad[testEnv], testEnv]()
	if !errors.Is(err, ErrDefine) {
		t.Errorf("got %v, want ErrDefine", err)
	}
}

func TestDefineActions_EmptyTag(t *testing.T) {
	t.Parallel()
	type bad[E any] struct {
		A Action[string, E] `rule:""`
	}
	_, err := DefineActions[bad[testEnv], testEnv]()
	if !errors.Is(err, ErrDefine) {
		t.Errorf("got %v, want ErrDefine", err)
	}
}

func TestDefineActions_MultipleTerminals(t *testing.T) {
	t.Parallel()
	type bad[E any] struct {
		A Action[NoArgs, E] `rule:"a,terminal"`
		B Action[NoArgs, E] `rule:"b,terminal"`
	}
	_, err := DefineActions[bad[testEnv], testEnv]()
	if !errors.Is(err, ErrMultipleTerminals) {
		t.Errorf("got %v, want ErrMultipleTerminals", err)
	}
}

func TestDefineActions_UnknownOption(t *testing.T) {
	t.Parallel()
	type bad[E any] struct {
		A Action[string, E] `rule:"a,bogus"`
	}
	_, err := DefineActions[bad[testEnv], testEnv]()
	if !errors.Is(err, ErrDefine) {
		t.Errorf("got %v, want ErrDefine", err)
	}
}

func TestDefineActions_NoActionFields(t *testing.T) {
	t.Parallel()
	type bad struct {
		Name string
	}
	_, err := DefineActions[bad, testEnv]()
	if !errors.Is(err, ErrDefine) {
		t.Errorf("got %v, want ErrDefine", err)
	}
}

func TestDefineActions_HyphenInName(t *testing.T) {
	t.Parallel()
	type ok[E any] struct {
		A Action[string, E] `rule:"my-action"`
	}
	_, err := DefineActions[ok[testEnv], testEnv]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefineActions_UnderscoreInName(t *testing.T) {
	t.Parallel()
	type ok[E any] struct {
		A Action[string, E] `rule:"my_action"`
	}
	_, err := DefineActions[ok[testEnv], testEnv]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Compile type checking ---

func TestDefineActions_Compile_TypeMismatch(t *testing.T) {
	t.Parallel()
	actions, err := DefineActions[testActions[testEnv], testEnv]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	af := actions.compilers["read"]
	_, err = af.compile("not-a-bool")
	if !errors.Is(err, ErrValueType) {
		t.Errorf("got %v, want ErrValueType", err)
	}
}

func TestDefineActions_Compile_TypeMatch(t *testing.T) {
	t.Parallel()
	actions, err := DefineActions[testActions[testEnv], testEnv]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	af := actions.compilers["label"]
	av, err := af.compile("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if av.actionName() != "label" {
		t.Errorf("got %q, want %q", av.actionName(), "label")
	}
	if av.stringValue() != "hello" {
		t.Errorf("got %q, want %q", av.stringValue(), "hello")
	}
}
