package rules

import "errors"

var (
	// ErrDefine is returned when DefineActions fails validation.
	ErrDefine = errors.New("rules: action definition failed")

	// ErrDuplicateRegistration is returned when two action fields share
	// the same tag name.
	ErrDuplicateRegistration = errors.New("rules: duplicate registration")

	// ErrDefinitionDuplicate is returned when two rule definitions carry
	// the same name during compilation.
	ErrDefinitionDuplicate = errors.New("rules: duplicate definition")

	// ErrCompile is returned when a when-expression fails to compile.
	ErrCompile = errors.New("rules: compilation failed")

	// ErrNoDefinitions is returned when Compile is called with an
	// empty definitions slice.
	ErrNoDefinitions = errors.New("rules: no rule definitions provided")

	// ErrUnknownAction is returned when a rule references an action
	// name that was not registered.
	ErrUnknownAction = errors.New("rules: unknown action")

	// ErrCardinalityViolation is returned when a single-cardinality
	// action appears more than once in the same rule.
	ErrCardinalityViolation = errors.New("rules: single-use action used multiple times")

	// ErrMultipleTerminals is returned when a single rule contains
	// more than one terminal action.
	ErrMultipleTerminals = errors.New("rules: multiple terminal actions in rule")

	// ErrActionValueType is returned when an action's value does not
	// match the expected type from the definition.
	ErrActionValueType = errors.New("rules: action value type mismatch")

	// ErrProgramZero is returned when a nil or uncompiled Program is
	// passed to NewEvaluator.
	ErrProgramZero = errors.New("rules: program is nil or not compiled")

	// ErrActionsZero is returned when a nil or uninitialised Actions
	// registry is passed to Compile.
	ErrActionsZero = errors.New("rules: actions registry is nil or not initialized")

	// ErrUnknownDefinition is returned when Update is called with a
	// definition name that is not registered.
	ErrUnknownDefinition = errors.New("rules: unknown definition")

	// ErrAssert is returned when an assertion expression fails to
	// compile or evaluate.
	ErrAssert = errors.New("rules: assertion error")

	// ErrAssertFailed is returned when an assertion expression
	// evaluates to false.
	ErrAssertFailed = errors.New("rules: assertion failed")
)
