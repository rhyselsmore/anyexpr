package dispatch

import (
	"context"
	"testing"

	"github.com/rhyselsmore/anyexpr"
	rules "github.com/rhyselsmore/anyexpr/rules"
	"github.com/rhyselsmore/anyexpr/rules/action"
)

type benchEnv struct {
	Name   string
	Active bool
}

type benchActions[E any] struct {
	Label  rules.Action[string, E]       `rule:"label,multi"`
	Move   rules.Action[string, E]       `rule:"move"`
	Delete rules.Action[action.NoArgs, E] `rule:"delete,terminal"`
}

func benchDispatchSetup(b *testing.B) (*Plan[benchEnv, benchActions[benchEnv]], *rules.Evaluation[benchEnv, benchActions[benchEnv]]) {
	b.Helper()

	actions, _ := rules.DefineActions[benchEnv, benchActions[benchEnv]]()
	compiler, _ := anyexpr.NewCompiler[benchEnv]()
	prog, _ := rules.Compile(compiler, actions, []rules.Definition{
		{
			Name: "r1",
			When: `Active`,
			Then: []rules.ActionEntry{
				{Name: "label", Value: "a"},
				{Name: "label", Value: "b"},
				{Name: "move", Value: "archive"},
			},
		},
	})
	ev, _ := rules.NewEvaluator(prog)
	eval, _ := ev.Run(context.Background(), benchEnv{Name: "test", Active: true})

	noop := func(ctx context.Context, eval *rules.Evaluation[benchEnv, benchActions[benchEnv]]) error {
		return nil
	}

	d, _ := New(
		Handle("handler-a", noop),
		Handle("handler-b", noop),
		Handle("handler-c", noop),
	)

	plan, _ := d.Plan(
		Run[benchEnv, benchActions[benchEnv]]("handler-a",
			When[benchEnv, benchActions[benchEnv]](`Result.Label.Triggered`),
		),
		Run[benchEnv, benchActions[benchEnv]]("handler-b",
			When[benchEnv, benchActions[benchEnv]](`Result.Move.Triggered`),
		),
		Run[benchEnv, benchActions[benchEnv]]("handler-c"),
	)

	return plan, eval
}

func BenchmarkPlan_Execute(b *testing.B) {
	plan, eval := benchDispatchSetup(b)
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_ = plan.Execute(ctx, eval)
	}
}

func BenchmarkPlan_Execute_WithGate(b *testing.B) {
	actions, _ := rules.DefineActions[benchEnv, benchActions[benchEnv]]()
	compiler, _ := anyexpr.NewCompiler[benchEnv]()
	prog, _ := rules.Compile(compiler, actions, []rules.Definition{
		{Name: "r1", When: `Active`, Then: []rules.ActionEntry{{Name: "label", Value: "a"}}},
	})
	ev, _ := rules.NewEvaluator(prog)
	eval, _ := ev.Run(context.Background(), benchEnv{Active: true})

	noop := func(ctx context.Context, eval *rules.Evaluation[benchEnv, benchActions[benchEnv]]) error {
		return nil
	}
	d, _ := New(Handle("h", noop))
	plan, _ := d.Plan(
		Gate[benchEnv, benchActions[benchEnv]](`len(Matched) > 0`),
		Run[benchEnv, benchActions[benchEnv]]("h"),
	)

	ctx := context.Background()
	b.ResetTimer()
	for b.Loop() {
		_ = plan.Execute(ctx, eval)
	}
}

func BenchmarkPlan_Execute_Concurrent(b *testing.B) {
	plan, eval := benchDispatchSetup(b)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = plan.Execute(ctx, eval)
		}
	})
}
