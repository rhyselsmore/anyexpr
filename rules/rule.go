package rules

// Definition is the input to rule compilation. Consumers construct
// definitions however they like — from YAML, JSON, a database, or
// directly in code.
type Definition struct {
	Name    string
	Tags    []string
	Enabled *bool
	Stop    bool
	When    string
	Skip    string // optional expression — if true, rule is skipped
	Mode    EvalMode
	Then    []ActionEntry
}

// IsEnabled returns whether the rule is enabled. Rules are enabled by
// default (nil Enabled field is treated as true).
func (d Definition) IsEnabled() bool {
	if d.Enabled == nil {
		return true
	}
	return *d.Enabled
}

// ActionEntry is a single action within a rule's Then list.
type ActionEntry struct {
	Name  string
	Value any
}

// EvalMode controls the order of When and Skip expression evaluation.
type EvalMode int

const (
	// WhenThenSkip evaluates When first. If it matches, Skip is
	// checked. If Skip returns true, the rule is skipped despite
	// matching. This is the default.
	WhenThenSkip EvalMode = iota

	// SkipThenWhen evaluates Skip first. If Skip returns true, the
	// When expression is never evaluated — the rule is skipped
	// without paying the cost of the match expression.
	SkipThenWhen
)

func (m EvalMode) String() string {
	switch m {
	case WhenThenSkip:
		return "when-then-skip"
	case SkipThenWhen:
		return "skip-then-when"
	default:
		return "unknown"
	}
}
