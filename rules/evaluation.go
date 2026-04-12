package rules

// Evaluation is the result of evaluating rules. It holds a populated
// copy of the actions struct with all matched values set.
type Evaluation[A any] struct {
	// Actions is the populated copy of the actions struct. Each Action
	// field's accessors (Value, Values, Fired, ByRule, ByTag, Rules)
	// return the resolved results.
	Actions A

	// Matched lists the rules that matched, in evaluation order.
	Matched []MatchedRule

	// Stopped is true if evaluation was halted by a stop or terminal.
	Stopped bool

	// StoppedBy is the name of the rule that halted evaluation.
	StoppedBy string
}

// MatchedRule records a rule that matched during evaluation.
type MatchedRule struct {
	Name    string
	Tags    []string
	Actions []FiredAction
}

// FiredAction records a single action entry from a matched rule.
type FiredAction struct {
	Name  string
	Value string
}
