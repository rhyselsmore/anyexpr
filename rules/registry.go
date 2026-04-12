package rules

import (
	"fmt"
	"sort"
)

// Registry holds action definitions and handler implementations. It is
// immutable after construction and safe for concurrent use.
type Registry struct {
	actions  map[string]Def
	handlers map[string]any
	parent   *Registry
}

// NewRegistry creates a new Registry with the given options.
func NewRegistry(opts ...RegistryOpt) (*Registry, error) {
	cfg := &registryConfig{
		actions:  make(map[string]Def),
		handlers: make(map[string]any),
	}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}
	return &Registry{
		actions:  cfg.actions,
		handlers: cfg.handlers,
	}, nil
}

// With returns a new registry inheriting from this one. The parent is
// not modified. Name collisions with the parent return an error.
func (r *Registry) With(opts ...RegistryOpt) (*Registry, error) {
	cfg := &registryConfig{
		actions:  make(map[string]Def),
		handlers: make(map[string]any),
	}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	// Check for collisions with the parent chain.
	for name := range cfg.actions {
		if _, exists := r.LookupAction(name); exists {
			return nil, fmt.Errorf("%w: %q conflicts with parent", ErrDuplicateRegistration, name)
		}
	}

	return &Registry{
		actions:  cfg.actions,
		handlers: cfg.handlers,
		parent:   r,
	}, nil
}

// LookupAction returns the action definition for the given name,
// checking the local registry first, then walking up the parent chain.
func (r *Registry) LookupAction(name string) (Def, bool) {
	if d, ok := r.actions[name]; ok {
		return d, true
	}
	if r.parent != nil {
		return r.parent.LookupAction(name)
	}
	return Def{}, false
}

// LookupHandler returns the handler function for the given name,
// checking the local registry first, then walking up the parent chain.
func (r *Registry) LookupHandler(name string) (any, bool) {
	if h, ok := r.handlers[name]; ok {
		return h, true
	}
	if r.parent != nil {
		return r.parent.LookupHandler(name)
	}
	return nil, false
}

// ActionNames returns all registered action names, including those from
// parent registries. Names are returned in sorted order.
func (r *Registry) ActionNames() []string {
	names := make(map[string]bool)
	r.collectActionNames(names)
	out := make([]string, 0, len(names))
	for n := range names {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func (r *Registry) collectActionNames(names map[string]bool) {
	for n := range r.actions {
		names[n] = true
	}
	if r.parent != nil {
		r.parent.collectActionNames(names)
	}
}

// HandlerNames returns all registered handler names, including those
// from parent registries. Names are returned in sorted order.
func (r *Registry) HandlerNames() []string {
	names := make(map[string]bool)
	r.collectHandlerNames(names)
	out := make([]string, 0, len(names))
	for n := range names {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func (r *Registry) collectHandlerNames(names map[string]bool) {
	for n := range r.handlers {
		names[n] = true
	}
	if r.parent != nil {
		r.parent.collectHandlerNames(names)
	}
}
