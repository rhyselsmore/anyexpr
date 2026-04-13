package rules

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/rhyselsmore/anyexpr"
)

// EvaluatorOpt configures an Evaluator.
type EvaluatorOpt func(*evaluatorConfig)

type evaluatorConfig struct {
	evalDefaults evaluationConfig
}

// OnEvaluation sets default evaluation options applied to every Run
// call. Per-call options passed to Run clobber these defaults.
func OnEvaluation(opts ...EvaluationOpt) EvaluatorOpt {
	return func(cfg *evaluatorConfig) {
		for _, o := range opts {
			o(&cfg.evalDefaults)
		}
	}
}

// EvaluationOpt configures a single evaluation (Run call).
type EvaluationOpt func(*evaluationConfig)

type evaluationConfig struct {
	sel        selector
	selectorExpr string // the raw expression string, for trace context
	trace      bool
}

// WithTrace enables per-rule tracing on the evaluation. Off by
// default. When enabled, the Evaluation.Trace slice is populated
// with a Step per rule showing outcome and expression duration.
func WithTrace(enabled bool) EvaluationOpt {
	return func(cfg *evaluationConfig) {
		cfg.trace = enabled
	}
}

// WithSelector filters rules using an expression evaluated against
// RuleMeta (Name string, Tags []string). The expression is compiled
// once when the option is created. Rules that don't pass the
// expression are excluded.
//
// Example: WithSelector(`Name != "spam" && "billing" in Tags`)
func WithSelector(expr string) (EvaluationOpt, error) {
	compiler, err := anyexpr.NewCompiler[RuleMeta]()
	if err != nil {
		return nil, fmt.Errorf("%w: selector compiler: %w", ErrCompile, err)
	}
	prog, err := compiler.Compile(anyexpr.NewSource("selector", expr))
	if err != nil {
		return nil, fmt.Errorf("%w: selector %q: %w", ErrCompile, expr, err)
	}
	return func(cfg *evaluationConfig) {
		cfg.sel.exprFilter = prog
		cfg.selectorExpr = expr
	}, nil
}

// MustWithSelector is like WithSelector but panics on error.
func MustWithSelector(expr string) EvaluationOpt {
	opt, err := WithSelector(expr)
	if err != nil {
		panic(fmt.Sprintf("rules.MustWithSelector: %v", err))
	}
	return opt
}

// WithTags limits evaluation to rules with at least one matching tag.
func WithTags(tags ...string) EvaluationOpt {
	return func(cfg *evaluationConfig) {
		if cfg.sel.onlyTags == nil {
			cfg.sel.onlyTags = make(map[string]bool)
		}
		for _, t := range tags {
			cfg.sel.onlyTags[t] = true
		}
	}
}

// WithNames limits evaluation to rules with matching names.
func WithNames(names ...string) EvaluationOpt {
	return func(cfg *evaluationConfig) {
		if cfg.sel.onlyNames == nil {
			cfg.sel.onlyNames = make(map[string]bool)
		}
		for _, n := range names {
			cfg.sel.onlyNames[n] = true
		}
	}
}

// ExcludeTags excludes rules with any of the given tags.
func ExcludeTags(tags ...string) EvaluationOpt {
	return func(cfg *evaluationConfig) {
		if cfg.sel.excludeTags == nil {
			cfg.sel.excludeTags = make(map[string]bool)
		}
		for _, t := range tags {
			cfg.sel.excludeTags[t] = true
		}
	}
}

// ExcludeNames excludes rules with any of the given names.
func ExcludeNames(names ...string) EvaluationOpt {
	return func(cfg *evaluationConfig) {
		if cfg.sel.excludeNames == nil {
			cfg.sel.excludeNames = make(map[string]bool)
		}
		for _, n := range names {
			cfg.sel.excludeNames[n] = true
		}
	}
}

// Evaluator evaluates rules against an environment and produces typed
// action results. Safe for concurrent use.
//
//   - E is the environment type (e.g. Email).
//   - A is the actions struct (e.g. EmailActions).
type Evaluator[E any, A any] struct {
	program      *Program[E, A]
	actions      *Actions[E, A]
	evalDefaults evaluationConfig
}

// NewEvaluator creates an evaluator from a compiled Program.
//
//   - E is the environment type (e.g. Email).
//   - A is the actions struct (e.g. EmailActions).
func NewEvaluator[E any, A any](
	program *Program[E, A],
	opts ...EvaluatorOpt,
) (*Evaluator[E, A], error) {
	if program.IsZero() {
		return nil, ErrProgramZero
	}

	cfg := &evaluatorConfig{}
	for _, o := range opts {
		o(cfg)
	}

	return &Evaluator[E, A]{
		program:      program,
		actions:      program.actions,
		evalDefaults: cfg.evalDefaults,
	}, nil
}

// Run evaluates rules top-to-bottom against env.
//
//  1. Matches rules against the environment, collecting matched rule
//     names and which actions they reference.
//  2. If any rules matched, copies the action schema and triggers
//     only the actions that were referenced by matched rules.
//  3. Returns an Evaluation with the populated actions, matched rules,
//     timing, and optional trace.
//
// Per-call EvaluationOpts are additive with the evaluator's defaults
// set via OnEvaluation.
func (ev *Evaluator[E, A]) Run(ctx context.Context, env E, opts ...EvaluationOpt) (*Evaluation[E, A], error) {
	cfg := ev.evalDefaults
	for _, o := range opts {
		o(&cfg)
	}
	sel := cfg.sel
	tracing := cfg.trace

	startedAt := time.Now()

	var trace []Step
	if tracing {
		trace = make([]Step, 0, len(ev.program.rules))
	}

	// Phase 1: Match rules.
	matchedRules := make(map[string]struct{})
	matchedOrder := make([]string, 0)
	seenActions := make(map[string]struct{})
	var stopped bool
	var stoppedBy string

	for _, rule := range ev.program.rules {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if !rule.isEnabled() {
			if tracing {
				trace = append(trace, Step{Rule: rule.name, Outcome: OutcomeDisabled, Mode: rule.mode})
			}
			continue
		}

		if !sel.includes(rule.name, rule.tags) {
			if tracing {
				trace = append(trace, Step{Rule: rule.name, Outcome: OutcomeExcluded, Mode: rule.mode, Selector: cfg.selectorExpr})
			}
			continue
		}

		var exprStart time.Time
		if tracing {
			exprStart = time.Now()
		}

		// SkipThenWhen: check skip first, avoid evaluating when if skipped.
		if rule.mode == SkipThenWhen && rule.skipper != nil {
			skipped, err := rule.skipper.Match(env)
			if err != nil {
				return nil, err
			}
			if skipped {
				if tracing {
					trace = append(trace, Step{
						Rule:     rule.name,
						Outcome:  OutcomeSkipExpr,
						Duration: time.Since(exprStart),
						Mode:     rule.mode,
						Skip:     rule.skipExpr,
					})
				}
				continue
			}
		}

		ok, err := rule.matcher.Match(env)
		if err != nil {
			return nil, err
		}

		if !ok {
			if tracing {
				trace = append(trace, Step{
					Rule:     rule.name,
					Outcome:  OutcomeSkipped,
					Duration: time.Since(exprStart),
					Mode:     rule.mode,
				})
			}
			continue
		}

		// WhenThenSkip (default): matched, now check skip.
		if rule.mode == WhenThenSkip && rule.skipper != nil {
			skipped, err := rule.skipper.Match(env)
			if err != nil {
				return nil, err
			}
			if skipped {
				if tracing {
					trace = append(trace, Step{
						Rule:     rule.name,
						Outcome:  OutcomeSkipExpr,
						Duration: time.Since(exprStart),
						Mode:     rule.mode,
						Skip:     rule.skipExpr,
					})
				}
				continue
			}
		}

		matchedRules[rule.name] = struct{}{}
		matchedOrder = append(matchedOrder, rule.name)
		for _, actionName := range rule.actions {
			seenActions[actionName] = struct{}{}
		}

		if tracing {
			trace = append(trace, Step{
				Rule:     rule.name,
				Outcome:  OutcomeMatched,
				Duration: time.Since(exprStart),
				Mode:     rule.mode,
				Actions:  rule.actions,
			})
		}

		if rule.stop {
			stopped = true
			stoppedBy = rule.name
			break
		}
	}

	// Phase 2: Copy schema and trigger matched actions.
	res := ev.actions.schema

	if len(matchedRules) > 0 {
		v := reflect.ValueOf(&res).Elem()

		for actionName := range seenActions {
			idx, ok := ev.actions.fields[actionName]
			if !ok {
				continue
			}
			trg, ok := v.Field(idx).Addr().Interface().(actionTriggerable[E])
			if !ok {
				continue
			}
			trg.trigger(matchedOrder)
		}
	}

	return &Evaluation[E, A]{
		Env:       env,
		Result:    res,
		Matched:   matchedOrder,
		Stopped:   stopped,
		StoppedBy: stoppedBy,
		StartedAt: startedAt,
		Duration:  time.Since(startedAt),
		Traced:    tracing,
		Trace:     trace,
	}, nil
}
