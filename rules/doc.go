// Package rules is a generic, typed rule evaluation engine built on [anyexpr].
//
// It provides a when/then model: match an expression against a typed
// environment, accumulate typed actions, and read results through
// compile-time-safe accessors.
//
// The package uses two type parameters:
//
//   - A is the actions struct type, containing [Action] fields that define
//     what a rule can do. Defined once via [DefineActions].
//   - E is the environment type, used for expression compilation and
//     evaluation via [anyexpr].
//
// [anyexpr]: https://pkg.go.dev/github.com/rhyselsmore/anyexpr
package rules
