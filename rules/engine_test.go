package rules

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/rhyselsmore/anyexpr"
)

type testVars struct {
	Log []string
}

func setupEngine(t *testing.T, defs []Definition, regOpts ...RegistryOpt) *Engine[testEnv, testVars] {
	t.Helper()
	defaults := []RegistryOpt{
		WithAction("tag", Multi, StringVal, false),
		WithAction("category", Single, StringVal, false),
		WithAction("flag", Single, BoolValue, false),
		WithAction("delete", Single, NoValue, true),
		WithAction("expr-tag", Multi, StringExpr, false),
	}
	reg, err := NewRegistry(append(defaults, regOpts...)...)
	if err != nil {
		t.Fatal(err)
	}
	compiler, err := anyexpr.NewCompiler[testEnv]()
	if err != nil {
		t.Fatal(err)
	}
	rs, err := Compile(reg, compiler, defs)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewEngine[testEnv, testVars](reg, rs)
	if err != nil {
		t.Fatal(err)
	}
	return engine
}

// --- Basic evaluation ---

func TestEngine_Run_SingleMatch(t *testing.T) {
	t.Parallel()
	e := setupEngine(t, []Definition{
		{Name: "r1", When: `has(Name, "alice")`, Then: []ActionEntry{{Name: "tag", Value: "vip"}}},
	})
	result, err := e.Run(context.Background(), testEnv{Name: "Alice"}, testVars{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Matched) != 1 {
		t.Fatalf("got %d matched, want 1", len(result.Matched))
	}
	if result.Matched[0].Name != "r1" {
		t.Error("wrong rule matched")
	}
}

func TestEngine_Run_NoMatch(t *testing.T) {
	t.Parallel()
	e := setupEngine(t, []Definition{
		{Name: "r1", When: `has(Name, "alice")`, Then: []ActionEntry{{Name: "tag", Value: "vip"}}},
	})
	result, err := e.Run(context.Background(), testEnv{Name: "Bob"}, testVars{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Matched) != 0 {
		t.Error("expected no matches")
	}
	if result.Actions.ByName == nil {
		t.Error("ByName should be initialised")
	}
	if result.Actions.Flags == nil {
		t.Error("Flags should be initialised")
	}
}

func TestEngine_Run_MultipleMatches(t *testing.T) {
	t.Parallel()
	e := setupEngine(t, []Definition{
		{Name: "r1", When: `has(Name, "alice")`, Then: []ActionEntry{{Name: "tag", Value: "a"}}},
		{Name: "r2", When: `has(Name, "alice")`, Then: []ActionEntry{{Name: "tag", Value: "b"}}},
	})
	result, err := e.Run(context.Background(), testEnv{Name: "Alice"}, testVars{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Matched) != 2 {
		t.Fatalf("got %d matched, want 2", len(result.Matched))
	}
	tags := result.Actions.ByName["tag"]
	if len(tags) != 2 {
		t.Errorf("got %d tags, want 2: %v", len(tags), tags)
	}
}

func TestEngine_Run_RuleOrder(t *testing.T) {
	t.Parallel()
	e := setupEngine(t, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "category", Value: "first"}}},
		{Name: "r2", When: "true", Then: []ActionEntry{{Name: "category", Value: "second"}}},
	})
	result, _ := e.Run(context.Background(), testEnv{}, testVars{})
	vals := result.Actions.ByName["category"]
	if len(vals) != 1 || vals[0] != "second" {
		t.Errorf("single should be last-wins, got %v", vals)
	}
}

// --- Stop and terminal ---

func TestEngine_Run_StopHaltsEvaluation(t *testing.T) {
	t.Parallel()
	e := setupEngine(t, []Definition{
		{Name: "r1", Stop: true, When: "true", Then: []ActionEntry{{Name: "tag", Value: "a"}}},
		{Name: "r2", When: "true", Then: []ActionEntry{{Name: "tag", Value: "b"}}},
	})
	result, _ := e.Run(context.Background(), testEnv{}, testVars{})
	if len(result.Matched) != 1 {
		t.Errorf("got %d matched, want 1 (stop should halt)", len(result.Matched))
	}
	if !result.Stopped {
		t.Error("expected Stopped true")
	}
	if result.StoppedBy != "r1" {
		t.Errorf("StoppedBy = %q, want r1", result.StoppedBy)
	}
}

func TestEngine_Run_TerminalImpliesStop(t *testing.T) {
	t.Parallel()
	e := setupEngine(t, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "delete"}}},
		{Name: "r2", When: "true", Then: []ActionEntry{{Name: "tag", Value: "b"}}},
	})
	result, _ := e.Run(context.Background(), testEnv{}, testVars{})
	if len(result.Matched) != 1 {
		t.Errorf("got %d matched, want 1", len(result.Matched))
	}
	if !result.Stopped {
		t.Error("expected Stopped")
	}
	if !result.Actions.Terminal {
		t.Error("expected Terminal")
	}
}

func TestEngine_Run_ActionsBeforeTerminalPreserved(t *testing.T) {
	t.Parallel()
	e := setupEngine(t, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "keep"}}},
		{Name: "r2", When: "true", Then: []ActionEntry{{Name: "delete"}}},
	})
	result, _ := e.Run(context.Background(), testEnv{}, testVars{})
	tags := result.Actions.ByName["tag"]
	if len(tags) != 1 || tags[0] != "keep" {
		t.Errorf("actions before terminal should be preserved, got %v", tags)
	}
}

// --- Dynamic values ---

func TestEngine_Run_DynamicStringExpr(t *testing.T) {
	t.Parallel()
	e := setupEngine(t, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "expr-tag", Value: `lower(Name)`}}},
	})
	result, _ := e.Run(context.Background(), testEnv{Name: "ALICE"}, testVars{})
	tags := result.Actions.ByName["expr-tag"]
	if len(tags) != 1 || tags[0] != "alice" {
		t.Errorf("got %v, want [alice]", tags)
	}
}

func TestEngine_Run_EmptyDynamicValueSkipped(t *testing.T) {
	t.Parallel()
	e := setupEngine(t, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "expr-tag", Value: `trim(Name)`}}},
	})
	result, _ := e.Run(context.Background(), testEnv{Name: "   "}, testVars{})
	tags := result.Actions.ByName["expr-tag"]
	if len(tags) != 0 {
		t.Errorf("empty dynamic values should be skipped, got %v", tags)
	}
}

// --- Selectors ---

func TestEngine_Run_OnlyTags(t *testing.T) {
	t.Parallel()
	reg := testRegistry(t)
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	rs, _ := Compile(reg, compiler, []Definition{
		{Name: "r1", Tags: []string{"urgent"}, When: "true", Then: []ActionEntry{{Name: "tag", Value: "a"}}},
		{Name: "r2", Tags: []string{"normal"}, When: "true", Then: []ActionEntry{{Name: "tag", Value: "b"}}},
	})
	engine, _ := NewEngine[testEnv, testVars](reg, rs, WithTags("urgent"))
	result, _ := engine.Run(context.Background(), testEnv{}, testVars{})
	if len(result.Matched) != 1 || result.Matched[0].Name != "r1" {
		t.Errorf("expected only r1, got %v", result.Matched)
	}
}

func TestEngine_Run_OnlyNames(t *testing.T) {
	t.Parallel()
	reg := testRegistry(t)
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	rs, _ := Compile(reg, compiler, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "a"}}},
		{Name: "r2", When: "true", Then: []ActionEntry{{Name: "tag", Value: "b"}}},
	})
	engine, _ := NewEngine[testEnv, testVars](reg, rs, WithNames("r2"))
	result, _ := engine.Run(context.Background(), testEnv{}, testVars{})
	if len(result.Matched) != 1 || result.Matched[0].Name != "r2" {
		t.Errorf("expected only r2, got %v", result.Matched)
	}
}

func TestEngine_Run_ExcludeNames(t *testing.T) {
	t.Parallel()
	e := setupEngine(t, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "a"}}},
		{Name: "r2", When: "true", Then: []ActionEntry{{Name: "tag", Value: "b"}}},
	})
	result, _ := e.Run(context.Background(), testEnv{}, testVars{}, ExcludeNames("r1"))
	if len(result.Matched) != 1 || result.Matched[0].Name != "r2" {
		t.Errorf("expected only r2")
	}
}

func TestEngine_Run_ExcludeTags(t *testing.T) {
	t.Parallel()
	reg := testRegistry(t)
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	rs, _ := Compile(reg, compiler, []Definition{
		{Name: "r1", Tags: []string{"skip"}, When: "true", Then: []ActionEntry{{Name: "tag", Value: "a"}}},
		{Name: "r2", When: "true", Then: []ActionEntry{{Name: "tag", Value: "b"}}},
	})
	engine, _ := NewEngine[testEnv, testVars](reg, rs)
	result, _ := engine.Run(context.Background(), testEnv{}, testVars{}, ExcludeTags("skip"))
	if len(result.Matched) != 1 || result.Matched[0].Name != "r2" {
		t.Errorf("expected only r2")
	}
}

func TestEngine_Run_PerExecutionSelector(t *testing.T) {
	t.Parallel()
	e := setupEngine(t, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "a"}}},
		{Name: "r2", When: "true", Then: []ActionEntry{{Name: "tag", Value: "b"}}},
	})
	result, _ := e.Run(context.Background(), testEnv{}, testVars{}, OnlyNames("r1"))
	if len(result.Matched) != 1 || result.Matched[0].Name != "r1" {
		t.Error("per-execution selector should filter")
	}
}

// --- Disabled rules ---

func TestEngine_Run_DisabledRuleSkipped(t *testing.T) {
	t.Parallel()
	f := false
	e := setupEngine(t, []Definition{
		{Name: "r1", Enabled: &f, When: "true", Then: []ActionEntry{{Name: "tag", Value: "a"}}},
		{Name: "r2", When: "true", Then: []ActionEntry{{Name: "tag", Value: "b"}}},
	})
	result, _ := e.Run(context.Background(), testEnv{}, testVars{})
	if len(result.Matched) != 1 || result.Matched[0].Name != "r2" {
		t.Error("disabled rule should be skipped")
	}
}

// --- Handlers ---

func TestEngine_Run_HandlerCalled(t *testing.T) {
	t.Parallel()
	called := false
	handler := func(ctx *Context[testEnv, testVars]) error {
		called = true
		return nil
	}
	reg, _ := NewRegistry(
		WithAction("tag", Multi, StringVal, false),
		WithHandler("test-handler", handler, Multi, false),
	)
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	rs, _ := Compile(reg, compiler, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{
			{Name: "tag", Value: "a"},
			{Name: "test-handler"},
		}},
	})
	engine, _ := NewEngine[testEnv, testVars](reg, rs)
	engine.Run(context.Background(), testEnv{Name: "Alice"}, testVars{})
	if !called {
		t.Error("handler was not called")
	}
}

func TestEngine_Run_HandlerReceivesVars(t *testing.T) {
	t.Parallel()
	var receivedVars testVars
	handler := func(ctx *Context[testEnv, testVars]) error {
		receivedVars = ctx.Vars
		return nil
	}
	reg, _ := NewRegistry(WithHandler("h", handler, Multi, false))
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	rs, _ := Compile(reg, compiler, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "h"}}},
	})
	engine, _ := NewEngine[testEnv, testVars](reg, rs)

	vars := testVars{Log: []string{"hello"}}
	engine.Run(context.Background(), testEnv{}, vars)
	if len(receivedVars.Log) != 1 || receivedVars.Log[0] != "hello" {
		t.Errorf("handler did not receive vars, got %v", receivedVars)
	}
}

func TestEngine_Run_HandlerError(t *testing.T) {
	t.Parallel()
	handler := func(ctx *Context[testEnv, testVars]) error {
		return fmt.Errorf("handler failed")
	}
	reg, _ := NewRegistry(
		WithAction("tag", Multi, StringVal, false),
		WithHandler("h", handler, Multi, false),
	)
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	rs, _ := Compile(reg, compiler, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "a"}, {Name: "h"}}},
	})
	engine, _ := NewEngine[testEnv, testVars](reg, rs)

	result, err := engine.Run(context.Background(), testEnv{}, testVars{})
	if err == nil {
		t.Error("expected error from handler")
	}
	// Result should still be populated.
	if len(result.Matched) != 1 {
		t.Error("result should be populated despite handler error")
	}
}

func TestEngine_Run_MultipleHandlers(t *testing.T) {
	t.Parallel()
	var order []string
	h1 := func(ctx *Context[testEnv, testVars]) error {
		order = append(order, "h1")
		return nil
	}
	h2 := func(ctx *Context[testEnv, testVars]) error {
		order = append(order, "h2")
		return nil
	}
	reg, _ := NewRegistry(
		WithHandler("h1", h1, Multi, false),
		WithHandler("h2", h2, Multi, false),
	)
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	rs, _ := Compile(reg, compiler, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "h1"}, {Name: "h2"}}},
	})
	engine, _ := NewEngine[testEnv, testVars](reg, rs)
	engine.Run(context.Background(), testEnv{}, testVars{})
	if len(order) != 2 || order[0] != "h1" || order[1] != "h2" {
		t.Errorf("handlers called in wrong order: %v", order)
	}
}

func TestEngine_DryRun_HandlersNotCalled(t *testing.T) {
	t.Parallel()
	called := false
	handler := func(ctx *Context[testEnv, testVars]) error {
		called = true
		return nil
	}
	reg, _ := NewRegistry(WithHandler("h", handler, Multi, false))
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	rs, _ := Compile(reg, compiler, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "h"}}},
	})
	engine, _ := NewEngine[testEnv, testVars](reg, rs)
	result, _ := engine.DryRun(context.Background(), testEnv{}, testVars{})
	if called {
		t.Error("handler should not be called in dry run")
	}
	if len(result.Matched) != 1 {
		t.Error("dry run should still populate result")
	}
}

// --- Resolution ---

func TestEngine_Run_MultiDedup(t *testing.T) {
	t.Parallel()
	e := setupEngine(t, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "same"}}},
		{Name: "r2", When: "true", Then: []ActionEntry{{Name: "tag", Value: "same"}}},
	})
	result, _ := e.Run(context.Background(), testEnv{}, testVars{})
	tags := result.Actions.ByName["tag"]
	if len(tags) != 1 {
		t.Errorf("multi should dedup, got %v", tags)
	}
}

func TestEngine_Run_SingleLastWins(t *testing.T) {
	t.Parallel()
	e := setupEngine(t, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "category", Value: "first"}}},
		{Name: "r2", When: "true", Then: []ActionEntry{{Name: "category", Value: "last"}}},
	})
	result, _ := e.Run(context.Background(), testEnv{}, testVars{})
	vals := result.Actions.ByName["category"]
	if len(vals) != 1 || vals[0] != "last" {
		t.Errorf("single should be last-wins, got %v", vals)
	}
}

func TestEngine_Run_BoolLastWins(t *testing.T) {
	t.Parallel()
	e := setupEngine(t, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "flag", Value: "true"}}},
		{Name: "r2", When: "true", Then: []ActionEntry{{Name: "flag", Value: "false"}}},
	})
	result, _ := e.Run(context.Background(), testEnv{}, testVars{})
	if result.Actions.Flags["flag"] == nil || *result.Actions.Flags["flag"] != false {
		t.Error("bool should be last-wins (false)")
	}
}

// --- Type safety ---

func TestNewEngine_HandlerTypeMismatch(t *testing.T) {
	t.Parallel()
	reg, _ := NewRegistry(WithHandler("h", "not-a-function", Multi, false))
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	rs, _ := Compile(reg, compiler, []Definition{})
	_, err := NewEngine[testEnv, testVars](reg, rs)
	if !errors.Is(err, ErrHandlerType) {
		t.Errorf("got %v, want ErrHandlerType", err)
	}
}

func TestNewEngine_HandlerCorrectType(t *testing.T) {
	t.Parallel()
	handler := func(ctx *Context[testEnv, testVars]) error { return nil }
	reg, _ := NewRegistry(WithHandler("h", handler, Multi, false))
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	rs, _ := Compile(reg, compiler, []Definition{})
	_, err := NewEngine[testEnv, testVars](reg, rs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Context cancellation ---

func TestEngine_Run_ContextCancellation(t *testing.T) {
	t.Parallel()
	e := setupEngine(t, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "a"}}},
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	_, err := e.Run(ctx, testEnv{}, testVars{})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("got %v, want context.Canceled", err)
	}
}

// --- Concurrency ---

func TestEngine_Run_Concurrent(t *testing.T) {
	t.Parallel()
	e := setupEngine(t, []Definition{
		{Name: "r1", When: `has(Name, "alice")`, Then: []ActionEntry{{Name: "tag", Value: "vip"}}},
	})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := e.Run(context.Background(), testEnv{Name: "Alice"}, testVars{})
			if err != nil {
				t.Errorf("error: %v", err)
			}
			if len(result.Matched) != 1 {
				t.Errorf("got %d matched, want 1", len(result.Matched))
			}
		}()
	}
	wg.Wait()
}

// --- No vars ---

func TestEngine_Run_EmptyVars(t *testing.T) {
	t.Parallel()
	reg, _ := NewRegistry(WithAction("tag", Multi, StringVal, false))
	compiler, _ := anyexpr.NewCompiler[testEnv]()
	rs, _ := Compile(reg, compiler, []Definition{
		{Name: "r1", When: "true", Then: []ActionEntry{{Name: "tag", Value: "a"}}},
	})
	engine, err := NewEngine[testEnv, struct{}](reg, rs)
	if err != nil {
		t.Fatal(err)
	}
	result, err := engine.Run(context.Background(), testEnv{}, struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Matched) != 1 {
		t.Error("expected 1 match")
	}
}
