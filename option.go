package anyexpr

import (
	"fmt"
	"io"
)

// --- Compiler options ---

// CompilerOpt configures a Compiler.
type CompilerOpt func(*compilerConfig) error

type compilerConfig struct {
	customFuncs   map[string]any
	replacedFuncs map[string]any
}

// WithFunction registers a custom function available in all expressions
// compiled by this compiler. The function signature must be compatible
// with expr-lang's function binding.
//
// Returns an error at compiler construction if:
//   - the name has already been registered
//   - the name conflicts with a built-in (use ReplaceFunction instead)
func WithFunction(name string, fn any) CompilerOpt {
	return func(c *compilerConfig) error {
		if builtinNames()[name] {
			return fmt.Errorf("%w: %q", ErrBuiltinConflict, name)
		}
		if _, exists := c.customFuncs[name]; exists {
			return fmt.Errorf("%w: %q", ErrDuplicateFunction, name)
		}
		if c.customFuncs == nil {
			c.customFuncs = make(map[string]any)
		}
		c.customFuncs[name] = fn
		return nil
	}
}

// ReplaceFunction overrides a built-in function. Returns an error at
// compiler construction if the name is not a known built-in.
func ReplaceFunction(name string, fn any) CompilerOpt {
	return func(c *compilerConfig) error {
		if !builtinNames()[name] {
			return fmt.Errorf("%w: %q", ErrNotBuiltin, name)
		}
		if c.replacedFuncs == nil {
			c.replacedFuncs = make(map[string]any)
		}
		c.replacedFuncs[name] = fn
		return nil
	}
}

// --- Compile / Check options ---

// CompileOpt configures a single Compile call.
type CompileOpt func(*compileConfig)

type compileConfig struct{}

// CheckOpt configures a Check call.
type CheckOpt func(*checkConfig)

type checkConfig struct{}

// --- Execution options ---

// MatchOpt configures a single Match call.
type MatchOpt func(*matchConfig)

// EvalOpt configures a single Eval call.
type EvalOpt func(*evalConfig)

type matchConfig struct {
	traceWriter io.Writer
}

type evalConfig struct {
	traceWriter io.Writer
}

// WithMatchTrace writes evaluation trace output to w.
func WithMatchTrace(w io.Writer) MatchOpt {
	return func(c *matchConfig) { c.traceWriter = w }
}

// WithEvalTrace writes evaluation trace output to w.
func WithEvalTrace(w io.Writer) EvalOpt {
	return func(c *evalConfig) { c.traceWriter = w }
}
