// Package anyexpr is a generic expression compilation and evaluation library.
//
// It wraps [expr-lang] with a typed compiler, a library of built-in string
// and pattern matching functions, and a compile-once-run-many execution model.
//
// The typical workflow is:
//
//  1. Create a [Compiler] parameterised on your environment struct.
//  2. Compile expression strings into [Program] values using [Compiler.Compile].
//  3. Evaluate programs against environment values using [Program.Match] or [Program.Eval].
//
// Expressions are validated at compile time against the fields and methods
// of the environment type, catching typos and type errors before evaluation.
//
// [expr-lang]: https://expr-lang.org
package anyexpr
