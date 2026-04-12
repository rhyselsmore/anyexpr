package rules_test

import (
	"context"
	"fmt"
	"log"

	"github.com/rhyselsmore/anyexpr"
	"github.com/rhyselsmore/anyexpr/rules"
)

type Email struct {
	From    string
	Subject string
	Body    string
	Amount  float64
}

type EmailActions[E any] struct {
	Label    rules.Action[string, E]       `rule:"label,multi"`
	Move     rules.Action[string, E]       `rule:"move"`
	Read     rules.Action[bool, E]         `rule:"read"`
	Priority rules.Action[int, E]          `rule:"priority"`
	Score    rules.Action[float64, E]      `rule:"score"`
	Delete   rules.Action[rules.NoArgs, E] `rule:"delete,terminal"`
}

func Example() {
	// Define actions.
	actions, err := rules.DefineActions[EmailActions[Email], Email]()
	if err != nil {
		log.Fatal(err)
	}

	// Build compiler.
	compiler, err := anyexpr.NewCompiler[Email]()
	if err != nil {
		log.Fatal(err)
	}

	// Compile rules.
	rs, err := rules.Compile(actions, compiler, []rules.Definition{
		{
			Name: "invoices",
			Tags: []string{"billing"},
			When: `has(Subject, "invoice")`,
			Then: []rules.ActionEntry{
				{Name: "label", Value: "billing"},
				{Name: "label", Value: "invoice"},
				{Name: "move", Value: "billing/invoices"},
				{Name: "read", Value: true},
				{Name: "priority", Value: 3},
				{Name: "score", Value: 0.95},
			},
		},
		{
			Name: "large",
			Tags: []string{"alerts"},
			When: `Amount > 1000`,
			Then: []rules.ActionEntry{
				{Name: "label", Value: "high-value"},
				{Name: "priority", Value: 5},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Build evaluator and run.
	evaluator, err := rules.NewEvaluator(actions, rs)
	if err != nil {
		log.Fatal(err)
	}

	eval, err := evaluator.Run(context.Background(), Email{
		From:    "billing@stripe.com",
		Subject: "Your January Invoice",
		Amount:  1500,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Typed accessors.
	fmt.Println("Labels:", eval.Actions.Label.Values())
	fmt.Println("Move:", must(eval.Actions.Move.Value()))
	fmt.Println("Read:", must(eval.Actions.Read.Value()))
	fmt.Println("Priority:", must(eval.Actions.Priority.Value()))
	fmt.Printf("Score: %.2f\n", must(eval.Actions.Score.Value()))
	fmt.Println("Matched:", len(eval.Matched), "rules")

	// Provenance.
	fmt.Println("Label rules:", eval.Actions.Label.Rules())
	fmt.Println("Labels from billing:", eval.Actions.Label.ByTag("billing"))

	// Output:
	// Labels: [billing invoice high-value]
	// Move: billing/invoices
	// Read: true
	// Priority: 5
	// Score: 0.95
	// Matched: 2 rules
	// Label rules: [invoices large]
	// Labels from billing: [billing invoice]
}

func must[T any](v T, _ bool) T { return v }
