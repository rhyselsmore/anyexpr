package action

import (
	"errors"
	"strings"
	"testing"
)

// --- Define: valid cases ---

func TestDefine_MinimalValid(t *testing.T) {
	t.Parallel()
	d, err := Define[string]("label")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Name() != "label" {
		t.Errorf("Name() = %q, want %q", d.Name(), "label")
	}
	if d.Description() != "" {
		t.Errorf("Description() = %q, want empty", d.Description())
	}
	if d.Terminal() {
		t.Error("Terminal[string]() = true, want false")
	}
	if d.Cardinality() != Single {
		t.Errorf("Cardinality() = %v, want Single", d.Cardinality())
	}
}

func TestDefine_WithAllOpts(t *testing.T) {
	t.Parallel()
	d, err := Define[string]("delete",
		WithDescription[string]("permanently removes the item"),
		WithCardinality[string](Multi),
		Terminal[string](true),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Name() != "delete" {
		t.Errorf("Name() = %q, want %q", d.Name(), "delete")
	}
	if d.Description() != "permanently removes the item" {
		t.Errorf("Description() = %q", d.Description())
	}
	if d.Cardinality() != Multi {
		t.Errorf("Cardinality() = %v, want Multi", d.Cardinality())
	}
	if !d.Terminal() {
		t.Error("Terminal[string]() = false, want true")
	}
}

func TestDefine_Underscore(t *testing.T) {
	t.Parallel()
	_, err := Define[string]("_private")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefine_Hyphen(t *testing.T) {
	t.Parallel()
	_, err := Define[string]("my-action")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefine_SingleCharName(t *testing.T) {
	t.Parallel()
	_, err := Define[string]("x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefine_DigitsAfterFirst(t *testing.T) {
	t.Parallel()
	_, err := Define[string]("rule42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefine_Description255(t *testing.T) {
	t.Parallel()
	_, err := Define[string]("ok", WithDescription[string](strings.Repeat("a", 255)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Define: name errors ---

func TestDefine_EmptyName(t *testing.T) {
	t.Parallel()
	_, err := Define[string]("")
	if !errors.Is(err, ErrNameEmpty) {
		t.Errorf("got %v, want ErrNameEmpty", err)
	}
}

func TestDefine_StartsWithDigit(t *testing.T) {
	t.Parallel()
	_, err := Define[string]("123abc")
	if !errors.Is(err, ErrNameInvalid) {
		t.Errorf("got %v, want ErrNameInvalid", err)
	}
}

func TestDefine_StartsWithHyphen(t *testing.T) {
	t.Parallel()
	_, err := Define[string]("-label")
	if !errors.Is(err, ErrNameInvalid) {
		t.Errorf("got %v, want ErrNameInvalid", err)
	}
}

func TestDefine_ContainsSpace(t *testing.T) {
	t.Parallel()
	_, err := Define[string]("my label")
	if !errors.Is(err, ErrNameInvalid) {
		t.Errorf("got %v, want ErrNameInvalid", err)
	}
}

func TestDefine_ContainsDot(t *testing.T) {
	t.Parallel()
	_, err := Define[string]("my.label")
	if !errors.Is(err, ErrNameInvalid) {
		t.Errorf("got %v, want ErrNameInvalid", err)
	}
}

func TestDefine_ContainsSlash(t *testing.T) {
	t.Parallel()
	_, err := Define[string]("my/label")
	if !errors.Is(err, ErrNameInvalid) {
		t.Errorf("got %v, want ErrNameInvalid", err)
	}
}

func TestDefine_ContainsAt(t *testing.T) {
	t.Parallel()
	_, err := Define[string]("label@thing")
	if !errors.Is(err, ErrNameInvalid) {
		t.Errorf("got %v, want ErrNameInvalid", err)
	}
}

// --- Define: description errors ---

func TestDefine_DescriptionTooLong(t *testing.T) {
	t.Parallel()
	_, err := Define[string]("ok", WithDescription[string](strings.Repeat("a", 256)))
	if !errors.Is(err, ErrDescriptionTooLong) {
		t.Errorf("got %v, want ErrDescriptionTooLong", err)
	}
}

func TestDefine_DescriptionWayTooLong(t *testing.T) {
	t.Parallel()
	_, err := Define[string]("ok", WithDescription[string](strings.Repeat("x", 1000)))
	if !errors.Is(err, ErrDescriptionTooLong) {
		t.Errorf("got %v, want ErrDescriptionTooLong", err)
	}
}

// --- Define: cardinality errors ---

func TestDefine_InvalidCardinality(t *testing.T) {
	t.Parallel()
	_, err := Define[string]("ok", WithCardinality[string](Cardinality(99)))
	if err == nil {
		t.Fatal("expected error for invalid cardinality")
	}
}

func TestDefine_SingleValid(t *testing.T) {
	t.Parallel()
	d, err := Define[string]("ok", WithCardinality[string](Single))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Cardinality() != Single {
		t.Errorf("Cardinality() = %v, want Single", d.Cardinality())
	}
}

func TestDefine_MultiValid(t *testing.T) {
	t.Parallel()
	d, err := Define[string]("ok", WithCardinality[string](Multi))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Cardinality() != Multi {
		t.Errorf("Cardinality() = %v, want Multi", d.Cardinality())
	}
}

// --- Define: terminal ---

func TestDefine_TerminalTrue(t *testing.T) {
	t.Parallel()
	d, _ := Define[string]("del", Terminal[string](true))
	if !d.Terminal() {
		t.Error("Terminal[string]() = false, want true")
	}
}

func TestDefine_TerminalFalse(t *testing.T) {
	t.Parallel()
	d, _ := Define[string]("label", Terminal[string](false))
	if d.Terminal() {
		t.Error("Terminal[string]() = true, want false")
	}
}

func TestDefine_TerminalDefault(t *testing.T) {
	t.Parallel()
	d, _ := Define[string]("label")
	if d.Terminal() {
		t.Error("Terminal[string]() should default to false")
	}
}

// --- MustDefine ---

func TestMustDefine_Valid(t *testing.T) {
	t.Parallel()
	d := MustDefine[string]("label", WithCardinality[string](Multi))
	if d.Name() != "label" {
		t.Errorf("Name() = %q, want %q", d.Name(), "label")
	}
}

func TestMustDefine_Panics(t *testing.T) {
	t.Parallel()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		s, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", r)
		}
		if !strings.Contains(s, "action.MustDefine") {
			t.Errorf("panic message %q doesn't mention MustDefine", s)
		}
	}()
	MustDefine[string]("")
}

func TestMustDefine_PanicsOnInvalidName(t *testing.T) {
	t.Parallel()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	MustDefine[string]("123bad")
}
