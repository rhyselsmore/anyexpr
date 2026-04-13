package rules

import (
	"fmt"

	"github.com/rhyselsmore/anyexpr"
	"github.com/rhyselsmore/anyexpr/rules/action"
)

// CompileOpt configures a Compile call.
type CompileOpt[E any, A any] func(*CompileOpts[E, A]) error

// CompileOpts holds the accumulated configuration for Compile.
type CompileOpts[E any, A any] struct {
}

type compiledRule[E any] struct {
	name     string
	tags     []string
	enabled  *bool
	stop     bool
	mode     EvalMode
	matcher  *anyexpr.Program[E]
	skipper  *anyexpr.Program[E] // nil if no Skip expression
	skipExpr string              // raw skip expression, for trace context
	actions  []string
}

// isEnabled returns whether the rule is enabled. Nil means enabled.
func (r compiledRule[E]) isEnabled() bool {
	return r.enabled == nil || *r.enabled
}

// Program holds compiled rules ready for evaluation.
//
//   - E is the environment type (e.g. Email).
//   - A is the actions struct (e.g. EmailActions).
type Program[E any, A any] struct {
	actions  *Actions[E, A]
	rules    []compiledRule[E]
	compiled bool
}

// IsZero returns true if the program was not created via Compile.
func (p *Program[E, A]) IsZero() bool { return p == nil || !p.compiled }

// Compile validates and compiles rule definitions against the
// registered actions and expression compiler.
//
//   - E is the environment type (e.g. Email).
//   - A is the actions struct (e.g. EmailActions).
//
// Action names in definitions are checked against the bound actions.
// Values are type-checked against the action's value type. Expressions
// are compiled via the anyexpr compiler.
func Compile[E any, A any](
	compiler *anyexpr.Compiler[E],
	actions *Actions[E, A],
	defs []Definition,
	opts ...CompileOpt[E, A],
) (*Program[E, A], error) {
	if actions.IsZero() {
		return nil, ErrActionsZero
	}

	// Build Compilation Opts
	co := &CompileOpts[E, A]{}
	for _, opt := range opts {
		if err := opt(co); err != nil {
			return nil, err
		}
	}

	if len(defs) == 0 {
		return nil, ErrNoDefinitions
	}

	// Build Program
	pro := &Program[E, A]{
		actions:  actions,
		rules:    make([]compiledRule[E], 0, len(defs)),
		compiled: true,
	}

	seen := make(map[string]bool, len(defs))
	for _, def := range defs {
		if seen[def.Name] {
			return nil, fmt.Errorf("%w: %q appears more than once in the definitions list", ErrDefinitionDuplicate, def.Name)
		}
		seen[def.Name] = true

		prog, err := compiler.Compile(anyexpr.NewSource(def.Name, def.When))
		if err != nil {
			return nil, fmt.Errorf("%w: rule %q when: %w", ErrCompile, def.Name, err)
		}

		var skipper *anyexpr.Program[E]
		if def.Skip != "" {
			skipper, err = compiler.Compile(anyexpr.NewSource(def.Name+"/skip", def.Skip))
			if err != nil {
				return nil, fmt.Errorf("%w: rule %q skip: %w", ErrCompile, def.Name, err)
			}
		}

		act, hasTerminal, err := compileRuleActions(actions, def, co)
		if err != nil {
			return nil, err
		}

		pro.rules = append(pro.rules, compiledRule[E]{
			name:     def.Name,
			tags:     def.Tags,
			enabled:  def.Enabled,
			stop:     def.Stop || hasTerminal,
			mode:     def.Mode,
			matcher:  prog,
			skipper:  skipper,
			skipExpr: def.Skip,
			actions:  act,
		})
	}

	return pro, nil
}

func compileRuleActions[E any, A any](actions *Actions[E, A], def Definition, opts *CompileOpts[E, A]) ([]string, bool, error) {
	seenSingle := make(map[string]bool)
	seenActions := make(map[string]struct{})
	terminalCount := 0
	hasTerminal := false

	for _, ae := range def.Then {
		ent, ok := actions.binders[ae.Name]
		if !ok {
			return nil, false, fmt.Errorf("%w: rule %q references %q", ErrUnknownAction, def.Name, ae.Name)
		}

		term, card, err := ent.bind(def.Name, def.Tags, ae.Value)
		if err != nil {
			return nil, false, err
		}
		seenActions[ae.Name] = struct{}{}

		if card == action.Single && seenSingle[ae.Name] {
			return nil, false, fmt.Errorf("%w: rule %q uses %q multiple times", ErrCardinalityViolation, def.Name, ae.Name)
		}
		seenSingle[ae.Name] = true

		if term {
			terminalCount++
			if terminalCount > 1 {
				return nil, false, fmt.Errorf("%w: rule %q", ErrMultipleTerminals, def.Name)
			}
			hasTerminal = true
		}
	}
	act := make([]string, 0, len(seenActions))
	for name := range seenActions {
		act = append(act, name)
	}

	return act, hasTerminal, nil
}
