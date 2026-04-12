package rules

import (
	"fmt"
	"strconv"

	"github.com/rhyselsmore/anyexpr"
)

type compiledRule[T any] struct {
	def     Definition
	matcher *anyexpr.Program[T]
	actions []compiledAction[T]
	stop    bool
}

type compiledAction[T any] struct {
	def       Def
	valueExpr *anyexpr.Program[T] // nil if static or no value
	static    string
	boolVal   *bool
}

// Ruleset is a compiled, immutable collection of rules. It is
// parameterised on T only — expression compilation and action
// validation do not need V.
type Ruleset[T any] struct {
	rules []compiledRule[T]
}

// Compile compiles a slice of definitions into a Ruleset. It validates
// all expressions, action references, cardinality constraints, and value
// types at compile time.
func Compile[T any](
	reg *Registry,
	compiler *anyexpr.Compiler[T],
	defs []Definition,
	opts ...CompileOpt,
) (*Ruleset[T], error) {
	_ = opts // reserved

	// Check for duplicate rule names.
	seen := make(map[string]bool, len(defs))
	for _, d := range defs {
		if d.Name == "" {
			continue
		}
		if seen[d.Name] {
			return nil, fmt.Errorf("%w: %q", ErrDuplicateRule, d.Name)
		}
		seen[d.Name] = true
	}

	compiled := make([]compiledRule[T], 0, len(defs))

	for _, d := range defs {
		// Compile the when expression.
		matcher, err := compiler.Compile(anyexpr.NewSource(d.Name, d.When))
		if err != nil {
			return nil, fmt.Errorf("%w: rule %q: %v", ErrCompile, d.Name, err)
		}

		// Compile actions.
		var actions []compiledAction[T]
		var actionDefs []Def

		for _, ae := range d.Then {
			def, ok := reg.LookupAction(ae.Name)
			if !ok {
				return nil, fmt.Errorf("%w: %q in rule %q", ErrUnknownAction, ae.Name, d.Name)
			}

			ca := compiledAction[T]{def: def}

			if def.IsHandler {
				// Verify handler exists.
				if _, ok := reg.LookupHandler(ae.Name); !ok {
					return nil, fmt.Errorf("%w: %q in rule %q", ErrUnknownHandler, ae.Name, d.Name)
				}
				actions = append(actions, ca)
				actionDefs = append(actionDefs, def)
				continue
			}

			// Validate and compile value based on ValueKind.
			switch def.Value {
			case NoValue:
				if ae.Value != "" {
					return nil, fmt.Errorf("%w: action %q in rule %q expects no value, got %q",
						ErrValueType, ae.Name, d.Name, ae.Value)
				}
			case BoolValue:
				b, err := strconv.ParseBool(ae.Value)
				if err != nil {
					return nil, fmt.Errorf("%w: action %q in rule %q: %q is not a valid bool",
						ErrValueType, ae.Name, d.Name, ae.Value)
				}
				ca.boolVal = &b
			case StringVal:
				ca.static = ae.Value
			case StringExpr:
				if ae.Value == "" {
					ca.static = ""
				} else {
					prog, err := compiler.Compile(
						anyexpr.NewSource(d.Name+"/"+ae.Name, ae.Value))
					if err != nil {
						return nil, fmt.Errorf("%w: rule %q action %q value: %v",
							ErrCompile, d.Name, ae.Name, err)
					}
					ca.valueExpr = prog
				}
			}

			actions = append(actions, ca)
			actionDefs = append(actionDefs, def)
		}

		// Validate action constraints for this rule.
		if err := ValidateActions(actionDefs); err != nil {
			return nil, fmt.Errorf("rule %q: %w", d.Name, err)
		}

		// Determine if this rule should stop evaluation.
		hasTerminal := false
		for _, ad := range actionDefs {
			if ad.Terminal {
				hasTerminal = true
				break
			}
		}

		compiled = append(compiled, compiledRule[T]{
			def:     d,
			matcher: matcher,
			actions: actions,
			stop:    d.Stop || hasTerminal,
		})
	}

	return &Ruleset[T]{rules: compiled}, nil
}

// Names returns the names of all rules in evaluation order.
func (rs *Ruleset[T]) Names() []string {
	names := make([]string, len(rs.rules))
	for i, r := range rs.rules {
		names[i] = r.def.Name
	}
	return names
}

// Tags returns the unique set of tags across all rules.
func (rs *Ruleset[T]) Tags() []string {
	seen := make(map[string]bool)
	var tags []string
	for _, r := range rs.rules {
		for _, t := range r.def.Tags {
			if !seen[t] {
				seen[t] = true
				tags = append(tags, t)
			}
		}
	}
	return tags
}

// Len returns the number of compiled rules.
func (rs *Ruleset[T]) Len() int {
	return len(rs.rules)
}

// MergeOpt controls merge behaviour.
type MergeOpt int

// AllowOverride allows the second ruleset's rules to replace the first's
// when names collide, keeping the original's position in evaluation order.
const AllowOverride MergeOpt = 1

// Merge combines two rulesets into a new one. Neither input is modified.
//
// By default, name collisions return ErrNameCollision. With AllowOverride,
// the second ruleset's rule replaces the first's at the original position.
// Remaining rules from the second are appended.
func (rs *Ruleset[T]) Merge(other *Ruleset[T], opts ...MergeOpt) (*Ruleset[T], error) {
	allowOverride := false
	for _, o := range opts {
		if o == AllowOverride {
			allowOverride = true
		}
	}

	// Index the other ruleset by name.
	otherByName := make(map[string]compiledRule[T])
	for _, r := range other.rules {
		if r.def.Name != "" {
			otherByName[r.def.Name] = r
		}
	}

	// Check for collisions.
	if !allowOverride {
		for _, r := range rs.rules {
			if r.def.Name == "" {
				continue
			}
			if _, exists := otherByName[r.def.Name]; exists {
				return nil, fmt.Errorf("%w: %q", ErrNameCollision, r.def.Name)
			}
		}
	}

	// Build merged rules.
	usedFromOther := make(map[string]bool)
	var merged []compiledRule[T]

	for _, r := range rs.rules {
		if r.def.Name != "" {
			if override, exists := otherByName[r.def.Name]; exists {
				merged = append(merged, override)
				usedFromOther[r.def.Name] = true
				continue
			}
		}
		merged = append(merged, r)
	}

	// Append remaining rules from other.
	for _, r := range other.rules {
		if r.def.Name == "" || !usedFromOther[r.def.Name] {
			merged = append(merged, r)
		}
	}

	return &Ruleset[T]{rules: merged}, nil
}
