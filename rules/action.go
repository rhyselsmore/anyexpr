package rules

import (
	"fmt"
	"reflect"
)

// Actionable constrains the types that Action[T] can hold.
// These map to JSON primitives: string, boolean, number, and null/omit.
type Actionable interface {
	string | bool | int | float64 | NoArgs
}

// Cardinality defines how an action accumulates values across rules.
type Cardinality int

const (
	// Single means at most once per rule. Across rules, last match wins.
	Single Cardinality = iota
	// Multi may appear multiple times. All values accumulate, duplicates
	// stripped.
	Multi
)

func (c Cardinality) String() string {
	switch c {
	case Single:
		return "single"
	case Multi:
		return "multi"
	default:
		return fmt.Sprintf("Cardinality(%d)", int(c))
	}
}

// entry stores a single value alongside the rule that produced it.
type entry[T Actionable] struct {
	value    T
	ruleName string
	ruleTags []string
}

// Action is a typed action field within an actions struct.
//
// T is the value type this action carries (string, bool, int, float64,
// or NoArgs), constrained by Actionable.
//
// E is the environment type — the struct that expressions are evaluated
// against (e.g. Email, Transaction). E flows through from the actions
// struct declaration, tying each action to the type it operates on.
type Action[T Actionable, E any] struct {
	// Schema fields — set by DefineActions, carried through copies.
	name        string
	cardinality Cardinality
	terminal    bool
	index       int
	configured  bool

	// Value fields — set during evaluation.
	entries []entry[T]
}

// --- Public accessors ---

// Name returns the action's registered name from the struct tag.
func (a *Action[T, E]) Name() string { return a.name }

// Value returns the resolved value and whether it was set. For Single
// actions, returns the winning (last) value. For Multi, returns the
// last value. For NoArgs, returns the zero value; use Fired instead.
func (a *Action[T, E]) Value() (T, bool) {
	if len(a.entries) == 0 {
		var zero T
		return zero, false
	}
	return a.entries[len(a.entries)-1].value, true
}

// Values returns all resolved values. For Multi actions, values are
// deduped. For Single, returns zero or one element. Returns a non-nil
// empty slice if the action was not triggered.
func (a *Action[T, E]) Values() []T {
	if len(a.entries) == 0 {
		return []T{}
	}
	result := make([]T, len(a.entries))
	for i, e := range a.entries {
		result[i] = e.value
	}
	return result
}

// Fired returns true if this action was triggered by any rule.
func (a *Action[T, E]) Fired() bool {
	return len(a.entries) > 0
}

// ByRule returns values contributed by the named rule. Returns a
// non-nil empty slice if no values from that rule.
func (a *Action[T, E]) ByRule(name string) []T {
	var result []T
	for _, e := range a.entries {
		if e.ruleName == name {
			result = append(result, e.value)
		}
	}
	if result == nil {
		return []T{}
	}
	return result
}

// ByTag returns values contributed by rules with the given tag.
// Returns a non-nil empty slice if no matching rules.
func (a *Action[T, E]) ByTag(tag string) []T {
	var result []T
	for _, e := range a.entries {
		for _, t := range e.ruleTags {
			if t == tag {
				result = append(result, e.value)
				break
			}
		}
	}
	if result == nil {
		return []T{}
	}
	return result
}

// Rules returns the names of rules that triggered this action, in
// evaluation order, deduped.
func (a *Action[T, E]) Rules() []string {
	if len(a.entries) == 0 {
		return []string{}
	}
	seen := make(map[string]bool)
	var result []string
	for _, e := range a.entries {
		if !seen[e.ruleName] {
			seen[e.ruleName] = true
			result = append(result, e.ruleName)
		}
	}
	return result
}

// --- Unexported interfaces ---

// actionField is implemented by *Action[T, E] for all T. Used by
// DefineActions to configure actions and by Compile to type-check
// values — all at init/compile time, not in the hot path.
type actionField[E any] interface {
	configure(name string, c Cardinality, terminal bool, index int)
	compile(v any) (actionValuer[E], error)
}

// actionResolver is implemented by *Action[T, E]. Used by the
// evaluator after all rules have been processed to apply dedup
// and last-wins semantics.
type actionResolver interface {
	resolve()
}

// --- actionField implementation ---

func (a *Action[T, E]) configure(name string, c Cardinality, terminal bool, index int) {
	a.name = name
	a.cardinality = c
	a.terminal = terminal
	a.index = index
	a.configured = true
}

func (a *Action[T, E]) compile(v any) (actionValuer[E], error) {
	val, ok := v.(T)
	if !ok {
		var zero T
		return nil, fmt.Errorf("%w: action %q: expected %T, got %T", ErrValueType, a.name, zero, v)
	}
	return &actionValue[T, E]{
		action: *a, // value copy — name, cardinality, terminal, index carry over; entries nil
		value:  val,
	}, nil
}

func (a *Action[T, E]) resolve() {
	if len(a.entries) == 0 {
		return
	}
	if a.cardinality == Multi {
		seen := make(map[any]bool)
		deduped := make([]entry[T], 0, len(a.entries))
		for _, e := range a.entries {
			key := any(e.value)
			if !seen[key] {
				seen[key] = true
				deduped = append(deduped, e)
			}
		}
		a.entries = deduped
	}
	// Single: keep all entries for provenance. Value() returns the last.
}

// --- actionValuer (compiled action with typed value) ---

// actionValuer is a compiled action entry — a copy of the Action with
// a typed value, ready to be applied during evaluation. Stored in
// compiledRule as an interface to erase T.
type actionValuer[E any] interface {
	actionName() string
	actionCardinality() Cardinality
	actionTerminal() bool
	// addEntry writes an entry to the corresponding Action field on the
	// schema copy pointed to by ptr, using the stored field index.
	addEntry(ptr reflect.Value, ruleName string, ruleTags []string)
	// stringValue returns the value as a string for audit/display.
	stringValue() string
}

type actionValue[T Actionable, E any] struct {
	action Action[T, E] // value copy from compile — has index, metadata, nil entries
	value  T
}

func (av *actionValue[T, E]) actionName() string             { return av.action.name }
func (av *actionValue[T, E]) actionCardinality() Cardinality { return av.action.cardinality }
func (av *actionValue[T, E]) actionTerminal() bool           { return av.action.terminal }
func (av *actionValue[T, E]) stringValue() string            { return fmt.Sprint(av.value) }

func (av *actionValue[T, E]) addEntry(ptr reflect.Value, ruleName string, ruleTags []string) {
	field := ptr.Field(av.action.index).Addr().Interface().(*Action[T, E])
	field.entries = append(field.entries, entry[T]{
		value:    av.value,
		ruleName: ruleName,
		ruleTags: ruleTags,
	})
}
