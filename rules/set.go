package rules

// Entry is a single action entry accumulated during rule evaluation.
type Entry struct {
	Def      Def
	Value    string
	BoolVal  *bool
	RuleName string
}

// Set accumulates action entries during evaluation and resolves them
// into a ResolvedActions.
type Set struct {
	entries []Entry
}

// NewSet creates a new empty action set.
func NewSet() *Set {
	return &Set{}
}

// Add appends an entry to the set.
func (s *Set) Add(e Entry) {
	s.entries = append(s.entries, e)
}

// Entries returns the accumulated entries in order.
func (s *Set) Entries() []Entry {
	return s.entries
}

// HasTerminal returns true if any entry is a terminal action.
func (s *Set) HasTerminal() bool {
	for _, e := range s.entries {
		if e.Def.Terminal {
			return true
		}
	}
	return false
}

// Resolve collapses entries into a ResolvedActions.
//
// Multi-cardinality actions collect all values with duplicates stripped.
// Single-cardinality actions keep the last value (last-wins).
// Bool-valued actions keep the last value.
// Handler names are collected in order with duplicates stripped.
// Empty string values are excluded from ByName slices.
func (s *Set) Resolve() ResolvedActions {
	ra := ResolvedActions{
		ByName:   make(map[string][]string),
		Flags:    make(map[string]*bool),
		Handlers: []string{},
	}

	// Track seen values for multi-action dedup.
	multiSeen := make(map[string]map[string]bool)
	// Track seen handler names for dedup.
	handlerSeen := make(map[string]bool)

	for _, e := range s.entries {
		name := e.Def.Name

		if e.Def.IsHandler {
			if !handlerSeen[name] {
				handlerSeen[name] = true
				ra.Handlers = append(ra.Handlers, name)
			}
			continue
		}

		if e.Def.Terminal {
			ra.Terminal = true
		}

		// Handle bool values.
		if e.Def.Value == BoolValue {
			b := *e.BoolVal
			ra.Flags[name] = &b
			continue
		}

		// Skip empty string values.
		if e.Value == "" {
			continue
		}

		switch e.Def.Cardinality {
		case Multi:
			if multiSeen[name] == nil {
				multiSeen[name] = make(map[string]bool)
			}
			if !multiSeen[name][e.Value] {
				multiSeen[name][e.Value] = true
				ra.ByName[name] = append(ra.ByName[name], e.Value)
			}
		case Single:
			ra.ByName[name] = []string{e.Value}
		}
	}

	return ra
}
