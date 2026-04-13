package action

import "fmt"

// Cardinality defines how an action accumulates values across rules.
type Cardinality int

const (
	// Single means at most once per rule. Across rules, last match wins.
	Single Cardinality = iota
	// Multi may appear multiple times. All values accumulate, duplicates
	// stripped.
	Multi
)

func (c Cardinality) String() string {
	switch c {
	case Single:
		return "single"
	case Multi:
		return "multi"
	default:
		return fmt.Sprintf("Cardinality(%d)", int(c))
	}
}

// IsValid returns true if the cardinality is a known value.
func (c Cardinality) IsValid() bool {
	return c == Single || c == Multi
}

// Validate returns an error if the cardinality is not a known value.
func (c Cardinality) Validate() error {
	if !c.IsValid() {
		return fmt.Errorf("%w: %d", ErrCardinalityInvalid, int(c))
	}
	return nil
}

// ParseCardinality parses a string into a Cardinality. Accepts
// "single" and "multi" (case-sensitive). Returns an error for
// unrecognised values.
func ParseCardinality(s string) (Cardinality, error) {
	switch s {
	case "single":
		return Single, nil
	case "multi":
		return Multi, nil
	default:
		return 0, fmt.Errorf("%w: %q", ErrCardinalityUnknown, s)
	}
}
