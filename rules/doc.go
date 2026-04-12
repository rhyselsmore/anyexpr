// Package rules is a generic rule evaluation engine built on [anyexpr].
//
// It provides a when/then model: match an expression against a typed
// environment, accumulate actions, resolve them, and optionally execute
// handlers. The engine is domain-agnostic — the consuming package
// registers its own action names, custom functions, and handler
// implementations.
//
// The package uses two type parameters:
//
//   - T is the environment type, used for expression compilation and evaluation.
//   - V is the vars type, domain-specific context passed to handlers at execution time.
//
// [anyexpr]: https://pkg.go.dev/github.com/rhyselsmore/anyexpr
package rules
