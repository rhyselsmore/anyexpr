package anyexpr

import (
	"fmt"

	"github.com/expr-lang/expr"
)

// Compiler compiles expression sources into programs. It is parameterised
// on T, the type of the environment passed to expressions at evaluation
// time. A Compiler is immutable after construction and safe for
// concurrent use.
type Compiler[T any] struct {
	exprOpts []expr.Option
}

// NewCompiler creates a new Compiler. Options configure custom functions
// and built-in overrides. Returns an error if any option fails validation
// (duplicate names, conflicts).
func NewCompiler[T any](opts ...CompilerOpt) (*Compiler[T], error) {
	cfg := &compilerConfig{}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	exprOpts := buildExprOpts[T](cfg)

	return &Compiler[T]{exprOpts: exprOpts}, nil
}

// wrapBool2 wraps a func(string, string) bool for expr.Function registration.
func wrapBool2(name string, fn func(string, string) bool) expr.Option {
	return expr.Function(name,
		func(params ...any) (any, error) {
			return fn(params[0].(string), params[1].(string)), nil
		},
		new(func(string, string) bool),
	)
}

// wrapString1 wraps a func(string) string for expr.Function registration.
func wrapString1(name string, fn func(string) string) expr.Option {
	return expr.Function(name,
		func(params ...any) (any, error) {
			return fn(params[0].(string)), nil
		},
		new(func(string) string),
	)
}

// wrapStrings1 wraps a func(string) []string for expr.Function registration.
func wrapStrings1(name string, fn func(string) []string) expr.Option {
	return expr.Function(name,
		func(params ...any) (any, error) {
			return fn(params[0].(string)), nil
		},
		new(func(string) []string),
	)
}

// wrapString2 wraps a func(string, string) string for expr.Function registration.
func wrapString2(name string, fn func(string, string) string) expr.Option {
	return expr.Function(name,
		func(params ...any) (any, error) {
			return fn(params[0].(string), params[1].(string)), nil
		},
		new(func(string, string) string),
	)
}


// defaultBuiltinOpts returns the expr.Option slice for all built-in functions.
func defaultBuiltinOpts() []expr.Option {
	return []expr.Option{
		// Case-insensitive string matching
		wrapBool2("has", biHas),
		wrapBool2("starts", biStarts),
		wrapBool2("ends", biEnds),
		wrapBool2("eq", biEq),

		// Case-sensitive string matching
		wrapBool2("xhas", biXhas),
		wrapBool2("xstarts", biXstarts),
		wrapBool2("xends", biXends),

		// Pattern matching
		wrapBool2("re", biRe),
		wrapBool2("xre", biXre),
		wrapBool2("glob", biGlob),

		// Transformation
		wrapString1("lower", biLower),
		wrapString1("upper", biUpper),
		wrapString1("trim", biTrim),
		wrapStrings1("words", biWords),
		wrapStrings1("lines", biLines),

		// Extraction
		wrapString2("extract", biExtract),
		wrapString1("email_domain", biDomain),
	}
}

// builtinOptsByName returns a name→Option map, used when replacements are needed.
func builtinOptsByName() map[string]expr.Option {
	return map[string]expr.Option{
		"has": wrapBool2("has", biHas), "starts": wrapBool2("starts", biStarts),
		"ends": wrapBool2("ends", biEnds), "eq": wrapBool2("eq", biEq),
		"xhas": wrapBool2("xhas", biXhas), "xstarts": wrapBool2("xstarts", biXstarts),
		"xends": wrapBool2("xends", biXends),
		"re": wrapBool2("re", biRe), "xre": wrapBool2("xre", biXre),
		"glob": wrapBool2("glob", biGlob),
		"lower": wrapString1("lower", biLower), "upper": wrapString1("upper", biUpper),
		"trim": wrapString1("trim", biTrim),
		"words": wrapStrings1("words", biWords), "lines": wrapStrings1("lines", biLines),
		"extract": wrapString2("extract", biExtract), "email_domain": wrapString1("email_domain", biDomain),
	}
}

// buildExprOpts assembles the expr.Option slice from the compiler config.
func buildExprOpts[T any](cfg *compilerConfig) []expr.Option {
	var opts []expr.Option

	if len(cfg.replacedFuncs) == 0 {
		// Fast path: no replacements, use defaults directly.
		opts = append(opts, defaultBuiltinOpts()...)
	} else {
		// Build with replacements applied.
		byName := builtinOptsByName()
		for name, fn := range cfg.replacedFuncs {
			// Replaced functions are registered as raw expr.Function options
			// by the caller via ReplaceFunction, so we wrap them here.
			// The fn is whatever the caller provided — we trust it matches
			// the expr.Function contract since it will fail at compile time
			// if not.
			byName[name] = expr.Function(name, fn.(func(...any) (any, error)))
		}
		for _, opt := range byName {
			opts = append(opts, opt)
		}
	}

	// Register custom functions.
	for name, fn := range cfg.customFuncs {
		opts = append(opts, expr.Function(name, fn.(func(...any) (any, error))))
	}

	// Type environment.
	var zero T
	opts = append(opts, expr.Env(zero))

	return opts
}

// Compile compiles a single Source into a Program.
func (c *Compiler[T]) Compile(src *Source, opts ...CompileOpt) (*Program[T], error) {
	_ = opts // reserved

	allOpts := make([]expr.Option, len(c.exprOpts))
	copy(allOpts, c.exprOpts)

	prog, err := expr.Compile(src.Expr(), allOpts...)
	if err != nil {
		return nil, fmt.Errorf("%w: source %q: %v", ErrCompile, src.Name(), err)
	}

	return &Program[T]{
		prog:   prog,
		name:   src.Name(),
		source: src.Expr(),
	}, nil
}

// Check validates one or more sources without producing programs. It
// returns the first error encountered, annotated with the source name.
func (c *Compiler[T]) Check(sources []*Source, opts ...CheckOpt) error {
	_ = opts // reserved
	for _, src := range sources {
		_, err := c.Compile(src)
		if err != nil {
			return err
		}
	}
	return nil
}
