package rules

// Cardinality controls how many times an action may appear per rule and
// how values are resolved across rules.
type Cardinality int

const (
	// Multi actions may appear multiple times and accumulate across rules.
	// Duplicate values are stripped at resolution.
	Multi Cardinality = iota
	// Single actions may appear at most once per rule. Across rules, the
	// last match wins.
	Single
)

// String returns the string representation of the cardinality.
func (c Cardinality) String() string {
	switch c {
	case Multi:
		return "multi"
	case Single:
		return "single"
	default:
		return "unknown"
	}
}

// ValueKind describes what kind of value an action carries.
type ValueKind int

const (
	// NoValue actions carry no value (e.g. delete, archive).
	NoValue ValueKind = iota
	// BoolValue actions carry a bool literal (e.g. read: true).
	BoolValue
	// StringVal actions carry a static string.
	StringVal
	// StringExpr actions carry a string expression evaluated against T.
	StringExpr
)

// String returns the string representation of the value kind.
func (v ValueKind) String() string {
	switch v {
	case NoValue:
		return "none"
	case BoolValue:
		return "bool"
	case StringVal:
		return "string"
	case StringExpr:
		return "expr"
	default:
		return "unknown"
	}
}

// Def describes a registered action or handler.
type Def struct {
	Name        string
	Terminal    bool
	Cardinality Cardinality
	Value       ValueKind
	IsHandler   bool
}

// String returns the action name.
func (d Def) String() string { return d.Name }
