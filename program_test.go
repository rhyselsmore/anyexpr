package anyexpr

import (
	"bytes"
	"errors"
	"sync"
	"testing"
)

// --- Match ---

func TestProgram_Match_True(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	prog, _ := c.Compile(NewSource("rule", `has(Name, "alice")`))
	got, err := prog.Match(testEnv{Name: "Alice Smith"})
	if err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Error("expected true")
	}
}

func TestProgram_Match_False(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	prog, _ := c.Compile(NewSource("rule", `has(Name, "alice")`))
	got, err := prog.Match(testEnv{Name: "Bob"})
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Error("expected false")
	}
}

func TestProgram_Match_NonBool(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	prog, _ := c.Compile(NewSource("rule", `lower(Name)`))
	_, err := prog.Match(testEnv{Name: "Alice"})
	if !errors.Is(err, ErrTypeMismatch) {
		t.Errorf("got %v, want ErrTypeMismatch", err)
	}
}

func TestProgram_Match_ConcurrentUse(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	prog, _ := c.Compile(NewSource("rule", `has(Name, "alice")`))

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := prog.Match(testEnv{Name: "Alice"})
			if err != nil {
				t.Errorf("match error: %v", err)
			}
			if !got {
				t.Error("expected true")
			}
		}()
	}
	wg.Wait()
}

func TestProgram_Match_WithTrace(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	prog, _ := c.Compile(NewSource("rule", `has(Name, "alice")`))
	var buf bytes.Buffer
	got, err := prog.Match(testEnv{Name: "Alice"}, WithMatchTrace(&buf))
	if err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Error("expected true")
	}
	// Trace is a no-op in v1, just verify the option is accepted.
}

// --- Eval ---

func TestProgram_Eval_Bool(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	prog, _ := c.Compile(NewSource("rule", `has(Name, "alice")`))
	out, err := prog.Eval(testEnv{Name: "Alice"})
	if err != nil {
		t.Fatal(err)
	}
	if out != true {
		t.Errorf("got %v, want true", out)
	}
}

func TestProgram_Eval_String(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	prog, _ := c.Compile(NewSource("rule", `lower(Name)`))
	out, err := prog.Eval(testEnv{Name: "ALICE"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "alice" {
		t.Errorf("got %v, want %q", out, "alice")
	}
}

func TestProgram_Eval_Int(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	prog, _ := c.Compile(NewSource("rule", `Age + 1`))
	out, err := prog.Eval(testEnv{Age: 30})
	if err != nil {
		t.Fatal(err)
	}
	if out != 31 {
		t.Errorf("got %v, want 31", out)
	}
}

func TestProgram_Eval_ConcurrentUse(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	prog, _ := c.Compile(NewSource("rule", `lower(Name)`))

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := prog.Eval(testEnv{Name: "ALICE"})
			if err != nil {
				t.Errorf("eval error: %v", err)
			}
			if out != "alice" {
				t.Errorf("got %v, want %q", out, "alice")
			}
		}()
	}
	wg.Wait()
}

// --- Metadata ---

func TestProgram_Name(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	prog, _ := c.Compile(NewSource("my-rule", `has(Name, "x")`))
	if prog.Name() != "my-rule" {
		t.Errorf("Name() = %q, want %q", prog.Name(), "my-rule")
	}
}

func TestProgram_Source(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()
	e := `has(Name, "x")`
	prog, _ := c.Compile(NewSource("rule", e))
	if prog.Source() != e {
		t.Errorf("Source() = %q, want %q", prog.Source(), e)
	}
}

// --- Integration: built-in functions through compilation and execution ---

func TestIntegration_HasCaseInsensitive(t *testing.T) {
	t.Parallel()

	type env struct {
		Subject string
	}

	compiler, err := NewCompiler[env]()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		expr string
		env  env
		want bool
	}{
		{"match lowercase", `has(Subject, "hello")`, env{"Hello World"}, true},
		{"match uppercase", `has(Subject, "HELLO")`, env{"Hello World"}, true},
		{"no match", `has(Subject, "goodbye")`, env{"Hello World"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			prog, err := compiler.Compile(NewSource(tt.name, tt.expr))
			if err != nil {
				t.Fatalf("compile: %v", err)
			}
			got, err := prog.Match(tt.env)
			if err != nil {
				t.Fatalf("match: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntegration_ExactMatchWithXhas(t *testing.T) {
	t.Parallel()
	type env struct{ Body string }
	c, _ := NewCompiler[env]()

	tests := []struct {
		name string
		expr string
		env  env
		want bool
	}{
		{"exact case match", `xhas(Body, "Hello")`, env{"Hello World"}, true},
		{"case mismatch", `xhas(Body, "hello")`, env{"Hello World"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			prog, _ := c.Compile(NewSource(tt.name, tt.expr))
			got, _ := prog.Match(tt.env)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntegration_RegexMatches(t *testing.T) {
	t.Parallel()
	type env struct{ Subject string }
	c, _ := NewCompiler[env]()

	prog, err := c.Compile(NewSource("re", `re(Subject, "^re:")`))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := prog.Match(env{"Re: hello"})
	if !got {
		t.Error("expected case-insensitive regex match")
	}
}

func TestIntegration_GlobPattern(t *testing.T) {
	t.Parallel()
	type env struct{ File string }
	c, _ := NewCompiler[env]()

	prog, _ := c.Compile(NewSource("glob", `glob(File, "*.txt")`))
	got, _ := prog.Match(env{"readme.txt"})
	if !got {
		t.Error("expected glob match")
	}
}

func TestIntegration_Extract(t *testing.T) {
	t.Parallel()
	type env struct{ Body string }
	c, _ := NewCompiler[env]()

	prog, _ := c.Compile(NewSource("extract", `extract(Body, "INV-\\d+")`))
	out, err := prog.Eval(env{"Please pay INV-42 by Friday"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "INV-42" {
		t.Errorf("got %v, want %q", out, "INV-42")
	}
}

func TestIntegration_Domain(t *testing.T) {
	t.Parallel()
	type env struct{ Email string }
	c, _ := NewCompiler[env]()

	prog, _ := c.Compile(NewSource("email_domain", `email_domain(Email)`))
	out, err := prog.Eval(env{"user@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "example.com" {
		t.Errorf("got %v, want %q", out, "example.com")
	}
}

func TestIntegration_ComplexExpression(t *testing.T) {
	t.Parallel()
	c, _ := NewCompiler[testEnv]()

	prog, err := c.Compile(NewSource("complex",
		`has(Name, "alice") && ends(Email, "example.com") && Age > 18`))
	if err != nil {
		t.Fatal(err)
	}

	got, err := prog.Match(testEnv{
		Name:  "Alice Smith",
		Email: "alice@example.com",
		Age:   25,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Error("expected true for complex expression")
	}
}
