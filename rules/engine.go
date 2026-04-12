package rules

import (
	"context"
	"errors"
	"fmt"
)

// Engine binds a registry and a ruleset for execution. It is safe for
// concurrent use.
type Engine[T, V any] struct {
	rules    []compiledRule[T]
	selector selector
	handlers map[string]func(*Context[T, V]) error
}

// NewEngine creates a new Engine. It type-asserts all registered handlers
// against func(*Context[T, V]) error at construction time.
func NewEngine[T, V any](
	reg *Registry,
	ruleset *Ruleset[T],
	opts ...EngineOpt,
) (*Engine[T, V], error) {
	// Type-assert handlers.
	handlers := make(map[string]func(*Context[T, V]) error)
	for _, name := range reg.HandlerNames() {
		raw, _ := reg.LookupHandler(name)
		h, ok := raw.(func(*Context[T, V]) error)
		if !ok {
			return nil, fmt.Errorf("%w: handler %q", ErrHandlerType, name)
		}
		handlers[name] = h
	}

	cfg := &engineConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Shallow copy rules — programs are immutable.
	rules := make([]compiledRule[T], len(ruleset.rules))
	copy(rules, ruleset.rules)

	return &Engine[T, V]{
		rules:    rules,
		selector: cfg.sel,
		handlers: handlers,
	}, nil
}

// Run evaluates all rules against env, resolves actions, and executes
// handlers. It is safe for concurrent use.
func (e *Engine[T, V]) Run(ctx context.Context, env T, vars V, opts ...RunOpt) (*Result, error) {
	return e.run(ctx, env, vars, false, opts...)
}

// DryRun evaluates all rules and resolves actions but does not execute
// handlers. The result is identical to Run except handlers are skipped.
func (e *Engine[T, V]) DryRun(ctx context.Context, env T, vars V, opts ...RunOpt) (*Result, error) {
	return e.run(ctx, env, vars, true, opts...)
}

func (e *Engine[T, V]) run(ctx context.Context, env T, vars V, dry bool, opts ...RunOpt) (*Result, error) {
	// Build per-execution selector.
	rcfg := &runConfig{}
	for _, opt := range opts {
		opt(rcfg)
	}
	sel := e.selector.merge(rcfg.sel)

	actionSet := NewSet()
	var matched []MatchedRule
	stopped := false
	stoppedBy := ""

	for _, rule := range e.rules {
		// Check context cancellation.
		select {
		case <-ctx.Done():
			resolved := actionSet.Resolve()
			return &Result{
				Matched:   matched,
				Actions:   resolved,
				Stopped:   stopped,
				StoppedBy: stoppedBy,
			}, ctx.Err()
		default:
		}

		// Check selector.
		if !sel.includes(rule.def.Name, rule.def.Tags) {
			continue
		}

		// Check enabled.
		if !rule.def.IsEnabled() {
			continue
		}

		// Evaluate expression.
		ok, err := rule.matcher.Match(env)
		if err != nil {
			return nil, fmt.Errorf("rules: eval %q: %w", rule.def.Name, err)
		}
		if !ok {
			continue
		}

		// Rule matched — process actions.
		var ruleActions []ResolvedAction

		for _, ca := range rule.actions {
			entry := Entry{
				Def:      ca.def,
				RuleName: rule.def.Name,
			}

			if ca.def.IsHandler {
				actionSet.Add(entry)
				ruleActions = append(ruleActions, ResolvedAction{
					Name: ca.def.Name,
				})
				continue
			}

			switch ca.def.Value {
			case NoValue:
				// No value to set.
			case BoolValue:
				entry.BoolVal = ca.boolVal
			case StringVal:
				entry.Value = ca.static
			case StringExpr:
				if ca.valueExpr != nil {
					out, err := ca.valueExpr.Eval(env)
					if err != nil {
						return nil, fmt.Errorf("rules: eval value for %q in %q: %w",
							ca.def.Name, rule.def.Name, err)
					}
					s, ok := out.(string)
					if !ok {
						s = fmt.Sprint(out)
					}
					if s == "" {
						continue // skip empty dynamic values
					}
					entry.Value = s
				} else {
					continue // nil expression, skip
				}
			}

			actionSet.Add(entry)
			ruleActions = append(ruleActions, ResolvedAction{
				Name:  ca.def.Name,
				Value: entry.Value,
			})
		}

		matched = append(matched, MatchedRule{
			Name:    rule.def.Name,
			Tags:    rule.def.Tags,
			Actions: ruleActions,
		})

		if rule.stop {
			stopped = true
			stoppedBy = rule.def.Name
			break
		}
	}

	resolved := actionSet.Resolve()
	result := &Result{
		Matched:   matched,
		Actions:   resolved,
		Stopped:   stopped,
		StoppedBy: stoppedBy,
	}

	// Execute handlers.
	if !dry {
		hCtx := &Context[T, V]{
			Env:     env,
			Actions: resolved,
			Vars:    vars,
		}

		var handlerErrors []error
		for _, name := range resolved.Handlers {
			h, ok := e.handlers[name]
			if !ok {
				continue
			}
			if err := h(hCtx); err != nil {
				handlerErrors = append(handlerErrors, fmt.Errorf("handler %q: %w", name, err))
			}
		}

		if len(handlerErrors) > 0 {
			return result, errors.Join(handlerErrors...)
		}
	}

	return result, nil
}
