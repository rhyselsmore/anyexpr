package rules

import "errors"

var (
	// ErrDuplicateRegistration is returned when an action or handler name
	// is registered more than once on the same registry.
	ErrDuplicateRegistration = errors.New("rules: duplicate registration")

	// ErrDuplicateRule is returned when two definitions share the same name.
	ErrDuplicateRule = errors.New("rules: duplicate rule name")

	// ErrUnknownAction is returned when a rule references an action name
	// that is not registered on the registry.
	ErrUnknownAction = errors.New("rules: unknown action")

	// ErrUnknownHandler is returned when a rule references a handler name
	// that is not registered on the registry.
	ErrUnknownHandler = errors.New("rules: unknown handler")

	// ErrMultipleTerminals is returned when a single rule contains more
	// than one terminal action.
	ErrMultipleTerminals = errors.New("rules: multiple terminal actions in rule")

	// ErrCardinalityViolation is returned when a single-cardinality action
	// appears more than once in the same rule.
	ErrCardinalityViolation = errors.New("rules: single-use action used multiple times")

	// ErrCompile is returned when a when-expression or value expression
	// fails to compile.
	ErrCompile = errors.New("rules: compilation failed")

	// ErrValueType is returned when an action's value does not match its
	// declared ValueKind.
	ErrValueType = errors.New("rules: action value type mismatch")

	// ErrNameCollision is returned when merging two rulesets that share
	// a rule name without AllowOverride.
	ErrNameCollision = errors.New("rules: rule name collision across rulesets")

	// ErrHandlerType is returned when a registered handler's function
	// signature does not match func(*Context[T, V]) error.
	ErrHandlerType = errors.New("rules: handler function type mismatch")
)
