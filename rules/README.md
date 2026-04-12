# anyexpr/rules

A typed rule evaluation engine built on [anyexpr](../README.md).

Define actions as a struct with generic fields, compile rule definitions
with type-checked values, and evaluate them against your environment
type. Results are accessed through typed struct fields — no string keys,
no type assertions.

## Install

```bash
go get github.com/rhyselsmore/anyexpr/rules
```

## Quick Start

```go
type Email struct {
    From    string
    Subject string
    Body    string
    Amount  float64
}

// Declare actions as a generic struct. E is the environment type.
type Actions[E any] struct {
    Label    rules.Action[string, E]       `rule:"label,multi"`
    Move     rules.Action[string, E]       `rule:"move"`
    Read     rules.Action[bool, E]         `rule:"read"`
    Priority rules.Action[int, E]          `rule:"priority"`
    Delete   rules.Action[rules.NoArgs, E] `rule:"delete,terminal"`
}

// Define actions — reflects over the struct once, configures all fields.
actions, err := rules.DefineActions[Actions[Email], Email]()

// Build the anyexpr compiler for your environment type.
compiler, err := anyexpr.NewCompiler[Email]()

// Compile rules — values are type-checked against the action's type.
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
        },
    },
    {
        Name: "spam",
        When: `has(From, "noreply@junk.com")`,
        Then: []rules.ActionEntry{
            {Name: "delete", Value: rules.NoArgs{}},
        },
    },
})

// Build evaluator.
evaluator, err := rules.NewEvaluator(actions, rs)

// Evaluate — returns typed results, no side effects.
eval, err := evaluator.Run(ctx, email)

// Read results through typed accessors.
labels := eval.Actions.Label.Values()       // []string
folder, ok := eval.Actions.Move.Value()     // string, bool
read, ok := eval.Actions.Read.Value()       // bool, bool
priority, ok := eval.Actions.Priority.Value() // int, bool
deleted := eval.Actions.Delete.Fired()      // bool
```

## Actions

Actions are declared as fields on a struct. Each field is an
`Action[T, E]` where `T` is the value type and `E` is the environment
type. The `rule` struct tag configures the action name and options.

### Supported Types

Action values are constrained by the `Actionable` interface, mapping
to JSON primitives:

| Type | Go type | Example value |
|------|---------|---------------|
| String | `Action[string, E]` | `"billing"` |
| Boolean | `Action[bool, E]` | `true` |
| Integer | `Action[int, E]` | `42` |
| Float | `Action[float64, E]` | `0.95` |
| Presence | `Action[NoArgs, E]` | `NoArgs{}` |

### Cardinality

- **Single** (default) — at most once per rule. Across rules, last
  match wins.
- **Multi** (`rule:"name,multi"`) — accumulates across rules,
  duplicates stripped.

### Terminal

An action marked `rule:"name,terminal"` halts evaluation when
triggered. At most one terminal action per struct, enforced by
`DefineActions`. Terminal implies stop.

## Typed Accessors

Every `Action[T, E]` field exposes typed methods on the evaluation
result:

```go
// Resolved values.
eval.Actions.Label.Value()       // (T, bool) — last value, ok
eval.Actions.Label.Values()      // []T — all values (deduped for Multi)
eval.Actions.Label.Fired()       // bool — was it triggered?

// Provenance — which rules contributed.
eval.Actions.Label.ByRule("invoices")  // []T — values from that rule
eval.Actions.Label.ByTag("billing")    // []T — values from rules with that tag
eval.Actions.Label.Rules()             // []string — rule names, deduped
```

## Compile-Time Validation

`Compile` validates everything upfront:

- Duplicate rule names
- Expression parse/type errors (the `when` clause)
- Unknown action names
- Value type mismatches (`"banana"` for a `bool` action)
- Single-cardinality action used multiple times in one rule
- Multiple terminal actions in one rule

## Selectors

Filter which rules evaluate, at the evaluator level or per-call:

```go
// Evaluator-level defaults — applied to every Run.
evaluator, _ := rules.NewEvaluator(actions, rs,
    rules.OnEvaluation(
        rules.WithTags("billing"),
    ),
)

// Per-call — additive with evaluator defaults.
eval, _ := evaluator.Run(ctx, email,
    rules.ExcludeTags("archived"),
)
```

Options: `WithTags`, `WithNames`, `ExcludeTags`, `ExcludeNames`.

## Rule Definitions

Definitions are plain structs — build them from YAML, JSON, a database,
or code:

```go
rules.Definition{
    Name:    "my-rule",
    Tags:    []string{"billing", "receipts"},
    Enabled: nil,          // nil = enabled (default)
    Stop:    false,         // halt evaluation after this rule
    When:    `has(Subject, "invoice")`,
    Then:    []rules.ActionEntry{
        {Name: "label", Value: "billing"},
    },
}
```

`ActionEntry.Value` is `any` — the type is checked against the action's
constraint at compile time.

## Merging Rulesets

Combine rulesets from different sources:

```go
merged, err := base.Merge(overrides, rules.AllowOverride)
```

- Default: name collision returns an error.
- `AllowOverride`: the second ruleset's rule replaces the first's,
  keeping the original's position in evaluation order.

## License

MIT
