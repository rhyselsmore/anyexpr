package rules

import (
	"fmt"

	"github.com/rhyselsmore/anyexpr"
)

type compiledRule[E any] struct {
	name    string
	tags    []string
	enabled *bool
	stop    bool
	matcher *anyexpr.Program[E]
	actions []actionValuer[E]
}

// Ruleset holds compiled rules ready for evaluation.
type Ruleset[E any] struct {
	rules []compiledRule[E]
}

// Compile validates and compiles rule definitions against the action
// schema and expression compiler. Action names in definitions are
// checked against the defined actions. Values are type-checked against
// the action's Actionable constraint. Expressions are compiled via the
// anyexpr compiler.
func Compile[A, E any](
	actions *Actions[A, E],
	compiler *anyexpr.Compiler[E],
	defs []Definition,
	opts ...CompileOpt,
) (*Ruleset[E], error) {
	if !actions.defined {
		return nil, ErrNotDefined
	}

	seen := make(map[string]bool)
	compiled := make([]compiledRule[E], 0, len(defs))

	for _, def := range defs {
		if seen[def.Name] {
			return nil, fmt.Errorf("%w: %q", ErrDuplicateRule, def.Name)
		}
		seen[def.Name] = true

		prog, err := compiler.Compile(anyexpr.NewSource(def.Name, def.When))
		if err != nil {
			return nil, fmt.Errorf("%w: rule %q: %v", ErrCompile, def.Name, err)
		}

		cas, hasTerminal, err := compileRuleActions(actions, def)
		if err != nil {
			return nil, err
		}

		compiled = append(compiled, compiledRule[E]{
			name:    def.Name,
			tags:    def.Tags,
			enabled: def.Enabled,
			stop:    def.Stop || hasTerminal,
			matcher: prog,
			actions: cas,
		})
	}

	return &Ruleset[E]{rules: compiled}, nil
}

func compileRuleActions[A, E any](actions *Actions[A, E], def Definition) ([]actionValuer[E], bool, error) {
	cas := make([]actionValuer[E], 0, len(def.Then))
	seenSingle := make(map[string]bool)
	terminalCount := 0
	hasTerminal := false

	for _, ae := range def.Then {
		af, ok := actions.compilers[ae.Name]
		if !ok {
			return nil, false, fmt.Errorf("%w: rule %q references %q", ErrUnknownAction, def.Name, ae.Name)
		}

		av, err := af.compile(ae.Value)
		if err != nil {
			return nil, false, err
		}

		if av.actionCardinality() == Single && seenSingle[ae.Name] {
			return nil, false, fmt.Errorf("%w: rule %q uses %q multiple times", ErrCardinalityViolation, def.Name, ae.Name)
		}
		seenSingle[ae.Name] = true

		if av.actionTerminal() {
			terminalCount++
			if terminalCount > 1 {
				return nil, false, fmt.Errorf("%w: rule %q", ErrMultipleTerminals, def.Name)
			}
			hasTerminal = true
		}

		cas = append(cas, av)
	}

	return cas, hasTerminal, nil
}

// Names returns all rule names in evaluation order.
func (rs *Ruleset[E]) Names() []string {
	names := make([]string, len(rs.rules))
	for i, r := range rs.rules {
		names[i] = r.name
	}
	return names
}

// Len returns the number of rules.
func (rs *Ruleset[E]) Len() int {
	return len(rs.rules)
}

// Merge combines two rulesets. By default, name collisions return an
// error. Use AllowOverride to let the second ruleset replace colliding
// rules while keeping the original's position in evaluation order.
func (rs *Ruleset[E]) Merge(other *Ruleset[E], opts ...MergeOpt) (*Ruleset[E], error) {
	cfg := &mergeConfig{}
	for _, o := range opts {
		o(cfg)
	}

	nameIdx := make(map[string]int, len(rs.rules))
	for i, r := range rs.rules {
		nameIdx[r.name] = i
	}

	merged := make([]compiledRule[E], len(rs.rules))
	copy(merged, rs.rules)

	for _, r := range other.rules {
		if idx, exists := nameIdx[r.name]; exists {
			if !cfg.allowOverride {
				return nil, fmt.Errorf("%w: %q", ErrNameCollision, r.name)
			}
			merged[idx] = r
		} else {
			merged = append(merged, r)
		}
	}

	return &Ruleset[E]{rules: merged}, nil
}
