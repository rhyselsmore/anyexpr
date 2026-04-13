// Package rules is a typed rule evaluation engine built on [anyexpr].
//
// It provides a when/then model: define actions as typed struct fields,
// compile rule definitions with type-checked values, evaluate them
// against a typed environment, and read results through typed fields —
// no string keys, no type assertions.
//
// # Type Parameters
//
// Two type parameters flow through the package:
//
//   - E is the environment type — the struct that expressions evaluate
//     against (e.g. Email, Transaction).
//   - A is the actions struct — a user-defined struct containing
//     [Action] fields with `rule` struct tags.
//
// # Workflow
//
//  1. Define an actions struct with [Action] fields and `rule` tags.
//  2. Call [DefineActions] to reflect over the struct and build the schema.
//  3. Call [Compile] with rule definitions to produce a [Program].
//  4. Call [NewEvaluator] with the program.
//  5. Call [Evaluator.Run] to evaluate rules against an environment value.
//  6. Read typed results from the returned [Evaluation].
//
// # Registry
//
// For dynamic rule management, use [Registry] to add, update, upsert,
// and remove rule definitions, then compile on demand.
//
// # Testing
//
// Use [Check] to validate expressions, [TestRule] to evaluate a single
// rule in isolation, and [RunTestCase] to run assertions against
// evaluation results using the same expression language.
//
// [anyexpr]: https://pkg.go.dev/github.com/rhyselsmore/anyexpr
package rules
