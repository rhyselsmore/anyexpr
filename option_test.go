package anyexpr

import (
	"errors"
	"testing"
)

func TestWithFunction_Valid(t *testing.T) {
	t.Parallel()
	cfg := &compilerConfig{}
	opt := WithFunction("myfunc", func(params ...any) (any, error) { return nil, nil })
	if err := opt(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cfg.customFuncs["myfunc"]; !ok {
		t.Error("function not registered")
	}
}

func TestWithFunction_DuplicateName(t *testing.T) {
	t.Parallel()
	cfg := &compilerConfig{}
	fn := func(params ...any) (any, error) { return nil, nil }
	opt := WithFunction("myfunc", fn)
	if err := opt(cfg); err != nil {
		t.Fatalf("first registration: %v", err)
	}
	opt2 := WithFunction("myfunc", fn)
	err := opt2(cfg)
	if !errors.Is(err, ErrDuplicateFunction) {
		t.Errorf("got %v, want ErrDuplicateFunction", err)
	}
}

func TestWithFunction_BuiltinConflict(t *testing.T) {
	t.Parallel()
	cfg := &compilerConfig{}
	opt := WithFunction("has", func(params ...any) (any, error) { return nil, nil })
	err := opt(cfg)
	if !errors.Is(err, ErrBuiltinConflict) {
		t.Errorf("got %v, want ErrBuiltinConflict", err)
	}
}

func TestReplaceFunction_Valid(t *testing.T) {
	t.Parallel()
	cfg := &compilerConfig{}
	opt := ReplaceFunction("has", func(params ...any) (any, error) { return nil, nil })
	if err := opt(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cfg.replacedFuncs["has"]; !ok {
		t.Error("function not registered as replacement")
	}
}

func TestReplaceFunction_NotBuiltin(t *testing.T) {
	t.Parallel()
	cfg := &compilerConfig{}
	opt := ReplaceFunction("notabuiltin", func(params ...any) (any, error) { return nil, nil })
	err := opt(cfg)
	if !errors.Is(err, ErrNotBuiltin) {
		t.Errorf("got %v, want ErrNotBuiltin", err)
	}
}
