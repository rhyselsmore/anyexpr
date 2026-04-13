package rules

import (
	"context"
	"testing"

	"github.com/rhyselsmore/anyexpr"
	"github.com/rhyselsmore/anyexpr/rules/action"
)

// --- Bench types ---

type benchEnv struct {
	From    string
	Subject string
	Amount  float64
	Active  bool
}

type benchActions[E any] struct {
	Label    Action[string, E]       `rule:"label,multi"`
	Move     Action[string, E]       `rule:"move"`
	Read     Action[bool, E]         `rule:"read"`
	Priority Action[int, E]          `rule:"priority"`
	Score    Action[float64, E]      `rule:"score"`
	Delete   Action[action.NoArgs, E] `rule:"delete,terminal"`
}

var benchDefs = []Definition{
	{
		Name: "invoices",
		Tags: []string{"billing"},
		When: `has(Subject, "invoice")`,
		Then: []ActionEntry{
			{Name: "label", Value: "billing"},
			{Name: "label", Value: "invoice"},
			{Name: "move", Value: "billing/invoices"},
			{Name: "read", Value: true},
			{Name: "priority", Value: 3},
			{Name: "score", Value: 0.95},
		},
	},
	{
		Name: "receipts",
		Tags: []string{"billing"},
		When: `has(Subject, "receipt")`,
		Then: []ActionEntry{
			{Name: "label", Value: "billing"},
			{Name: "label", Value: "receipt"},
			{Name: "move", Value: "billing/receipts"},
			{Name: "priority", Value: 1},
		},
	},
	{
		Name: "large",
		Tags: []string{"alerts"},
		When: `Amount > 1000`,
		Then: []ActionEntry{
			{Name: "label", Value: "high-value"},
			{Name: "priority", Value: 5},
		},
	},
	{
		Name: "spam",
		Tags: []string{"cleanup"},
		When: `has(From, "noreply@junk.com")`,
		Then: []ActionEntry{
			{Name: "delete"},
		},
	},
}

var benchEnvMatch = benchEnv{
	From:    "billing@stripe.com",
	Subject: "Your January Invoice",
	Amount:  1500,
	Active:  true,
}

var benchEnvNoMatch = benchEnv{
	From:    "friend@example.com",
	Subject: "Hey how are you?",
	Amount:  0,
	Active:  false,
}

func benchSetup(b *testing.B) (*Evaluator[benchEnv, benchActions[benchEnv]], *Actions[benchEnv, benchActions[benchEnv]], *anyexpr.Compiler[benchEnv]) {
	b.Helper()
	actions, err := DefineActions[benchEnv, benchActions[benchEnv]]()
	if err != nil {
		b.Fatal(err)
	}
	compiler, err := anyexpr.NewCompiler[benchEnv]()
	if err != nil {
		b.Fatal(err)
	}
	prog, err := Compile(compiler, actions, benchDefs)
	if err != nil {
		b.Fatal(err)
	}
	ev, err := NewEvaluator(prog)
	if err != nil {
		b.Fatal(err)
	}
	return ev, actions, compiler
}

// --- DefineActions ---

func BenchmarkDefineActions(b *testing.B) {
	for b.Loop() {
		_, err := DefineActions[benchEnv, benchActions[benchEnv]]()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --- Compile ---

func BenchmarkCompile(b *testing.B) {
	actions, err := DefineActions[benchEnv, benchActions[benchEnv]]()
	if err != nil {
		b.Fatal(err)
	}
	compiler, err := anyexpr.NewCompiler[benchEnv]()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := Compile(compiler, actions, benchDefs)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --- Run: match ---

func BenchmarkRun_Match(b *testing.B) {
	ev, _, _ := benchSetup(b)
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, err := ev.Run(ctx, benchEnvMatch)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --- Run: no match ---

func BenchmarkRun_NoMatch(b *testing.B) {
	ev, _, _ := benchSetup(b)
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, err := ev.Run(ctx, benchEnvNoMatch)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --- Run: with trace ---

func BenchmarkRun_WithTrace(b *testing.B) {
	ev, _, _ := benchSetup(b)
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, err := ev.Run(ctx, benchEnvMatch, WithTrace(true))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --- Run: with selector ---

func BenchmarkRun_WithSelector(b *testing.B) {
	ev, _, _ := benchSetup(b)
	ctx := context.Background()
	sel := MustWithSelector(`"billing" in Tags`)

	b.ResetTimer()
	for b.Loop() {
		_, err := ev.Run(ctx, benchEnvMatch, sel)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --- Run: many rules ---

func BenchmarkRun_TenRules(b *testing.B) {
	actions, err := DefineActions[benchEnv, benchActions[benchEnv]]()
	if err != nil {
		b.Fatal(err)
	}
	compiler, err := anyexpr.NewCompiler[benchEnv]()
	if err != nil {
		b.Fatal(err)
	}

	defs := make([]Definition, 10)
	for i := range defs {
		defs[i] = Definition{
			Name: "rule-" + string(rune('a'+i)),
			When: `Active`,
			Then: []ActionEntry{{Name: "label", Value: "tag"}},
		}
	}

	prog, err := Compile(compiler, actions, defs)
	if err != nil {
		b.Fatal(err)
	}
	ev, err := NewEvaluator(prog)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	b.ResetTimer()
	for b.Loop() {
		_, err := ev.Run(ctx, benchEnvMatch)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --- Run: concurrent ---

func BenchmarkRun_Concurrent(b *testing.B) {
	ev, _, _ := benchSetup(b)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := ev.Run(ctx, benchEnvMatch)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// --- Describe ---

func BenchmarkDescribe(b *testing.B) {
	actions, err := DefineActions[benchEnv, benchActions[benchEnv]]()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		_ = actions.Describe()
	}
}
