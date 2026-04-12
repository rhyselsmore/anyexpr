package anyexpr

import "errors"

var (
	// ErrDuplicateFunction is returned when a function name is registered
	// more than once on the same compiler.
	ErrDuplicateFunction = errors.New("anyexpr: duplicate function registration")

	// ErrBuiltinConflict is returned when WithFunction is called with a
	// name that matches a built-in function. Use ReplaceFunction instead.
	ErrBuiltinConflict = errors.New("anyexpr: function name conflicts with built-in")

	// ErrNotBuiltin is returned when ReplaceFunction is called with a
	// name that is not a known built-in.
	ErrNotBuiltin = errors.New("anyexpr: name is not a built-in function")

	// ErrCompile is returned when an expression fails to parse or type-check.
	ErrCompile = errors.New("anyexpr: compilation failed")

	// ErrTypeMismatch is returned when Match is called on an expression
	// that does not return bool.
	ErrTypeMismatch = errors.New("anyexpr: expression return type mismatch")
)
