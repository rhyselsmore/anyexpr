package rules

import (
	"errors"
	"testing"

	"github.com/rhyselsmore/anyexpr"
	"github.com/rhyselsmore/anyexpr/rules/action"
)

// --- Check ---

func TestCheck_Valid(t *testing.T) {
	t.Parallel()
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	err := Check(compiler, `has(Name, "alice")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheck_Invalid(t *testing.T) {
	t.Parallel()
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	err := Check(compiler, `invalid!!!`)
	if !errors.Is(err, ErrCompile) {
		t.Errorf("got %v, want ErrCompile", err)
	}
}

func TestCheck_UnknownField(t *testing.T) {
	t.Parallel()
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	err := Check(compiler, `has(NoSuchField, "x")`)
	if !errors.Is(err, ErrCompile) {
		t.Errorf("got %v, want ErrCompile", err)
	}
}

// --- TestRule ---

func TestTestRule_Match(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	eval, err := TestRule(compiler, actions, Definition{
		Name: "greet",
		When: `has(Name, "alice")`,
		Then: []ActionEntry{{Name: "label", Value: "friend"}},
	}, testEnv{Name: "alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !eval.Result.Label.Triggered {
		t.Error("Label should be triggered")
	}
	if eval.Result.Label.Value != "friend" {
		t.Errorf("Label.Value = %q, want friend", eval.Result.Label.Value)
	}
	if !eval.Traced {
		t.Error("TestRule should enable tracing")
	}
}

func TestTestRule_NoMatch(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	eval, err := TestRule(compiler, actions, Definition{
		Name: "greet",
		When: `has(Name, "alice")`,
		Then: []ActionEntry{{Name: "label", Value: "friend"}},
	}, testEnv{Name: "bob"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eval.Result.Label.Triggered {
		t.Error("Label should not be triggered")
	}
}

func TestTestRule_BadExpression(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	_, err := TestRule(compiler, actions, Definition{
		Name: "bad",
		When: `invalid!!!`,
	}, testEnv{})
	if !errors.Is(err, ErrCompile) {
		t.Errorf("got %v, want ErrCompile", err)
	}
}

// --- Assertion ---

func TestAssertion_Pass(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	eval, _ := TestRule(compiler, actions, Definition{
		Name: "r1",
		When: `Active`,
		Then: []ActionEntry{{Name: "label", Value: "billing"}},
	}, testEnv{Active: true})

	assertion, err := NewAssertion[testActions[testEnv]](`Label.Triggered`)
	if err != nil {
		t.Fatalf("NewAssertion: %v", err)
	}
	if err := assertion.AssertResult(eval.Result); err != nil {
		t.Errorf("assertion should pass: %v", err)
	}
}

func TestAssertion_Fail(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	eval, _ := TestRule(compiler, actions, Definition{
		Name: "r1",
		When: `Active`,
		Then: []ActionEntry{{Name: "label", Value: "billing"}},
	}, testEnv{Active: true})

	assertion, err := NewAssertion[testActions[testEnv]](`!Label.Triggered`)
	if err != nil {
		t.Fatalf("NewAssertion: %v", err)
	}
	if err := assertion.AssertResult(eval.Result); !errors.Is(err, ErrAssertFailed) {
		t.Errorf("got %v, want ErrAssertFailed", err)
	}
}

func TestAssertion_CompileError(t *testing.T) {
	t.Parallel()
	_, err := NewAssertion[testActions[testEnv]](`invalid!!!`)
	if !errors.Is(err, ErrAssert) {
		t.Errorf("got %v, want ErrAssert", err)
	}
}

func TestAssertion_ValueCheck(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	eval, _ := TestRule(compiler, actions, Definition{
		Name: "r1",
		When: `Active`,
		Then: []ActionEntry{
			{Name: "label", Value: "billing"},
			{Name: "priority", Value: 5},
		},
	}, testEnv{Active: true})

	assertion, err := NewAssertion[testActions[testEnv]](`Label.Value == "billing" && Priority.Value == 5`)
	if err != nil {
		t.Fatalf("NewAssertion: %v", err)
	}
	if err := assertion.AssertResult(eval.Result); err != nil {
		t.Errorf("assertion should pass: %v", err)
	}
}

// --- RunTestCase ---

func TestRunTestCase_Pass(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	result := RunTestCase(compiler, actions, TestCase[testEnv, testActions[testEnv]]{
		Name: "invoice labelling",
		Rule: Definition{
			Name: "invoices",
			When: `has(Name, "invoice")`,
			Then: []ActionEntry{
				{Name: "label", Value: "billing"},
				{Name: "read", Value: true},
			},
		},
		Env: testEnv{Name: "invoice-123"},
		Assertions: []string{
			`Label.Triggered`,
			`Label.Value == "billing"`,
			`Read.Value == true`,
			`!Delete.Triggered`,
		},
	})

	if !result.Passed {
		t.Errorf("expected pass, failures: %v, error: %v", result.Failures, result.Error)
	}
}

func TestRunTestCase_Fail(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	result := RunTestCase(compiler, actions, TestCase[testEnv, testActions[testEnv]]{
		Name: "wrong expectation",
		Rule: Definition{
			Name: "r1",
			When: `Active`,
			Then: []ActionEntry{{Name: "label", Value: "actual"}},
		},
		Env: testEnv{Active: true},
		Assertions: []string{
			`Label.Value == "expected"`, // will fail
		},
	})

	if result.Passed {
		t.Error("expected failure")
	}
	if len(result.Failures) != 1 {
		t.Errorf("got %d failures, want 1", len(result.Failures))
	}
}

func TestRunTestCase_CompileError(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	result := RunTestCase(compiler, actions, TestCase[testEnv, testActions[testEnv]]{
		Name: "bad rule",
		Rule: Definition{
			Name: "bad",
			When: `invalid!!!`,
		},
		Env: testEnv{},
	})

	if result.Passed {
		t.Error("expected failure")
	}
	if result.Error == nil {
		t.Error("expected error")
	}
}

func TestRunTestCase_NoMatch(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	result := RunTestCase(compiler, actions, TestCase[testEnv, testActions[testEnv]]{
		Name: "no match expected",
		Rule: Definition{
			Name: "r1",
			When: `has(Name, "alice")`,
			Then: []ActionEntry{{Name: "label", Value: "friend"}},
		},
		Env: testEnv{Name: "bob"},
		Assertions: []string{
			`!Label.Triggered`,
		},
	})

	if !result.Passed {
		t.Errorf("expected pass, failures: %v", result.Failures)
	}
}

// --- NoArgs in assertions ---

func TestRunTestCase_NoArgs(t *testing.T) {
	t.Parallel()
	actions, compiler := compileSetup(t)
	result := RunTestCase(compiler, actions, TestCase[testEnv, testActions[testEnv]]{
		Name: "delete fires",
		Rule: Definition{
			Name: "cleanup",
			When: `Active`,
			Then: []ActionEntry{{Name: "delete", Value: action.NoArgs{}}},
		},
		Env: testEnv{Active: true},
		Assertions: []string{
			`Delete.Triggered`,
		},
	})

	if !result.Passed {
		t.Errorf("expected pass, failures: %v, error: %v", result.Failures, result.Error)
	}
}
