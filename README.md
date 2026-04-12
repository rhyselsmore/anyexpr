# anyexpr

A generic expression compilation and evaluation library for Go.

anyexpr wraps [expr-lang](https://expr-lang.org) with a typed compiler,
a library of built-in string and pattern matching functions, and a
compile-once-run-many execution model. It's designed for systems that
evaluate user-authored filter expressions, routing rules, or matching
logic against structured data.

## Install

```bash
go get github.com/rhyselsmore/anyexpr
```

## Quick Start

Define your environment struct, create a compiler, compile an expression,
and evaluate it:

```go
type Email struct {
    From    string
    Subject string
    Body    string
}

// Create a compiler parameterised on your environment type.
compiler, err := anyexpr.NewCompiler[Email]()
if err != nil {
    log.Fatal(err)
}

// Compile an expression. Field names and function calls are validated
// against the environment type at compile time.
prog, err := compiler.Compile(
    anyexpr.NewSource("invoice-filter",
        `has(Subject, "invoice") && ends(From, "stripe.com")`))
if err != nil {
    log.Fatal(err)
}

// Evaluate against a value. Programs are safe for concurrent use.
matched, err := prog.Match(Email{
    From:    "billing@stripe.com",
    Subject: "Your January Invoice",
    Body:    "...",
})
// matched == true
```

## Features

- **Typed compilation** — expressions are validated against your struct
  fields at compile time, not at evaluation time.
- **Built-in functions** — `has`, `starts`, `ends`, `eq`, `re`, `glob`,
  `extract`, `domain`, and more. Case-insensitive by default, with `x`-prefixed
  case-sensitive variants (`xhas`, `xstarts`, etc.).
- **Compile once, run many** — compiled programs are immutable and safe
  for concurrent use across goroutines.
- **Extensible** — register custom functions with `WithFunction`, or
  override built-ins with `ReplaceFunction`.
- **Safe** — expressions run in a sandboxed evaluator with no I/O, no
  imports, and no side effects.

## Built-in Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `has` | `(s, substr) bool` | Case-insensitive substring match |
| `xhas` | `(s, substr) bool` | Case-sensitive substring match |
| `starts` | `(s, prefix) bool` | Case-insensitive prefix match |
| `xstarts` | `(s, prefix) bool` | Case-sensitive prefix match |
| `ends` | `(s, suffix) bool` | Case-insensitive suffix match |
| `xends` | `(s, suffix) bool` | Case-sensitive suffix match |
| `eq` | `(a, b) bool` | Case-insensitive equality |
| `re` | `(s, pattern) bool` | Case-insensitive regex match |
| `xre` | `(s, pattern) bool` | Case-sensitive regex match |
| `glob` | `(s, pattern) bool` | Case-insensitive glob match |
| `lower` | `(s) string` | Lowercase |
| `upper` | `(s) string` | Uppercase |
| `trim` | `(s) string` | Trim whitespace |
| `replace` | `(s, old, new) string` | Replace all occurrences |
| `split` | `(s, sep) []string` | Split on delimiter |
| `words` | `(s) []string` | Split on whitespace |
| `lines` | `(s) []string` | Split on newlines |
| `extract` | `(s, pattern) string` | First regex match |
| `domain` | `(addr) string` | Domain from email address |
| `len` | `(v) int` | Length of string, array, slice, or map |

See [doc/reference.md](doc/reference.md) for full documentation and examples.

## Custom Functions

```go
compiler, err := anyexpr.NewCompiler[MyEnv](
    anyexpr.WithFunction("reverse", func(params ...any) (any, error) {
        s := params[0].(string)
        runes := []rune(s)
        for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
            runes[i], runes[j] = runes[j], runes[i]
        }
        return string(runes), nil
    }),
)
```

## Eval

`Match` returns a bool. For expressions that return other types, use `Eval`:

```go
prog, _ := compiler.Compile(
    anyexpr.NewSource("get-domain", `domain(Email)`))

result, err := prog.Eval(env) // result is "example.com"
```

## Validation

Check expressions without compiling them into programs:

```go
err := compiler.Check([]*anyexpr.Source{
    anyexpr.NewSource("rule-1", `has(Name, "alice")`),
    anyexpr.NewSource("rule-2", `starts(Name, "b")`),
})
```

## Rules Engine

The [anyexpr/rules](rules/) subpackage builds on top of anyexpr with a
when/then rule evaluation engine — register domain-specific actions,
compile rule definitions, and evaluate them against your environment type.

## License

MIT
