package rules

import "errors"

var (
	// ErrDefine is returned when DefineActions fails validation.
	ErrDefine = errors.New("rules: action definition failed")

	// ErrDuplicateRegistration is returned when two action fields share
	// the same tag name.
	ErrDuplicateRegistration = errors.New("rules: duplicate registration")

	// ErrDuplicateRule is returned when two rule definitions share the
	// same name.
	ErrDuplicateRule = errors.New("rules: duplicate rule name")

	// ErrUnknownAction is returned when a rule references an action name
	// not present in the actions struct.
	ErrUnknownAction = errors.New("rules: unknown action")

	// ErrMultipleTerminals is returned when more than one terminal action
	// is declared or used in a single rule.
	ErrMultipleTerminals = errors.New("rules: multiple terminal actions")

	// ErrCardinalityViolation is returned when a single-cardinality action
	// appears more than once in the same rule.
	ErrCardinalityViolation = errors.New("rules: single-use action used multiple times")

	// ErrCompile is returned when a when-expression fails to compile.
	ErrCompile = errors.New("rules: compilation failed")

	// ErrValueType is returned when an action's value does not match the
	// expected type.
	ErrValueType = errors.New("rules: action value type mismatch")

	// ErrNameCollision is returned when merging rulesets with overlapping
	// rule names without AllowOverride.
	ErrNameCollision = errors.New("rules: rule name collision across rulesets")

	// ErrNotDefined is returned when an uninitialised *Actions is passed
	// to NewEvaluator or Compile.
	ErrNotDefined = errors.New("rules: actions not defined")
)
