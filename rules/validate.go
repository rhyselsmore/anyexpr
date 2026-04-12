package rules

import "fmt"

// ValidateActions checks that a list of action defs for a single rule
// satisfies all constraints:
//   - at most one terminal action
//   - single-cardinality actions appear at most once
func ValidateActions(defs []Def) error {
	terminalCount := 0
	singleSeen := make(map[string]bool)

	for _, d := range defs {
		if d.Terminal {
			terminalCount++
			if terminalCount > 1 {
				return fmt.Errorf("%w: %q", ErrMultipleTerminals, d.Name)
			}
		}
		if d.Cardinality == Single {
			if singleSeen[d.Name] {
				return fmt.Errorf("%w: %q", ErrCardinalityViolation, d.Name)
			}
			singleSeen[d.Name] = true
		}
	}

	return nil
}
