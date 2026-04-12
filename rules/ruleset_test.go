package rules

import (
	"errors"
	"testing"

	"github.com/rhyselsmore/anyexpr"
)

type testEnv struct {
	Name   string
	Amount float64
	Tags   []string
	Active bool
}

func testRegistry(t *testing.T, opts ...RegistryOpt) *Registry {
	t.Helper()
	defaults := []RegistryOpt{
		WithAction("tag", Multi, StringVal, false),
		WithAction("category", Single, StringVal, false),
		WithAction("flag", Single, BoolValue, false),
		WithAction("delete", Single, NoValue, true),
		WithAction("expr-tag", Multi, StringExpr, false),
	}
	r, err := NewRegistry(append(defaults, opts...)...)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func testCompiler(t *testing.T) *anyexpr.Compiler[testEnv] {
	t.Helper()
	c, err := anyexpr.NewCompiler[testEnv]()
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// --- Compilation ---

func TestCompile_ValidRules(t *testing.T) {
	t.Parallel()
	rs, err := Compile(testRegistry(t), testCompiler(t), []Definition{
		{Name: "r1", When: `has(Name, "alice")`, Then: []ActionEntry{{Name: "tag", Value: "vip"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rs.Len() != 1 {
		t.Errorf("got %d rules, want 1", rs.Len())
	}
}

func TestCompile_DuplicateRuleNames(t *testing.T) {
	t.Parallel()
	_, err := Compile(testRegistry(t), testCompiler(t), []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "a"}}},
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "b"}}},
	})
	if !errors.Is(err, ErrDuplicateRule) {
		t.Errorf("got %v, want ErrDuplicateRule", err)
	}
}

func TestCompile_UnknownAction(t *testing.T) {
	t.Parallel()
	_, err := Compile(testRegistry(t), testCompiler(t), []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "nope", Value: "x"}}},
	})
	if !errors.Is(err, ErrUnknownAction) {
		t.Errorf("got %v, want ErrUnknownAction", err)
	}
}

func TestCompile_MultipleTerminals(t *testing.T) {
	t.Parallel()
	reg, _ := NewRegistry(
		WithAction("del1", Single, NoValue, true),
		WithAction("del2", Single, NoValue, true),
	)
	_, err := Compile(reg, testCompiler(t), []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "del1"}, {Name: "del2"}}},
	})
	if !errors.Is(err, ErrMultipleTerminals) {
		t.Errorf("got %v, want ErrMultipleTerminals", err)
	}
}

func TestCompile_SingleUseViolation(t *testing.T) {
	t.Parallel()
	_, err := Compile(testRegistry(t), testCompiler(t), []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{
			{Name: "category", Value: "a"},
			{Name: "category", Value: "b"},
		}},
	})
	if !errors.Is(err, ErrCardinalityViolation) {
		t.Errorf("got %v, want ErrCardinalityViolation", err)
	}
}

func TestCompile_BadWhenExpression(t *testing.T) {
	t.Parallel()
	_, err := Compile(testRegistry(t), testCompiler(t), []Definition{
		{Name: "r1", When: `has(Name, `, Then: []ActionEntry{{Name: "tag", Value: "x"}}},
	})
	if !errors.Is(err, ErrCompile) {
		t.Errorf("got %v, want ErrCompile", err)
	}
}

func TestCompile_BadValueExpression(t *testing.T) {
	t.Parallel()
	_, err := Compile(testRegistry(t), testCompiler(t), []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "expr-tag", Value: `has(Name, `}}},
	})
	if !errors.Is(err, ErrCompile) {
		t.Errorf("got %v, want ErrCompile", err)
	}
}

func TestCompile_BoolValueParsing(t *testing.T) {
	t.Parallel()
	for _, val := range []string{"true", "false"} {
		_, err := Compile(testRegistry(t), testCompiler(t), []Definition{
			{Name: "r-" + val, When: "true", Then: []ActionEntry{{Name: "flag", Value: val}}},
		})
		if err != nil {
			t.Errorf("bool %q: %v", val, err)
		}
	}
}

func TestCompile_BoolValueInvalid(t *testing.T) {
	t.Parallel()
	_, err := Compile(testRegistry(t), testCompiler(t), []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "flag", Value: "yes"}}},
	})
	if !errors.Is(err, ErrValueType) {
		t.Errorf("got %v, want ErrValueType", err)
	}
}

func TestCompile_NoValueWithValue(t *testing.T) {
	t.Parallel()
	_, err := Compile(testRegistry(t), testCompiler(t), []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "delete", Value: "oops"}}},
	})
	if !errors.Is(err, ErrValueType) {
		t.Errorf("got %v, want ErrValueType", err)
	}
}

func TestCompile_DisabledRule(t *testing.T) {
	t.Parallel()
	f := false
	_, err := Compile(testRegistry(t), testCompiler(t), []Definition{
		{Name: "r1", Enabled: &f, When: "true", Then: []ActionEntry{{Name: "tag", Value: "x"}}},
	})
	if err != nil {
		t.Fatalf("disabled rules should still compile: %v", err)
	}
}

func TestCompile_StopImpliedByTerminal(t *testing.T) {
	t.Parallel()
	rs, _ := Compile(testRegistry(t), testCompiler(t), []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "delete"}}},
	})
	if !rs.rules[0].stop {
		t.Error("terminal action should imply stop")
	}
}

func TestCompile_EmptyDefinitions(t *testing.T) {
	t.Parallel()
	rs, err := Compile(testRegistry(t), testCompiler(t), []Definition{})
	if err != nil {
		t.Fatal(err)
	}
	if rs.Len() != 0 {
		t.Errorf("got %d, want 0", rs.Len())
	}
}

// --- Properties ---

func TestRuleset_Names(t *testing.T) {
	t.Parallel()
	rs, _ := Compile(testRegistry(t), testCompiler(t), []Definition{
		{Name: "b", When: "true", Then: []ActionEntry{{Name: "tag", Value: "x"}}},
		{Name: "a", When: "true", Then: []ActionEntry{{Name: "tag", Value: "y"}}},
	})
	names := rs.Names()
	if len(names) != 2 || names[0] != "b" || names[1] != "a" {
		t.Errorf("got %v, want [b a] (definition order)", names)
	}
}

func TestRuleset_Tags(t *testing.T) {
	t.Parallel()
	rs, _ := Compile(testRegistry(t), testCompiler(t), []Definition{
		{Name: "r1", Tags: []string{"x", "y"}, When: "true", Then: []ActionEntry{{Name: "tag", Value: "a"}}},
		{Name: "r2", Tags: []string{"y", "z"}, When: "true", Then: []ActionEntry{{Name: "tag", Value: "b"}}},
	})
	tags := rs.Tags()
	if len(tags) != 3 {
		t.Errorf("got %v, want 3 unique tags", tags)
	}
}

func TestRuleset_Len(t *testing.T) {
	t.Parallel()
	rs, _ := Compile(testRegistry(t), testCompiler(t), []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "a"}}},
		{Name: "r2", When: "true", Then: []ActionEntry{{Name: "tag", Value: "b"}}},
	})
	if rs.Len() != 2 {
		t.Errorf("got %d, want 2", rs.Len())
	}
}

// --- Merge ---

func TestRuleset_Merge_NoCollision(t *testing.T) {
	t.Parallel()
	reg := testRegistry(t)
	c := testCompiler(t)
	a, _ := Compile(reg, c, []Definition{{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "a"}}}})
	b, _ := Compile(reg, c, []Definition{{Name: "r2", When: "true", Then: []ActionEntry{{Name: "tag", Value: "b"}}}})

	merged, err := a.Merge(b)
	if err != nil {
		t.Fatal(err)
	}
	if merged.Len() != 2 {
		t.Errorf("got %d, want 2", merged.Len())
	}
}

func TestRuleset_Merge_Collision(t *testing.T) {
	t.Parallel()
	reg := testRegistry(t)
	c := testCompiler(t)
	a, _ := Compile(reg, c, []Definition{{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "a"}}}})
	b, _ := Compile(reg, c, []Definition{{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "b"}}}})

	_, err := a.Merge(b)
	if !errors.Is(err, ErrNameCollision) {
		t.Errorf("got %v, want ErrNameCollision", err)
	}
}

func TestRuleset_Merge_AllowOverride(t *testing.T) {
	t.Parallel()
	reg := testRegistry(t)
	c := testCompiler(t)
	a, _ := Compile(reg, c, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "original"}}},
		{Name: "r2", When: "true", Then: []ActionEntry{{Name: "tag", Value: "keep"}}},
	})
	b, _ := Compile(reg, c, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "override"}}},
	})

	merged, err := a.Merge(b, AllowOverride)
	if err != nil {
		t.Fatal(err)
	}
	if merged.Len() != 2 {
		t.Errorf("got %d, want 2", merged.Len())
	}
	// r1 should be at position 0 (original position) with override value.
	if merged.Names()[0] != "r1" {
		t.Error("r1 should keep its position")
	}
}

func TestRuleset_Merge_NeitherModified(t *testing.T) {
	t.Parallel()
	reg := testRegistry(t)
	c := testCompiler(t)
	a, _ := Compile(reg, c, []Definition{{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "a"}}}})
	b, _ := Compile(reg, c, []Definition{{Name: "r2", When: "true", Then: []ActionEntry{{Name: "tag", Value: "b"}}}})

	a.Merge(b)
	if a.Len() != 1 {
		t.Error("a was modified")
	}
	if b.Len() != 1 {
		t.Error("b was modified")
	}
}
