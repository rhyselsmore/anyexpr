package rules_test

import (
	"context"
	"fmt"
	"log"

	"github.com/rhyselsmore/anyexpr"
	"github.com/rhyselsmore/anyexpr/rules"
)

func Example() {
	type Transaction struct {
		Merchant string
		Amount   float64
		Currency string
	}

	// Register domain-specific actions.
	reg, err := rules.NewRegistry(
		rules.WithAction("categorize", rules.Single, rules.StringVal, false),
		rules.WithAction("tag", rules.Multi, rules.StringVal, false),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create a compiler for the environment type.
	compiler, err := anyexpr.NewCompiler[Transaction]()
	if err != nil {
		log.Fatal(err)
	}

	// Compile rules from definitions.
	rs, err := rules.Compile(reg, compiler, []rules.Definition{
		{
			Name: "groceries",
			When: `has(Merchant, "woolworths") && Currency == "AUD"`,
			Then: []rules.ActionEntry{
				{Name: "categorize", Value: "groceries"},
				{Name: "tag", Value: "supermarket"},
			},
		},
		{
			Name: "high-value",
			When: `Amount > 100`,
			Then: []rules.ActionEntry{
				{Name: "tag", Value: "high-value"},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Build and run the engine.
	engine, err := rules.NewEngine[Transaction, struct{}](reg, rs)
	if err != nil {
		log.Fatal(err)
	}

	result, err := engine.Run(context.Background(),
		Transaction{Merchant: "Woolworths Metro", Amount: 142.50, Currency: "AUD"},
		struct{}{},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("matched:", len(result.Matched))
	fmt.Println("category:", result.Actions.ByName["categorize"])
	fmt.Println("tags:", result.Actions.ByName["tag"])
	// Output:
	// matched: 2
	// category: [groceries]
	// tags: [supermarket high-value]
}

func Example_handlers() {
	type Event struct {
		Type    string
		Payload string
	}

	type Deps struct {
		Notified bool
	}

	handler := func(ctx *rules.Context[Event, *Deps]) error {
		ctx.Vars.Notified = true
		return nil
	}

	reg, err := rules.NewRegistry(
		rules.WithAction("tag", rules.Multi, rules.StringVal, false),
		rules.WithHandler("notify", handler, rules.Multi, false),
	)
	if err != nil {
		log.Fatal(err)
	}

	compiler, _ := anyexpr.NewCompiler[Event]()
	rs, _ := rules.Compile(reg, compiler, []rules.Definition{
		{
			Name: "critical",
			When: `eq(Type, "error")`,
			Then: []rules.ActionEntry{
				{Name: "tag", Value: "critical"},
				{Name: "notify"},
			},
		},
	})

	engine, _ := rules.NewEngine[Event, *Deps](reg, rs)
	deps := &Deps{}
	result, _ := engine.Run(context.Background(),
		Event{Type: "error", Payload: "disk full"},
		deps,
	)

	fmt.Println("matched:", len(result.Matched))
	fmt.Println("notified:", deps.Notified)
	// Output:
	// matched: 1
	// notified: true
}

func Example_dryRun() {
	type Env struct {
		Name string
	}

	reg, _ := rules.NewRegistry(
		rules.WithAction("tag", rules.Multi, rules.StringVal, false),
	)
	compiler, _ := anyexpr.NewCompiler[Env]()
	rs, _ := rules.Compile(reg, compiler, []rules.Definition{
		{Name: "r1", When: `has(Name, "alice")`, Then: []rules.ActionEntry{{Name: "tag", Value: "vip"}}},
	})
	engine, _ := rules.NewEngine[Env, struct{}](reg, rs)

	result, _ := engine.DryRun(context.Background(), Env{Name: "Alice"}, struct{}{})
	fmt.Println("would match:", len(result.Matched))
	fmt.Println("tags:", result.Actions.ByName["tag"])
	// Output:
	// would match: 1
	// tags: [vip]
}
