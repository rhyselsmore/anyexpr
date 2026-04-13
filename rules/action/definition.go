package action

import (
	"fmt"
	"unicode"
)

// NoArgs is the type parameter for actions that carry no value.
// Use it for presence-only actions like "delete" or "archive".
type NoArgs struct{}

// Valuable constrains the types that Definition[T] can hold.
// These map to JSON primitives: string, boolean, number, and null/omit.
type Valuable interface {
	string | bool | int | float64 | NoArgs
}

// DefinitionOpt configures a Definition.
type DefinitionOpt[V Valuable] func(*Definition[V])

// Terminal marks the action as terminal — evaluation halts when it
// fires.
func Terminal[V Valuable](v bool) DefinitionOpt[V] {
	return func(d *Definition[V]) {
		d.terminal = v
	}
}

// WithMulti sets the cardinality to Multi. Shorthand for
// WithCardinality(Multi).
func WithMulti[V Valuable](d *Definition[V]) {
	d.cardinality = Multi
}

// WithDescription sets an optional human-readable description. Must be
// 255 characters or fewer.
func WithDescription[V Valuable](s string) DefinitionOpt[V] {
	return func(d *Definition[V]) {
		d.description = s
	}
}

// WithCardinality sets how values accumulate — Single (last wins) or
// Multi (all values collected, deduped).
func WithCardinality[V Valuable](c Cardinality) DefinitionOpt[V] {
	return func(d *Definition[V]) {
		d.cardinality = c
	}
}

// Definition is the validated metadata for an action. Created via
// Define or MustDefine.
type Definition[V Valuable] struct {
	name        string
	description string
	cardinality Cardinality
	terminal    bool
}

// IsZero returns true if the definition was not created via Define or
// MustDefine.
func (d Definition[V]) IsZero() bool { return d.name == "" }

// Name returns the action's name.
func (d Definition[V]) Name() string { return d.name }

// Description returns the action's description.
func (d Definition[V]) Description() string { return d.description }

// Terminal returns whether the action is terminal.
func (d Definition[V]) Terminal() bool { return d.terminal }

// Cardinality returns the action's cardinality.
func (d Definition[V]) Cardinality() Cardinality { return d.cardinality }

// Define creates a validated Definition. Name must be non-empty and a
// valid JSON key (letters, digits, underscores, hyphens; must start
// with a letter or underscore). Description, if present, must be 255
// characters or fewer.
func Define[V Valuable](name string, opts ...DefinitionOpt[V]) (Definition[V], error) {
	d := Definition[V]{name: name}
	for _, o := range opts {
		o(&d)
	}

	if err := d.validate(); err != nil {
		return Definition[V]{}, err
	}

	return d, nil
}

// MustDefine is like Define but panics on error.
func MustDefine[V Valuable](name string, opts ...DefinitionOpt[V]) Definition[V] {
	d, err := Define[V](name, opts...)
	if err != nil {
		panic(fmt.Sprintf("action.MustDefine: %v", err))
	}
	return d
}

func (d Definition[V]) validate() error {
	if d.name == "" {
		return ErrNameEmpty
	}
	if !isValidName(d.name) {
		return fmt.Errorf("%w: %q", ErrNameInvalid, d.name)
	}
	if err := d.cardinality.Validate(); err != nil {
		return err
	}
	if len(d.description) > 255 {
		return fmt.Errorf("%w: got %d characters", ErrDescriptionTooLong, len(d.description))
	}
	return nil
}

// isValidName checks that a name starts with a letter or underscore,
// followed by letters, digits, underscores, or hyphens.
func isValidName(name string) bool {
	for i, r := range name {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
		} else {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
				return false
			}
		}
	}
	return true
}
