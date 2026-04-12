package rules

import (
	"context"
	"testing"

	"github.com/rhyselsmore/anyexpr"
)

func benchRegistry(b *testing.B) *Registry {
	b.Helper()
	r, _ := NewRegistry(
		WithAction("tag", Multi, StringVal, false),
		WithAction("category", Single, StringVal, false),
		WithAction("flag", Single, BoolValue, false),
		WithAction("delete", Single, NoValue, true),
	)
	return r
}

func benchCompiler(b *testing.B) *anyexpr.Compiler[testEnv] {
	b.Helper()
	c, _ := anyexpr.NewCompiler[testEnv]()
	return c
}

func BenchmarkEngine_Run_SingleRule(b *testing.B) {
	reg := benchRegistry(b)
	rs, _ := Compile(reg, benchCompiler(b), []Definition{
		{Name: "r1", When: `has(Name, "alice")`, Then: []ActionEntry{{Name: "tag", Value: "vip"}}},
	})
	engine, _ := NewEngine[testEnv, struct{}](reg, rs)
	env := testEnv{Name: "Alice"}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Run(ctx, env, struct{}{})
	}
}

func BenchmarkEngine_Run_TenRules(b *testing.B) {
	reg := benchRegistry(b)
	var defs []Definition
	for i := 0; i < 10; i++ {
		defs = append(defs, Definition{
			Name: "r" + string(rune('0'+i)),
			When: `has(Name, "alice")`,
			Then: []ActionEntry{{Name: "tag", Value: "t" + string(rune('0'+i))}},
		})
	}
	rs, _ := Compile(reg, benchCompiler(b), defs)
	engine, _ := NewEngine[testEnv, struct{}](reg, rs)
	env := testEnv{Name: "Alice"}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Run(ctx, env, struct{}{})
	}
}

func BenchmarkEngine_DryRun(b *testing.B) {
	reg := benchRegistry(b)
	rs, _ := Compile(reg, benchCompiler(b), []Definition{
		{Name: "r1", When: `has(Name, "alice")`, Then: []ActionEntry{{Name: "tag", Value: "vip"}}},
	})
	engine, _ := NewEngine[testEnv, struct{}](reg, rs)
	env := testEnv{Name: "Alice"}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.DryRun(ctx, env, struct{}{})
	}
}

func BenchmarkCompile(b *testing.B) {
	reg := benchRegistry(b)
	compiler := benchCompiler(b)
	defs := []Definition{
		{Name: "r1", When: `has(Name, "alice")`, Then: []ActionEntry{{Name: "tag", Value: "vip"}}},
		{Name: "r2", When: `Amount > 100`, Then: []ActionEntry{{Name: "category", Value: "high"}}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Compile(reg, compiler, defs)
	}
}

func BenchmarkSet_Resolve(b *testing.B) {
	s := NewSet()
	for i := 0; i < 20; i++ {
		s.Add(Entry{
			Def:   Def{Name: "tag", Cardinality: Multi, Value: StringVal},
			Value: "val" + string(rune('a'+i)),
		})
	}
	s.Add(Entry{Def: Def{Name: "category", Cardinality: Single, Value: StringVal}, Value: "x"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Resolve()
	}
}
