package rules_test

import (
	"context"
	"fmt"
	"log"

	"github.com/rhyselsmore/anyexpr"
	rules "github.com/rhyselsmore/anyexpr/rules"
	"github.com/rhyselsmore/anyexpr/rules/action"
)

type Email struct {
	From    string
	Subject string
	Amount  float64
}

type EmailActions[E any] struct {
	Label    rules.Action[string, E]       `rule:"label,multi" description:"categorisation labels"`
	Move     rules.Action[string, E]       `rule:"move" description:"destination folder"`
	Read     rules.Action[bool, E]         `rule:"read"`
	Priority rules.Action[int, E]          `rule:"priority"`
	Delete   rules.Action[action.NoArgs, E] `rule:"delete,terminal"`
}

func Example() {
	// Define actions from struct tags.
	actions, err := rules.DefineActions[Email, EmailActions[Email]]()
	if err != nil {
		log.Fatal(err)
	}

	// Build the expression compiler.
	compiler, err := anyexpr.NewCompiler[Email]()
	if err != nil {
		log.Fatal(err)
	}

	// Compile rules — values are type-checked at compile time.
	prog, err := rules.Compile(compiler, actions, []rules.Definition{
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

	// Create evaluator and run.
	evaluator, err := rules.NewEvaluator(prog)
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

	// Read typed results directly from struct fields.
	fmt.Println("Labels:", eval.Result.Label.Values)
	fmt.Println("Move:", eval.Result.Move.Value)
	fmt.Println("Read:", eval.Result.Read.Value)
	fmt.Println("Priority:", eval.Result.Priority.Value)
	fmt.Println("Matched:", eval.Matched)

	// Output:
	// Labels: [billing invoice high-value]
	// Move: billing/invoices
	// Read: true
	// Priority: 5
	// Matched: [invoices large]
}

func Example_introspection() {
	actions, err := rules.DefineActions[Email, EmailActions[Email]]()
	if err != nil {
		log.Fatal(err)
	}

	for _, info := range actions.Describe() {
		line := fmt.Sprintf("%-10s %s", info.Name, info.ValueType)
		if info.Description != "" {
			line += " — " + info.Description
		}
		fmt.Println(line)
	}

	// Output:
	// label      string — categorisation labels
	// move       string — destination folder
	// read       bool
	// priority   int
	// delete     action.NoArgs
}

func Example_testCase() {
	actions, err := rules.DefineActions[Email, EmailActions[Email]]()
	if err != nil {
		log.Fatal(err)
	}
	compiler, err := anyexpr.NewCompiler[Email]()
	if err != nil {
		log.Fatal(err)
	}

	result := rules.RunTestCase(compiler, actions, rules.TestCase[Email, EmailActions[Email]]{
		Name: "invoice labelling",
		Rule: rules.Definition{
			Name: "invoices",
			When: `has(Subject, "invoice")`,
			Then: []rules.ActionEntry{
				{Name: "label", Value: "billing"},
				{Name: "read", Value: true},
			},
		},
		Env: Email{Subject: "Your Invoice"},
		Assertions: []string{
			`Label.Triggered`,
			`Label.Value == "billing"`,
			`Read.Value == true`,
			`!Delete.Triggered`,
		},
	})

	fmt.Println("Passed:", result.Passed)

	// Output:
	// Passed: true
}

func Example_registry() {
	actions, err := rules.DefineActions[Email, EmailActions[Email]]()
	if err != nil {
		log.Fatal(err)
	}
	compiler, err := anyexpr.NewCompiler[Email]()
	if err != nil {
		log.Fatal(err)
	}

	reg, err := rules.NewRegistry(compiler, actions)
	if err != nil {
		log.Fatal(err)
	}

	reg.Add(rules.Definition{
		Name: "invoices",
		When: `has(Subject, "invoice")`,
		Then: []rules.ActionEntry{{Name: "label", Value: "billing"}},
	})

	reg.Add(rules.Definition{
		Name: "spam",
		When: `has(From, "junk")`,
		Then: []rules.ActionEntry{{Name: "delete"}},
	})

	fmt.Println("Rules:", reg.Len())

	reg.Remove("spam")
	fmt.Println("After remove:", reg.Len())

	prog, err := reg.Compile()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Compiled:", !prog.IsZero())

	// Output:
	// Rules: 2
	// After remove: 1
	// Compiled: true
}
