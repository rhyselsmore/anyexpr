package rules

import (
	"context"
	"reflect"
)

// Evaluator evaluates rules against an environment and produces an
// Evaluation with typed action results. Safe for concurrent use.
type Evaluator[A, E any] struct {
	actions      *Actions[A, E]
	rules        []compiledRule[E]
	evalDefaults evaluationConfig
}

// NewEvaluator creates an evaluator from the action schema and a
// compiled ruleset.
func NewEvaluator[A, E any](
	actions *Actions[A, E],
	ruleset *Ruleset[E],
	opts ...EvaluatorOpt,
) (*Evaluator[A, E], error) {
	if !actions.defined {
		return nil, ErrNotDefined
	}

	cfg := &evaluatorConfig{}
	for _, o := range opts {
		o(cfg)
	}

	return &Evaluator[A, E]{
		actions:      actions,
		rules:        ruleset.rules,
		evalDefaults: cfg.evalDefaults,
	}, nil
}

// Run evaluates rules top-to-bottom against env. It copies the action
// schema, populates values from matching rules, resolves cardinality
// (dedup for Multi, last-wins for Single), and returns the result.
//
// Per-call EvaluationOpts are additive with the evaluator's defaults
// set via OnEvaluation. No side effects — handlers are not executed.
func (ev *Evaluator[A, E]) Run(ctx context.Context, env E, opts ...EvaluationOpt) (*Evaluation[A], error) {
	cfg := ev.evalDefaults
	for _, o := range opts {
		o(&cfg)
	}
	sel := cfg.sel

	// Value-copy the schema. All Action[T, E] fields carry their
	// metadata (name, cardinality, terminal, index) because those are
	// plain values set by DefineActions. The entries slice on each
	// Action is nil on the schema (never populated), so the copy
	// starts with a clean slate — no reset needed.
	res := ev.actions.schema

	// One reflect call per Run to get an addressable handle on the
	// copy. Used by addEntry to write into the correct Action field
	// by index, and by resolve at the end.
	ptr := reflect.ValueOf(&res).Elem()

	var matched []MatchedRule
	var stopped bool
	var stoppedBy string

	for _, rule := range ev.rules {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if rule.enabled != nil && !*rule.enabled {
			continue
		}

		if !sel.includes(rule.name, rule.tags) {
			continue
		}

		ok, err := rule.matcher.Match(env)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		var firedActions []FiredAction
		for _, av := range rule.actions {
			av.addEntry(ptr, rule.name, rule.tags)

			firedActions = append(firedActions, FiredAction{
				Name:  av.actionName(),
				Value: av.stringValue(),
			})
		}

		matched = append(matched, MatchedRule{
			Name:    rule.name,
			Tags:    rule.tags,
			Actions: firedActions,
		})

		if rule.stop {
			stopped = true
			stoppedBy = rule.name
			break
		}
	}

	// Resolve all actions — dedup Multi, keep all for Single.
	for _, idx := range ev.actions.fields {
		ptr.Field(idx).Addr().Interface().(actionResolver).resolve()
	}

	if matched == nil {
		matched = []MatchedRule{}
	}

	return &Evaluation[A]{
		Actions:   res,
		Matched:   matched,
		Stopped:   stopped,
		StoppedBy: stoppedBy,
	}, nil
}
