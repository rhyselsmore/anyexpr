package rules

import (
	"fmt"
	"sync"

	"github.com/rhyselsmore/anyexpr"
)

// Registry manages rule definitions and compiles them into Programs
// on demand. It holds the compiler and actions schema, letting
// callers add, update, upsert, and remove rules by name, then
// compile when ready.
//
//   - E is the environment type (e.g. Email).
//   - A is the actions struct (e.g. EmailActions).
//
// Safe for concurrent use.
type Registry[E any, A any] struct {
	mu       sync.RWMutex
	compiler *anyexpr.Compiler[E]
	actions  *Actions[E, A]
	defs     map[string]Definition
	order    []string // insertion order for deterministic compilation
}

// NewRegistry creates a Registry with the given compiler and actions
// schema.
func NewRegistry[E any, A any](
	compiler *anyexpr.Compiler[E],
	actions *Actions[E, A],
) (*Registry[E, A], error) {
	if actions.IsZero() {
		return nil, ErrActionsZero
	}
	return &Registry[E, A]{
		compiler: compiler,
		actions:  actions,
		defs:     make(map[string]Definition),
	}, nil
}

// Add registers one or more definitions. Returns an error if any
// definition name is already registered.
func (r *Registry[E, A]) Add(defs ...Definition) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, def := range defs {
		if _, exists := r.defs[def.Name]; exists {
			return fmt.Errorf("%w: %q", ErrDefinitionDuplicate, def.Name)
		}
		r.defs[def.Name] = def
		r.order = append(r.order, def.Name)
	}
	return nil
}

// Update replaces one or more existing definitions. Returns an error
// if any definition name is not registered.
func (r *Registry[E, A]) Update(defs ...Definition) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, def := range defs {
		if _, exists := r.defs[def.Name]; !exists {
			return fmt.Errorf("%w: %q does not exist", ErrUnknownDefinition, def.Name)
		}
		r.defs[def.Name] = def
	}
	return nil
}

// Upsert adds or updates one or more definitions. If a definition
// name exists, it is replaced. If it does not exist, it is added.
func (r *Registry[E, A]) Upsert(defs ...Definition) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, def := range defs {
		if _, exists := r.defs[def.Name]; !exists {
			r.order = append(r.order, def.Name)
		}
		r.defs[def.Name] = def
	}
}

// Remove deletes one or more definitions by name. Unknown names are
// silently ignored.
func (r *Registry[E, A]) Remove(names ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, name := range names {
		delete(r.defs, name)
	}
	// Rebuild order without removed names.
	cleaned := make([]string, 0, len(r.order))
	for _, name := range r.order {
		if _, exists := r.defs[name]; exists {
			cleaned = append(cleaned, name)
		}
	}
	r.order = cleaned
}

// Definitions returns a copy of all registered definitions in
// insertion order.
func (r *Registry[E, A]) Definitions() []Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]Definition, 0, len(r.order))
	for _, name := range r.order {
		if def, ok := r.defs[name]; ok {
			defs = append(defs, def)
		}
	}
	return defs
}

// Len returns the number of registered definitions.
func (r *Registry[E, A]) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.defs)
}

// Compile compiles all registered definitions into a Program.
// Returns an error if there are no definitions or if compilation
// fails.
func (r *Registry[E, A]) Compile(opts ...CompileOpt[E, A]) (*Program[E, A], error) {
	defs := r.Definitions()
	return Compile(r.compiler, r.actions, defs, opts...)
}
