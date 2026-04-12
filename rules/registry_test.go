package rules

import (
	"errors"
	"sync"
	"testing"
)

func TestNewRegistry_NoOpts(t *testing.T) {
	t.Parallel()
	r, err := NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if r == nil {
		t.Fatal("registry is nil")
	}
}

func TestNewRegistry_WithAction(t *testing.T) {
	t.Parallel()
	r, err := NewRegistry(WithAction("tag", Multi, StringVal, false))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.LookupAction("tag"); !ok {
		t.Error("action not found")
	}
}

func TestNewRegistry_WithHandler(t *testing.T) {
	t.Parallel()
	r, err := NewRegistry(WithHandler("h1", "placeholder", Multi, false))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.LookupAction("h1"); !ok {
		t.Error("handler action def not found")
	}
	if _, ok := r.LookupHandler("h1"); !ok {
		t.Error("handler not found")
	}
}

func TestNewRegistry_DuplicateAction(t *testing.T) {
	t.Parallel()
	_, err := NewRegistry(
		WithAction("tag", Multi, StringVal, false),
		WithAction("tag", Single, BoolValue, false),
	)
	if !errors.Is(err, ErrDuplicateRegistration) {
		t.Errorf("got %v, want ErrDuplicateRegistration", err)
	}
}

func TestNewRegistry_DuplicateHandler(t *testing.T) {
	t.Parallel()
	_, err := NewRegistry(
		WithHandler("h1", "a", Multi, false),
		WithHandler("h1", "b", Multi, false),
	)
	if !errors.Is(err, ErrDuplicateRegistration) {
		t.Errorf("got %v, want ErrDuplicateRegistration", err)
	}
}

func TestNewRegistry_ActionAndHandlerSameName(t *testing.T) {
	t.Parallel()
	_, err := NewRegistry(
		WithAction("x", Multi, StringVal, false),
		WithHandler("x", "fn", Multi, false),
	)
	if !errors.Is(err, ErrDuplicateRegistration) {
		t.Errorf("got %v, want ErrDuplicateRegistration", err)
	}
}

func TestRegistry_With_AddsToParent(t *testing.T) {
	t.Parallel()
	base, _ := NewRegistry(WithAction("tag", Multi, StringVal, false))
	ext, err := base.With(WithAction("category", Single, StringVal, false))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ext.LookupAction("category"); !ok {
		t.Error("new action not found")
	}
	if _, ok := ext.LookupAction("tag"); !ok {
		t.Error("parent action not found through child")
	}
}

func TestRegistry_With_ParentUnmodified(t *testing.T) {
	t.Parallel()
	base, _ := NewRegistry(WithAction("tag", Multi, StringVal, false))
	base.With(WithAction("category", Single, StringVal, false))
	if _, ok := base.LookupAction("category"); ok {
		t.Error("parent should not have child's action")
	}
}

func TestRegistry_With_NameCollision(t *testing.T) {
	t.Parallel()
	base, _ := NewRegistry(WithAction("tag", Multi, StringVal, false))
	_, err := base.With(WithAction("tag", Single, StringVal, false))
	if !errors.Is(err, ErrDuplicateRegistration) {
		t.Errorf("got %v, want ErrDuplicateRegistration", err)
	}
}

func TestRegistry_With_GrandchildWorks(t *testing.T) {
	t.Parallel()
	base, _ := NewRegistry(WithAction("a", Multi, StringVal, false))
	child, _ := base.With(WithAction("b", Multi, StringVal, false))
	grand, err := child.With(WithAction("c", Multi, StringVal, false))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a", "b", "c"} {
		if _, ok := grand.LookupAction(name); !ok {
			t.Errorf("action %q not found", name)
		}
	}
}

func TestRegistry_With_LookupWalksParent(t *testing.T) {
	t.Parallel()
	base, _ := NewRegistry(WithHandler("h1", "fn", Multi, false))
	child, _ := base.With(WithAction("extra", Multi, StringVal, false))
	if _, ok := child.LookupHandler("h1"); !ok {
		t.Error("handler not found through parent")
	}
}

func TestRegistry_LookupAction_NotFound(t *testing.T) {
	t.Parallel()
	r, _ := NewRegistry()
	if _, ok := r.LookupAction("nope"); ok {
		t.Error("expected not found")
	}
}

func TestRegistry_LookupHandler_NotFound(t *testing.T) {
	t.Parallel()
	r, _ := NewRegistry()
	if _, ok := r.LookupHandler("nope"); ok {
		t.Error("expected not found")
	}
}

func TestRegistry_ActionNames(t *testing.T) {
	t.Parallel()
	base, _ := NewRegistry(WithAction("b", Multi, StringVal, false))
	child, _ := base.With(WithAction("a", Multi, StringVal, false))
	names := child.ActionNames()
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Errorf("got %v, want [a b]", names)
	}
}

func TestRegistry_HandlerNames(t *testing.T) {
	t.Parallel()
	base, _ := NewRegistry(WithHandler("h2", "fn", Multi, false))
	child, _ := base.With(WithHandler("h1", "fn", Multi, false))
	names := child.HandlerNames()
	if len(names) != 2 || names[0] != "h1" || names[1] != "h2" {
		t.Errorf("got %v, want [h1 h2]", names)
	}
}

func TestRegistry_ConcurrentLookup(t *testing.T) {
	t.Parallel()
	r, _ := NewRegistry(
		WithAction("tag", Multi, StringVal, false),
		WithHandler("h1", "fn", Multi, false),
	)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.LookupAction("tag")
			r.LookupHandler("h1")
			r.ActionNames()
			r.HandlerNames()
		}()
	}
	wg.Wait()
}
