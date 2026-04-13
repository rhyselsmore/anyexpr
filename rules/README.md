# anyexpr/rules

A typed rule evaluation engine built on [anyexpr](../README.md).

Define actions as a struct with typed fields, compile rule definitions
with type-checked values, evaluate them against your environment, and
read results through typed struct fields. Dispatch evaluation results
to named handlers gated by expressions.

## Install

```bash
go get github.com/rhyselsmore/anyexpr
```

## Quick Start

```go
type Email struct {
    From    string
    Subject string
    Amount  float64
}

// Declare actions as a generic struct. E is the environment type.
type Actions[E any] struct {
    Label    rules.Action[string, E]       `rule:"label,multi" description:"categorisation labels"`
    Move     rules.Action[string, E]       `rule:"move"`
    Read     rules.Action[bool, E]         `rule:"read"`
    Priority rules.Action[int, E]          `rule:"priority"`
    Delete   rules.Action[action.NoArgs, E] `rule:"delete,terminal"`
}

// Define actions — reflects over the struct once.
actions, err := rules.DefineActions[Email, Actions[Email]]()

// Build the expression compiler.
compiler, err := anyexpr.NewCompiler[Email]()

// Compile rules — values are type-checked against the action's type.
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
        Name: "spam",
        When: `has(From, "noreply@junk.com")`,
        Then: []rules.ActionEntry{
            {Name: "delete"},
        },
    },
})

// Evaluate.
evaluator, err := rules.NewEvaluator(prog)
eval, err := evaluator.Run(ctx, email)

// Read typed results.
eval.Result.Label.Values     // []string{"billing", "invoice"}
eval.Result.Move.Value       // "billing/invoices"
eval.Result.Read.Value       // true
eval.Result.Priority.Value   // 3
eval.Result.Delete.Triggered // false
```

## Actions

Actions are declared as fields on a struct. Each field is an
`Action[V, E]` where `V` is the value type and `E` is the environment
type. The `rule` struct tag configures the action name and options.
An optional `description` tag provides human-readable metadata.

### Supported Value Types

| Type | Go type | Example |
|------|---------|---------|
| String | `Action[string, E]` | `"billing"` |
| Boolean | `Action[bool, E]` | `true` |
| Integer | `Action[int, E]` | `42` |
| Float | `Action[float64, E]` | `0.95` |
| Presence | `Action[action.NoArgs, E]` | `nil` or `action.NoArgs{}` |

### Cardinality

- **Single** (default) — at most once per rule. Across rules, last
  match wins.
- **Multi** (`rule:"name,multi"`) — accumulates across rules,
  duplicates stripped.

### Terminal

An action marked `rule:"name,terminal"` halts evaluation when
triggered. Multiple terminal actions are allowed in the struct (the
user picks which to use). A single rule can reference at most one
terminal action, enforced at compile time.

## Action Results

After evaluation, each `Action` field on the result has:

```go
eval.Result.Label.Triggered  // bool — was it triggered?
eval.Result.Label.Value      // string — last value (last wins for Single)
eval.Result.Label.Values     // []string — all values (deduped for Multi)
eval.Result.Label.Triggers   // []Trigger[string] — full provenance
```

Each `Trigger` records the rule name, tags, and value:

```go
for _, t := range eval.Result.Label.Triggers {
    fmt.Println(t.Rule, t.Tags, t.Value)
}
```

## Rule Definitions

Definitions are plain structs — build them from YAML, JSON, a
database, or code:

```go
rules.Definition{
    Name:    "my-rule",
    Tags:    []string{"billing"},
    Enabled: nil,          // nil = enabled (default)
    Stop:    false,         // halt evaluation after this rule
    When:    `has(Subject, "invoice")`,
    Skip:    `Amount < 10`, // optional skip expression
    Mode:    rules.WhenThenSkip, // or SkipThenWhen
    Then:    []rules.ActionEntry{
        {Name: "label", Value: "billing"},
    },
}
```

`ActionEntry.Value` is `any` — the type is checked against the
action's constraint at compile time.

### Skip Expressions

Rules can have an optional `Skip` expression evaluated against the
environment. If it returns true, the rule is skipped.

`Mode` controls evaluation order:

- **WhenThenSkip** (default) — evaluate When first. If it matches,
  check Skip. If Skip is true, suppress the match.
- **SkipThenWhen** — check Skip first. If true, skip without
  evaluating When (avoids paying the expression cost).

## Compile-Time Validation

`Compile` validates everything upfront:

- Duplicate rule names
- Expression parse/type errors (When and Skip clauses)
- Unknown action names
- Value type mismatches (`"banana"` for a `bool` action)
- Single-cardinality action used multiple times in one rule
- Multiple terminal actions in one rule

## Evaluation

```go
evaluator, err := rules.NewEvaluator(prog)
eval, err := evaluator.Run(ctx, email)
```

The `Evaluation` includes:

- `Env` — the environment value that was evaluated
- `Result` — the actions struct with triggered values
- `Matched` — rule names that matched, in order
- `Stopped` / `StoppedBy` — terminal/stop state
- `StartedAt` / `Duration` — timing
- `Trace` — per-rule steps (when tracing is enabled)

### Tracing

Enable per-rule tracing to see what happened:

```go
eval, err := evaluator.Run(ctx, email, rules.WithTrace(true))

for _, step := range eval.Trace {
    fmt.Println(step.Rule, step.Outcome, step.Duration)
}
```

Each step records the outcome (`matched`, `skipped`, `disabled`,
`excluded`, `skip-expr`), expression duration, evaluation mode,
and which actions were referenced.

### Selectors

Filter which rules evaluate:

```go
// Per-call selectors.
eval, _ := evaluator.Run(ctx, email,
    rules.WithTags("billing"),
    rules.ExcludeNames("spam"),
)

// Evaluator-level defaults.
evaluator, _ := rules.NewEvaluator(prog,
    rules.OnEvaluation(rules.WithTags("billing")),
)

// Expression-based selectors.
sel, _ := rules.WithSelector(`Name != "spam" && "billing" in Tags`)
eval, _ := evaluator.Run(ctx, email, sel)
```

### Debug Output

```go
fmt.Print(eval.Debug())
```

Prints a human-readable summary with timing, matched rules, results,
and trace.

## Introspection

Discover what actions are available:

```go
for _, info := range actions.Describe() {
    fmt.Println(info.Name, info.ValueType, info.Description)
}
```

Returns `[]ActionInfo` with name, description, cardinality, terminal
flag, and value type for each action.

## Registry

For dynamic rule management:

```go
reg, err := rules.NewRegistry(compiler, actions)

reg.Add(def1, def2)       // errors on duplicate names
reg.Update(updatedDef1)   // errors on unknown names
reg.Upsert(def3)          // adds or replaces
reg.Remove("spam")         // silent on unknown

prog, err := reg.Compile()
```

Insertion order is preserved. Safe for concurrent use.

## Expression Validation

Validate an expression without compiling a full program:

```go
err := rules.Check(compiler, `has(Subject, "invoice")`)
```

## Single Rule Testing

Test a rule in isolation:

```go
eval, err := rules.TestRule(compiler, actions, def, email)
```

Returns a full evaluation with tracing enabled.

## Assertions

Write assertions in the same expression language, evaluated against
the actions struct:

```go
result := rules.RunTestCase(compiler, actions, rules.TestCase[Email, Actions[Email]]{
    Name: "invoice labelling",
    Rule: def,
    Env:  Email{Subject: "Your Invoice"},
    Assertions: []string{
        `Label.Triggered`,
        `Label.Value == "billing"`,
        `Priority.Value >= 3`,
        `!Delete.Triggered`,
    },
})
fmt.Println(result.Passed, result.Failures)
```

## Dispatch

The [dispatch](dispatch/) subpackage routes evaluation results to
named handler functions, gated by expressions evaluated against the
full evaluation.

### Handlers

Register handlers immutably on a Dispatcher:

```go
d, err := dispatch.New(
    dispatch.WithLogger[Email, Actions[Email]](logger),
    dispatch.Handle("move-email", moveHandler,
        dispatch.WithDescription("moves email to the target folder"),
    ),
    dispatch.Handle("alert", alertHandler,
        dispatch.WithDescription("sends a priority alert"),
    ),
    dispatch.Handle("audit", auditHandler),
)
```

Discover registered handlers:

```go
for _, h := range d.Describe() {
    fmt.Println(h.Name, h.Description)
}
```

### Plans

Build named, immutable plans that specify which handlers run and
under what conditions:

```go
plan, err := d.Plan(
    dispatch.WithName[Email, Actions[Email]]("billing"),
    dispatch.WithStrategy[Email, Actions[Email]](dispatch.AllContinue),
    dispatch.Gate[Email, Actions[Email]](`len(Matched) > 0`),
    dispatch.Run("move-email",
        dispatch.When[Email, Actions[Email]](`Result.Move.Triggered`),
    ),
    dispatch.Run("alert",
        dispatch.When[Email, Actions[Email]](`Result.Priority.Value >= 3`),
    ),
    dispatch.Run[Email, Actions[Email]]("audit"), // no When — always runs
)
```

Inspect a plan:

```go
fmt.Println(plan.Name())
for _, entry := range plan.Describe() {
    fmt.Println(entry.Handler, entry.Whens)
}
```

### Execution

```go
result := plan.Execute(ctx, eval)

fmt.Println(result.OK())        // true if no errors
fmt.Println(result.Dispatched)  // what ran, timing, errors
fmt.Print(result.Debug())       // human-readable summary
```

### Strategies

- **AllContinue** (default) — run all matching handlers, collect
  errors.
- **AllHaltOnError** — stop on first error.
- **FirstMatch** — run only the first matching handler.

### Features

- **Panic recovery** — handler panics are caught and reported as
  errors.
- **Context cancellation** — respects `ctx.Done()` between handlers.
- **Gate expressions** — top-level expression that must pass before
  any handler runs.
- **When expressions** — per-handler gating. Multiple When expressions
  per handler; any match triggers the handler.
- **Structured logging** — optional `slog.Logger` for dispatch events.
- **Timing** — per-handler and total duration on the result.

## License

MIT
