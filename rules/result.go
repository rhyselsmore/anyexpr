package rules

// Result is the outcome of a rule engine execution.
type Result struct {
	Matched   []MatchedRule
	Actions   ResolvedActions
	Stopped   bool
	StoppedBy string
}

// MatchedRule records a single rule that matched during evaluation.
type MatchedRule struct {
	Name    string
	Tags    []string
	Actions []ResolvedAction
}

// ResolvedAction is a single action produced by a matched rule.
type ResolvedAction struct {
	Name  string
	Value string
}

// ResolvedActions is the collapsed set of all actions after resolution.
type ResolvedActions struct {
	ByName   map[string][]string // action name → resolved values
	Flags    map[string]*bool    // bool-valued actions
	Terminal bool                // a terminal action was triggered
	Handlers []string            // handler names in execution order
}
