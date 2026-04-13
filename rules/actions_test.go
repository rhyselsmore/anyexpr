package rules

import (
	"errors"
	"testing"

	"github.com/rhyselsmore/anyexpr/rules2/action"
)

// --- Test types ---

type testEnv struct {
	Name   string
	Amount float64
	Active bool
}

type testActions[E any] struct {
	Label    Action[string, E]       `rule:"label,multi" description:"categorisation labels"`
	Move     Action[string, E]       `rule:"move"`
	Read     Action[bool, E]         `rule:"read"`
	Priority Action[int, E]          `rule:"priority"`
	Score    Action[float64, E]      `rule:"score"`
	Delete   Action[action.NoArgs, E] `rule:"delete,terminal"`
}

func defineTestActions(t *testing.T) *Actions[testEnv, testActions[testEnv]] {
	t.Helper()
	actions, err := DefineActions[testEnv, testActions[testEnv]]()
	if err != nil {
		t.Fatalf("DefineActions: %v", err)
	}
	return actions
}

// --- DefineActions ---

func TestDefineActions_Valid(t *testing.T) {
	t.Parallel()
	actions := defineTestActions(t)
	if actions.IsZero() {
		t.Fatal("expected defined")
	}
}

func TestDefineActions_FieldCount(t *testing.T) {
	t.Parallel()
	actions := defineTestActions(t)
	if len(actions.fields) != 6 {
		t.Errorf("got %d fields, want 6", len(actions.fields))
	}
}

func TestDefineActions_DuplicateNames(t *testing.T) {
	t.Parallel()
	type bad[E any] struct {
		A Action[string, E] `rule:"same"`
		B Action[string, E] `rule:"same"`
	}
	_, err := DefineActions[testEnv, bad[testEnv]]()
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
	_, err := DefineActions[testEnv, bad[testEnv]]()
	if !errors.Is(err, ErrDefine) {
		t.Errorf("got %v, want ErrDefine", err)
	}
}

func TestDefineActions_EmptyTag(t *testing.T) {
	t.Parallel()
	type bad[E any] struct {
		A Action[string, E] `rule:""`
	}
	_, err := DefineActions[testEnv, bad[testEnv]]()
	if !errors.Is(err, ErrDefine) {
		t.Errorf("got %v, want ErrDefine", err)
	}
}

func TestDefineActions_InvalidTagOption(t *testing.T) {
	t.Parallel()
	type bad[E any] struct {
		A Action[string, E] `rule:"ok,bogus"`
	}
	_, err := DefineActions[testEnv, bad[testEnv]]()
	if !errors.Is(err, ErrDefine) {
		t.Errorf("got %v, want ErrDefine", err)
	}
}

func TestDefineActions_NoActionFields(t *testing.T) {
	t.Parallel()
	type bad struct {
		Name string
	}
	_, err := DefineActions[testEnv, bad]()
	if !errors.Is(err, ErrDefine) {
		t.Errorf("got %v, want ErrDefine", err)
	}
}

func TestDefineActions_IsZero_Nil(t *testing.T) {
	t.Parallel()
	var a *Actions[testEnv, testActions[testEnv]]
	if !a.IsZero() {
		t.Error("nil Actions should be zero")
	}
}

func TestDefineActions_IsZero_Uninitialised(t *testing.T) {
	t.Parallel()
	a := &Actions[testEnv, testActions[testEnv]]{}
	if !a.IsZero() {
		t.Error("uninitialised Actions should be zero")
	}
}

// --- Describe ---

func TestDescribe_ReturnsAllActions(t *testing.T) {
	t.Parallel()
	actions := defineTestActions(t)
	infos := actions.Describe()
	if len(infos) != 6 {
		t.Fatalf("got %d infos, want 6", len(infos))
	}
}

func TestDescribe_ActionInfo(t *testing.T) {
	t.Parallel()
	actions := defineTestActions(t)
	infos := actions.Describe()

	// Find label.
	var label ActionInfo
	for _, info := range infos {
		if info.Name == "label" {
			label = info
			break
		}
	}

	if label.Name != "label" {
		t.Fatal("label action not found")
	}
	if label.Cardinality != action.Multi {
		t.Errorf("cardinality = %v, want Multi", label.Cardinality)
	}
	if label.Terminal {
		t.Error("label should not be terminal")
	}
	if label.ValueType != "string" {
		t.Errorf("value type = %q, want string", label.ValueType)
	}
	if label.Description != "categorisation labels" {
		t.Errorf("description = %q, want %q", label.Description, "categorisation labels")
	}
}

func TestDescribe_TerminalAction(t *testing.T) {
	t.Parallel()
	actions := defineTestActions(t)
	infos := actions.Describe()

	var del ActionInfo
	for _, info := range infos {
		if info.Name == "delete" {
			del = info
			break
		}
	}

	if !del.Terminal {
		t.Error("delete should be terminal")
	}
	if del.ValueType != "action.NoArgs" {
		t.Errorf("value type = %q, want action.NoArgs", del.ValueType)
	}
}

func TestDescribe_NoDescription(t *testing.T) {
	t.Parallel()
	actions := defineTestActions(t)
	infos := actions.Describe()

	var move ActionInfo
	for _, info := range infos {
		if info.Name == "move" {
			move = info
			break
		}
	}

	if move.Description != "" {
		t.Errorf("description = %q, want empty", move.Description)
	}
}
