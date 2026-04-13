package rules

import (
	"fmt"

	"github.com/rhyselsmore/anyexpr/rules2/action"
)

type actionBinding[V action.Valuable, E any] struct {
	tags  []string
	value V
}

// Trigger records a single action triggered by a matched rule.
type Trigger[V action.Valuable] struct {
	Rule  string
	Tags  []string
	Value V
}

// Action is a typed action field within an actions struct.
//
//   - V is the value type (string, bool, int, float64, or NoArgs),
//     constrained by action.Valuable.
//   - E is the environment type (e.g. Email).
type Action[V action.Valuable, E any] struct {
	definition action.Definition[V]
	bindings   map[string][]actionBinding[V, E]
	index      int // field index

	// Triggered is true if any rule set this action.
	Triggered bool

	// Value is the resolved value. For Single cardinality, last wins.
	// For Multi, the last trigger's value.
	Value V

	// Values holds all resolved values. Populated for Multi cardinality
	// (deduped). For Single, contains zero or one element.
	Values []V

	// Triggers holds the full provenance — every rule that set this
	// action, with its tags and value.
	Triggers []Trigger[V]
}

func (b *Action[V, E]) define(name string, description string, cardinality action.Cardinality, terminal bool) error {
	opts := []action.DefinitionOpt[V]{
		action.WithCardinality[V](cardinality),
		action.Terminal[V](terminal),
	}
	if description != "" {
		opts = append(opts, action.WithDescription[V](description))
	}
	def, err := action.Define(name, opts...)
	if err != nil {
		return err
	}
	b.definition = def
	b.bindings = make(map[string][]actionBinding[V, E])
	return nil
}

func (b *Action[V, E]) bind(ruleName string, ruleTags []string, v any) (bool, action.Cardinality, error) {
	var val V

	// For NoArgs actions, accept nil or NoArgs{}.
	if v == nil {
		if _, isNoArgs := any(val).(action.NoArgs); !isNoArgs {
			return false, 0, fmt.Errorf("%w: action %q: expected %T, got nil", ErrActionValueType, b.definition.Name(), val)
		}
	} else {
		var ok bool
		val, ok = v.(V)
		if !ok {
			return false, 0, fmt.Errorf("%w: action %q: expected %T, got %T", ErrActionValueType, b.definition.Name(), val, v)
		}
	}

	if _, ok := b.bindings[ruleName]; !ok {
		b.bindings[ruleName] = make([]actionBinding[V, E], 0)
	}

	b.bindings[ruleName] = append(b.bindings[ruleName], actionBinding[V, E]{
		tags:  ruleTags,
		value: val,
	})

	return b.definition.Terminal(), b.definition.Cardinality(), nil
}

func (b *Action[V, E]) trigger(matched []string) {
	for _, rule := range matched {
		entries, ok := b.bindings[rule]
		if !ok {
			continue
		}
		b.Triggered = true
		for _, entry := range entries {
			b.Triggers = append(b.Triggers, Trigger[V]{
				Rule:  rule,
				Tags:  entry.tags,
				Value: entry.value,
			})
		}
	}

	if !b.Triggered {
		return
	}

	// Last value wins for Value — works for both Single and Multi.
	b.Value = b.Triggers[len(b.Triggers)-1].Value

	// Build Values — for Multi, collect all and dedup. For Single,
	// just the winning value.
	if b.definition.Cardinality() == action.Multi {
		seen := make(map[any]bool)
		for _, t := range b.Triggers {
			key := any(t.Value)
			if !seen[key] {
				seen[key] = true
				b.Values = append(b.Values, t.Value)
			}
		}
	} else {
		b.Values = []V{b.Value}
	}
}

// ActionInfo is the type-erased metadata for a defined action,
// returned by Actions.Describe.
type ActionInfo struct {
	// Name is the action's registered name from the struct tag.
	Name string

	// Description is the human-readable description from the
	// `description` struct tag, if present.
	Description string

	// Cardinality is Single or Multi.
	Cardinality action.Cardinality

	// Terminal is true if triggering this action halts evaluation.
	Terminal bool

	// ValueType is the Go type name of the value (e.g. "string", "bool").
	ValueType string
}

func (b *Action[V, E]) describe() ActionInfo {
	var zero V
	return ActionInfo{
		Name:        b.definition.Name(),
		Description: b.definition.Description(),
		Cardinality: b.definition.Cardinality(),
		Terminal:    b.definition.Terminal(),
		ValueType:   fmt.Sprintf("%T", zero),
	}
}

type actionTriggerable[E any] interface {
	trigger([]string)
}

type actionBinder[E any] interface {
	bind(ruleName string, ruleTags []string, value any) (bool, action.Cardinality, error)
}

type actionDescriber interface {
	describe() ActionInfo
}

type actionDefiner[E any] interface {
	actionBinder[E]
	actionDescriber
	define(name string, description string, cardinality action.Cardinality, terminal bool) error
}
