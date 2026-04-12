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
	Value string
}
