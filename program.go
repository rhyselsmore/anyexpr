package anyexpr

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Program is a compiled expression. It is immutable and safe for
// concurrent use. Programs are only produced by Compiler.Compile.
type Program[T any] struct {
	prog   *vm.Program
	name   string
	source string
}

// Name returns the source name provided at compilation.
func (p *Program[T]) Name() string { return p.name }

// Source returns the original expression string.
func (p *Program[T]) Source() string { return p.source }

// Match evaluates the expression against env and returns true/false.
// Returns ErrTypeMismatch if the expression does not return a bool.
func (p *Program[T]) Match(env T, opts ...MatchOpt) (bool, error) {
	cfg := &matchConfig{}
	for _, o := range opts {
		o(cfg)
	}

	out, err := expr.Run(p.prog, env)
	if err != nil {
		return false, fmt.Errorf("anyexpr: eval %q: %w", p.name, err)
	}

	b, ok := out.(bool)
	if !ok {
		return false, fmt.Errorf("%w: %q returned %T, want bool", ErrTypeMismatch, p.name, out)
	}
	return b, nil
}

// Eval evaluates the expression against env and returns the raw result.
func (p *Program[T]) Eval(env T, opts ...EvalOpt) (any, error) {
	cfg := &evalConfig{}
	for _, o := range opts {
		o(cfg)
	}

	out, err := expr.Run(p.prog, env)
	if err != nil {
		return nil, fmt.Errorf("anyexpr: eval %q: %w", p.name, err)
	}
	return out, nil
}
