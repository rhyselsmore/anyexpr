// Package dispatch routes evaluation results to named handler
// functions. Handlers are registered immutably on a [Dispatcher],
// then composed into [Plan]s that specify which handlers run and
// under what conditions.
//
// [rules]: github.com/rhyselsmore/anyexpr/rules
package dispatch

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/rhyselsmore/anyexpr"
	rules "github.com/rhyselsmore/anyexpr/rules"
)

// HandlerFunc is the function signature for dispatch handlers.
type HandlerFunc[E any, A any] func(ctx context.Context, eval *rules.Evaluation[E, A]) error

// --- Dispatcher (handler registry) ---

// DispatcherOpt configures a Dispatcher.
type DispatcherOpt[E any, A any] func(*dispatcherConfig[E, A]) error

type dispatcherConfig[E any, A any] struct {
	logger   *slog.Logger
	handlers []namedHandler[E, A]
}

type namedHandler[E any, A any] struct {
	name        string
	description string
	fn          HandlerFunc[E, A]
}

// HandleOpt configures a handler registration.
type HandleOpt func(*handleConfig)

type handleConfig struct {
	description string
}

// WithDescription sets a human-readable description for a handler.
// Surfaced via Dispatcher.Describe for introspection.
func WithDescription(desc string) HandleOpt {
	return func(c *handleConfig) { c.description = desc }
}

// WithLogger sets a structured logger for dispatch events.
func WithLogger[E any, A any](l *slog.Logger) DispatcherOpt[E, A] {
	return func(c *dispatcherConfig[E, A]) error {
		c.logger = l
		return nil
	}
}

// Handle registers a named handler function with optional
// description.
func Handle[E any, A any](name string, fn HandlerFunc[E, A], opts ...HandleOpt) DispatcherOpt[E, A] {
	return func(c *dispatcherConfig[E, A]) error {
		for _, h := range c.handlers {
			if h.name == name {
				return fmt.Errorf("%w: %q", ErrDuplicateHandler, name)
			}
		}
		hc := &handleConfig{}
		for _, o := range opts {
			o(hc)
		}
		c.handlers = append(c.handlers, namedHandler[E, A]{
			name:        name,
			description: hc.description,
			fn:          fn,
		})
		return nil
	}
}

// Dispatcher is an immutable registry of named handlers. Build plans
// from it to control which handlers run and when.
//
//   - E is the environment type (e.g. Email).
//   - A is the actions struct (e.g. EmailActions).
type Dispatcher[E any, A any] struct {
	compiler     *anyexpr.Compiler[rules.Evaluation[E, A]]
	handlers     map[string]HandlerFunc[E, A]
	descriptions map[string]string
	order        []string // registration order
	logger       *slog.Logger
}

// HandlerInfo describes a registered handler.
type HandlerInfo struct {
	// Name is the handler's registered name.
	Name string
	// Description is the human-readable description, if provided.
	Description string
}

// Describe returns metadata for all registered handlers, in
// registration order. Useful for introspection — an agent can
// discover what handlers are available.
func (d *Dispatcher[E, A]) Describe() []HandlerInfo {
	infos := make([]HandlerInfo, len(d.order))
	for i, name := range d.order {
		infos[i] = HandlerInfo{
			Name:        name,
			Description: d.descriptions[name],
		}
	}
	return infos
}

// New creates an immutable Dispatcher with the given handlers.
func New[E any, A any](opts ...DispatcherOpt[E, A]) (*Dispatcher[E, A], error) {
	cfg := &dispatcherConfig[E, A]{}
	for _, o := range opts {
		if err := o(cfg); err != nil {
			return nil, err
		}
	}

	compiler, err := anyexpr.NewCompiler[rules.Evaluation[E, A]]()
	if err != nil {
		return nil, fmt.Errorf("dispatch: compiler: %w", err)
	}

	handlers := make(map[string]HandlerFunc[E, A], len(cfg.handlers))
	descriptions := make(map[string]string, len(cfg.handlers))
	order := make([]string, 0, len(cfg.handlers))
	for _, h := range cfg.handlers {
		handlers[h.name] = h.fn
		descriptions[h.name] = h.description
		order = append(order, h.name)
	}

	return &Dispatcher[E, A]{
		compiler:     compiler,
		handlers:     handlers,
		descriptions: descriptions,
		order:        order,
		logger:       cfg.logger,
	}, nil
}

// --- Plan ---

// PlanOpt configures a Plan.
type PlanOpt[E any, A any] func(*planConfig[E, A]) error

type planConfig[E any, A any] struct {
	name     string
	strategy Strategy
	gate     *whenBinding[E, A]
	entries  []planEntry[E, A]
}

// WithName sets a name for the plan. Used in logging and Debug output.
func WithName[E any, A any](name string) PlanOpt[E, A] {
	return func(c *planConfig[E, A]) error {
		c.name = name
		return nil
	}
}

type planEntry[E any, A any] struct {
	name  string
	fn    HandlerFunc[E, A]
	whens []whenBinding[E, A]
}

type whenBinding[E any, A any] struct {
	expr string
	prog *anyexpr.Program[rules.Evaluation[E, A]]
}

// WithStrategy sets the execution strategy for the plan. Default is
// AllContinue.
func WithStrategy[E any, A any](s Strategy) PlanOpt[E, A] {
	return func(c *planConfig[E, A]) error {
		c.strategy = s
		return nil
	}
}

// RunOpt configures a handler entry within a plan.
type RunOpt[E any, A any] func(*runConfig[E, A])

type runConfig[E any, A any] struct {
	whens []string
}

// When adds a gating expression to a handler within a plan. The
// expression is evaluated against Evaluation[E, A]. The handler
// runs if any of its When expressions match. No When means always
// run.
func When[E any, A any](expr string) RunOpt[E, A] {
	return func(c *runConfig[E, A]) {
		c.whens = append(c.whens, expr)
	}
}

// Run includes a named handler in the plan with optional When
// expressions.
func Run[E any, A any](name string, opts ...RunOpt[E, A]) PlanOpt[E, A] {
	return func(c *planConfig[E, A]) error {
		rc := &runConfig[E, A]{}
		for _, o := range opts {
			o(rc)
		}
		c.entries = append(c.entries, planEntry[E, A]{
			name: name,
			// fn and whens populated during Plan() when we have the compiler
		})
		// stash the raw exprs for later compilation
		c.entries[len(c.entries)-1].whens = make([]whenBinding[E, A], 0, len(rc.whens))
		for _, expr := range rc.whens {
			c.entries[len(c.entries)-1].whens = append(c.entries[len(c.entries)-1].whens, whenBinding[E, A]{expr: expr})
		}
		return nil
	}
}

// Gate sets a top-level expression that must pass before any handlers
// execute. If the gate expression returns false, no handlers run.
func Gate[E any, A any](expr string) PlanOpt[E, A] {
	return func(c *planConfig[E, A]) error {
		c.gate = &whenBinding[E, A]{expr: expr}
		return nil
	}
}

// Plan is an immutable, compiled dispatch plan. Created from a
// Dispatcher via Plan(). Safe for concurrent use.
type Plan[E any, A any] struct {
	name     string
	entries  []planEntry[E, A]
	gate     *whenBinding[E, A]
	strategy Strategy
	logger   *slog.Logger
}

// PlanInfo describes a handler entry within a plan.
type PlanInfo struct {
	// Handler is the handler name.
	Handler string
	// Whens lists the gating expressions for this handler. Empty
	// means always runs.
	Whens []string
}

// Name returns the plan's name.
func (p *Plan[E, A]) Name() string { return p.name }

// Describe returns metadata for each handler in the plan, in
// execution order.
func (p *Plan[E, A]) Describe() []PlanInfo {
	infos := make([]PlanInfo, len(p.entries))
	for i, e := range p.entries {
		whens := make([]string, len(e.whens))
		for j, w := range e.whens {
			whens[j] = w.expr
		}
		infos[i] = PlanInfo{
			Handler: e.name,
			Whens:   whens,
		}
	}
	return infos
}

// Plan builds an immutable plan from the dispatcher's registered
// handlers. Handler names are validated, When expressions are compiled.
func (d *Dispatcher[E, A]) Plan(opts ...PlanOpt[E, A]) (*Plan[E, A], error) {
	cfg := &planConfig[E, A]{strategy: AllContinue}
	for _, o := range opts {
		if err := o(cfg); err != nil {
			return nil, err
		}
	}

	entries := make([]planEntry[E, A], 0, len(cfg.entries))
	for _, e := range cfg.entries {
		fn, ok := d.handlers[e.name]
		if !ok {
			return nil, fmt.Errorf("%w: %q", ErrUnknownHandler, e.name)
		}

		compiled := make([]whenBinding[E, A], 0, len(e.whens))
		for _, w := range e.whens {
			prog, err := d.compiler.Compile(anyexpr.NewSource(e.name+"/when", w.expr))
			if err != nil {
				return nil, fmt.Errorf("dispatch: handler %q when %q: %w", e.name, w.expr, err)
			}
			compiled = append(compiled, whenBinding[E, A]{expr: w.expr, prog: prog})
		}

		entries = append(entries, planEntry[E, A]{
			name:  e.name,
			fn:    fn,
			whens: compiled,
		})
	}

	var gate *whenBinding[E, A]
	if cfg.gate != nil {
		prog, err := d.compiler.Compile(anyexpr.NewSource("gate", cfg.gate.expr))
		if err != nil {
			return nil, fmt.Errorf("dispatch: gate %q: %w", cfg.gate.expr, err)
		}
		gate = &whenBinding[E, A]{expr: cfg.gate.expr, prog: prog}
	}

	return &Plan[E, A]{
		name:     cfg.name,
		entries:  entries,
		gate:     gate,
		strategy: cfg.strategy,
		logger:   d.logger,
	}, nil
}

// Execute runs the plan against an evaluation result.
func (p *Plan[E, A]) Execute(ctx context.Context, eval *rules.Evaluation[E, A]) *Result[E, A] {
	start := time.Now()
	result := &Result[E, A]{Plan: p.name, Evaluation: eval}

	// Check gate.
	if p.gate != nil {
		result.Gated = true
		result.GateExpr = p.gate.expr

		passed, err := p.gate.prog.Match(*eval)
		if err != nil || !passed {
			if p.logger != nil {
				p.logger.Info("dispatch gate blocked", "expr", p.gate.expr)
			}
			result.GatePassed = false
			result.Duration = time.Since(start)
			return result
		}
		result.GatePassed = true
	}

	// Dispatch handlers.
	for _, e := range p.entries {
		select {
		case <-ctx.Done():
			result.Duration = time.Since(start)
			return result
		default:
		}

		matchedExpr, shouldRun := p.shouldRun(e, eval)
		if !shouldRun {
			if p.logger != nil {
				p.logger.Debug("dispatch skip", "handler", e.name)
			}
			continue
		}

		if p.logger != nil {
			p.logger.Info("dispatch invoke", "handler", e.name, "matched", matchedExpr)
		}

		dispatched := p.invoke(ctx, e, eval, matchedExpr)
		result.Dispatched = append(result.Dispatched, dispatched)

		if dispatched.Error != nil {
			if p.logger != nil {
				p.logger.Warn("dispatch error",
					"handler", e.name,
					"error", dispatched.Error,
					"panicked", dispatched.Panicked,
				)
			}
			if p.strategy == AllHaltOnError {
				result.Duration = time.Since(start)
				return result
			}
		}

		if p.strategy == FirstMatch {
			result.Duration = time.Since(start)
			return result
		}
	}

	result.Duration = time.Since(start)
	return result
}

func (p *Plan[E, A]) shouldRun(e planEntry[E, A], eval *rules.Evaluation[E, A]) (string, bool) {
	if len(e.whens) == 0 {
		return "", true
	}
	for _, w := range e.whens {
		ok, err := w.prog.Match(*eval)
		if err == nil && ok {
			return w.expr, true
		}
	}
	return "", false
}

func (p *Plan[E, A]) invoke(ctx context.Context, e planEntry[E, A], eval *rules.Evaluation[E, A], matchedExpr string) Dispatched {
	start := time.Now()
	dis := Dispatched{
		Handler:     e.name,
		MatchedExpr: matchedExpr,
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				dis.Panicked = true
				dis.Error = fmt.Errorf("dispatch: handler %q panicked: %v", e.name, r)
			}
		}()
		dis.Error = e.fn(ctx, eval)
	}()

	dis.Duration = time.Since(start)
	return dis
}

// --- Errors ---

var (
	// ErrDuplicateHandler is returned when two handlers share the
	// same name.
	ErrDuplicateHandler = errors.New("dispatch: duplicate handler")

	// ErrUnknownHandler is returned when a plan references a handler
	// name not registered on the dispatcher.
	ErrUnknownHandler = errors.New("dispatch: unknown handler")
)
