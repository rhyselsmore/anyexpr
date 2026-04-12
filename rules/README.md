# anyexpr/rules

A generic rule evaluation engine built on [anyexpr](../README.md).

Rules follow a when/then model: match an expression against a typed
environment, accumulate actions, resolve them, and optionally execute
handlers. The engine is domain-agnostic — the consuming package
registers its own action names, custom functions, and handler
implementations.

## Install

```bash
go get github.com/rhyselsmore/anyexpr/rules
```

## Quick Start

```go
type Transaction struct {
    Merchant string
    Amount   float64
    Currency string
}

// Register domain-specific actions.
reg, _ := rules.NewRegistry(
    rules.WithAction("categorize", rules.Single, rules.StringVal, false),
    rules.WithAction("tag",        rules.Multi,  rules.StringVal, false),
)

// Compile rules.
compiler, _ := anyexpr.NewCompiler[Transaction]()
rs, _ := rules.Compile(reg, compiler, []rules.Definition{
    {
        Name: "groceries",
        When: `has(Merchant, "woolworths") && Currency == "AUD"`,
        Then: []rules.ActionEntry{
            {Name: "categorize", Value: "groceries"},
            {Name: "tag", Value: "supermarket"},
        },
    },
})

// Run.
engine, _ := rules.NewEngine[Transaction, struct{}](reg, rs)
result, _ := engine.Run(ctx, tx, struct{}{})

fmt.Println(result.Actions.ByName["categorize"]) // [groceries]
fmt.Println(result.Actions.ByName["tag"])         // [supermarket]
```

## Type Parameters

The engine uses two type parameters:

- **`T`** — the environment type. Expressions are compiled and evaluated
  against `T`. Flows from compilation through to execution.
- **`V`** — the vars type. Domain-specific context passed to handlers
  (DB connections, API clients, etc.). Only appears at the engine and
  handler boundary.

If you don't need handler vars, use `struct{}`.

## Actions

Actions are the things rules do. They are registered on a `Registry`
by the domain layer — the engine has no built-in actions.

```go
rules.WithAction("label",    rules.Multi,  rules.StringExpr, false)
rules.WithAction("archive",  rules.Single, rules.NoValue,    true)  // terminal
rules.WithAction("read",     rules.Single, rules.BoolValue,  false)
```

### Cardinality

- **Multi** — may appear multiple times, accumulates across rules, duplicates stripped.
- **Single** — at most once per rule, last-wins across rules.

### Value Kinds

| Kind | Description |
|------|-------------|
| `NoValue` | No value (e.g. delete, archive) |
| `BoolValue` | Bool literal (e.g. read: true) |
| `StringVal` | Static string |
| `StringExpr` | Expression evaluated against `T` at runtime |

### Terminal Actions

A terminal action halts rule evaluation. A rule may contain at most one
terminal action, enforced at compile time. Accumulated actions from
earlier rules still resolve.

## Handlers

Handlers are functions executed after rule evaluation. They receive a
typed `Context[T, V]` with the environment, resolved actions, and vars:

```go
handler := func(ctx *rules.Context[Message, MailVars]) error {
    // ctx.Env — the message being evaluated
    // ctx.Actions — all resolved actions
    // ctx.Vars — domain-specific dependencies
    return nil
}

reg, _ := rules.NewRegistry(
    rules.WithHandler("process-invoice", handler, rules.Multi, false),
)
```

Handler errors don't abort execution — all errors are collected and
returned via `errors.Join`. The result is always populated.

## Compilation

`Compile` validates everything at compile time:

- Duplicate rule names
- Unknown action/handler references
- Expression parse and type errors
- Cardinality violations (single-use action repeated)
- Multiple terminal actions in one rule
- Value type mismatches

```go
rs, err := rules.Compile(reg, compiler, defs)
// err catches all of the above
```

## Rule Definitions

Definitions are plain structs — construct them however you like (YAML,
JSON, database, code):

```go
rules.Definition{
    Name:    "my-rule",
    Tags:    []string{"billing", "receipts"},
    Enabled: nil,         // nil = enabled (default)
    Stop:    false,        // halt evaluation after this rule
    When:    `has(Subject, "invoice")`,
    Then:    []rules.ActionEntry{
        {Name: "label", Value: "billing"},
    },
}
```

## Selectors

Filter which rules execute at engine construction or per-execution:

```go
// Engine-level: only run rules tagged "billing".
engine, _ := rules.NewEngine[T, V](reg, rs, rules.WithTags("billing"))

// Per-execution: additionally exclude "archived" rules.
result, _ := engine.Run(ctx, env, vars, rules.ExcludeTags("archived"))
```

## Merging Rulesets

Combine rulesets from different sources:

```go
merged, err := base.Merge(overrides, rules.AllowOverride)
```

- Default: name collision returns an error.
- `AllowOverride`: the second ruleset's rule replaces the first's,
  keeping the original's position in evaluation order.

## DryRun

Preview which rules would match without executing handlers:

```go
result, _ := engine.DryRun(ctx, env, struct{}{})
```

## License

MIT
