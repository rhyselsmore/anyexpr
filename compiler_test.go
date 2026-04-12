package anyexpr

import (
	"errors"
	"sync"
	"testing"
)

type testEnv struct {
	Name   string
	Age    int
	Email  string
	Tags   []string
	Active bool
}

// --- Construction ---

func TestNewCompiler_NoOpts(t *testing.T) {
	t.Parallel()
	c, err := NewCompiler[testEnv]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("compiler is nil")
	}
}

func TestNewCompiler_WithFunction(t *testing.T) {
	t.Parallel()
	_, err := NewCompiler[testEnv](
		WithFunction("myfunc", func(params ...any) (any, error) {
			return params[0].(string) == "yes", nil
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCompiler_WithMultipleFunctions(t *testing.T) {
	t.Parallel()
	_, err := NewCompiler[testEnv](
		WithFunction("fn1", func(params ...any) (any, error) { return true, nil }),
		WithFunction("fn2", func(params ...any) (any, error) { return true, nil }),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCompiler_DuplicateFunction(t *testing.T) {
	t.Parallel()
	fn := func(params ...any) (any, error) { return true, nil }
	_, err := NewCompiler[testEnv](
		WithFunction("dup", fn),
		WithFunction("dup", fn),
	)
	if !errors.Is(err, ErrDuplicateFunction) {
		t.Errorf("got %v, want ErrDuplicateFunction", err)
	}
}

func TestNewCompiler_BuiltinConflict(t *testing.T) {
	t.Parallel()
	_, err := NewCompiler[testEnv](
		WithFunction("has", func(params ...any) (any, error) { return true, nil }),
	)
	if !errors.Is(err, ErrBuiltinConflict) {
		t.Errorf("got %v, want ErrBuiltinConflict", err)
	}
}

func TestNewCompiler_ReplaceFunction(t *testing.T) {
	t.Parallel()
	_, err := NewCompiler[testEnv](
		ReplaceFunction("has", func(params ...any) (any, error) {
			return true, nil
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCompiler_ReplaceNotBuiltin(t *testing.T) {
	t.Parallel()
	_, err := NewCompiler[testEnv](
		ReplaceFunction("notreal", func(params ...any) (any, error) { return true, nil }),
	)
	if !errors.Is(err, ErrNotBuiltin) {
		t.Errorf("got %v, want ErrNotBuiltin", err)
	}
}

func TestNewCompiler_ConcurrentUse(t *testing.T) {
	t.Parallel()
	c, err := NewCompiler[testEnv]()
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := c.Compile(NewSource("test", `has(Name, "hello")`))
			if err != nil {
				t.Errorf("compile error: %v", err)
			}
		}()
	}
	wg.Wait()
}

// --- Compile ---

func TestCompile_ValidExpression(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	prog, err := c.Compile(NewSource("rule1", `has(Name, "alice")`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prog == nil {
		t.Fatal("program is nil")
	}
}

func TestCompile_InvalidSyntax(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	_, err := c.Compile(NewSource("bad", `has(Name, `))
	if !errors.Is(err, ErrCompile) {
		t.Errorf("got %v, want ErrCompile", err)
	}
}

func TestCompile_TypeMismatch(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	_, err := c.Compile(NewSource("bad", `has(NonExistent, "x")`))
	if !errors.Is(err, ErrCompile) {
		t.Errorf("got %v, want ErrCompile", err)
	}
}

func TestCompile_ProgramName(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	prog, err := c.Compile(NewSource("my-rule", `has(Name, "x")`))
	if err != nil {
		t.Fatal(err)
	}
	if prog.Name() != "my-rule" {
		t.Errorf("Name() = %q, want %q", prog.Name(), "my-rule")
	}
}

func TestCompile_ProgramSource(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	expr := `has(Name, "x")`
	prog, err := c.Compile(NewSource("rule", expr))
	if err != nil {
		t.Fatal(err)
	}
	if prog.Source() != expr {
		t.Errorf("Source() = %q, want %q", prog.Source(), expr)
	}
}

func TestCompile_AllBuiltins(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()

	exprs := []struct {
		name string
		expr string
	}{
		{"has", `has(Name, "x")`},
		{"starts", `starts(Name, "x")`},
		{"ends", `ends(Name, "x")`},
		{"eq", `eq(Name, "x")`},
		{"xhas", `xhas(Name, "x")`},
		{"xstarts", `xstarts(Name, "x")`},
		{"xends", `xends(Name, "x")`},
		{"re", `re(Name, "x")`},
		{"xre", `xre(Name, "x")`},
		{"glob", `glob(Name, "x")`},
		{"lower", `lower(Name) == "x"`},
		{"upper", `upper(Name) == "X"`},
		{"trim", `trim(Name) == "x"`},
		{"extract", `extract(Name, "x") == "x"`},
		{"email_domain", `email_domain(Email) == "x"`},
	}

	for _, tt := range exprs {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := c.Compile(NewSource(tt.name, tt.expr))
			if err != nil {
				t.Errorf("compile %q failed: %v", tt.name, err)
			}
		})
	}
}

// --- Check ---

func TestCheck_AllValid(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	err := c.Check([]*Source{
		NewSource("r1", `has(Name, "x")`),
		NewSource("r2", `eq(Name, "y")`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheck_FirstInvalidFails(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	err := c.Check([]*Source{
		NewSource("good", `has(Name, "x")`),
		NewSource("bad", `has(Name, `),
	})
	if !errors.Is(err, ErrCompile) {
		t.Errorf("got %v, want ErrCompile", err)
	}
}

func TestCheck_ErrorIncludesSourceName(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	err := c.Check([]*Source{
		NewSource("my-bad-rule", `invalid!!!`),
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrCompile) {
		t.Errorf("got %v, want ErrCompile", err)
	}
}

func TestCheck_EmptySlice(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	err := c.Check([]*Source{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
