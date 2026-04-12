package anyexpr

import "testing"

func TestNewSource_NameAndExpr(t *testing.T) {
	t.Parallel()
	s := NewSource("my-rule", `has(Name, "hello")`)
	if s.Name() != "my-rule" {
		t.Errorf("Name() = %q, want %q", s.Name(), "my-rule")
	}
	if s.Expr() != `has(Name, "hello")` {
		t.Errorf("Expr() = %q, want %q", s.Expr(), `has(Name, "hello")`)
	}
}

func TestNewSource_EmptyName(t *testing.T) {
	t.Parallel()
	s := NewSource("", "true")
	if s.Name() != "" {
		t.Errorf("Name() = %q, want empty", s.Name())
	}
}

func TestNewSource_EmptyExpr(t *testing.T) {
	t.Parallel()
	s := NewSource("rule", "")
	if s.Expr() != "" {
		t.Errorf("Expr() = %q, want empty", s.Expr())
	}
}

func TestNewSource_WithOpts(t *testing.T) {
	t.Parallel()
	// Opts are accepted without error (future-proofing).
	noop := func(*sourceConfig) {}
	s := NewSource("rule", "true", noop)
	if s.Name() != "rule" {
		t.Errorf("Name() = %q, want %q", s.Name(), "rule")
	}
}
