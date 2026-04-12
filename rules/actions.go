package rules

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// Actions holds the action schema for type A. Created once via
// DefineActions and passed to Compile and NewEvaluator. Safe for
// concurrent use after construction.
//
// A is the actions struct containing Action[T, E] fields.
// E is the environment type — the struct that expressions are
// evaluated against (e.g. Email, Transaction).
type Actions[A any, E any] struct {
	schema    A
	fields    map[string]int
	compilers map[string]actionField[E]
	defined   bool
}

// DefineActions reflects over A to build the action schema.
//
// A is the actions struct containing Action[T, E] fields.
// E is the environment type — the struct that expressions are
// evaluated against. It threads through to each Action field,
// tying the actions to the type they operate on.
//
// It walks exported fields of A, looking for Action[T, E] types with
// a `rule` struct tag. Each tag is parsed for the action name and
// options (multi, terminal). The Action fields on the internal
// instance are configured with their name, cardinality, terminal
// flag, and value kind (inferred from T via the Actionable
// constraint).
//
// The returned *Actions[A, E] holds the configured instance. It is
// the seed used by Compile (for validation) and NewEvaluator (which
// copies it per evaluation to populate values).
func DefineActions[A any, E any]() (*Actions[A, E], error) {
	var schema A
	v := reflect.ValueOf(&schema).Elem()
	t := v.Type()

	compilers := make(map[string]actionField[E])
	indexes := make(map[string]int)
	names := make(map[string]string) // action name → field name (for errors)
	terminalCount := 0

	for i := 0; i < t.NumField(); i++ {
		idx := i
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		fieldPtr := v.Field(i).Addr().Interface()
		af, ok := fieldPtr.(actionField[E])
		if !ok {
			continue
		}

		tag, hasTag := f.Tag.Lookup("rule")
		if !hasTag {
			return nil, fmt.Errorf("%w: field %q has no rule tag", ErrDefine, f.Name)
		}

		name, cardinality, terminal, err := parseTag(tag)
		if err != nil {
			return nil, fmt.Errorf("%w: field %q: %v", ErrDefine, f.Name, err)
		}

		if !isValidName(name) {
			return nil, fmt.Errorf("%w: field %q: name %q is not a valid identifier", ErrDefine, f.Name, name)
		}

		if prev, exists := names[name]; exists {
			return nil, fmt.Errorf("%w: name %q used by both %q and %q", ErrDuplicateRegistration, name, prev, f.Name)
		}
		names[name] = f.Name

		if terminal {
			terminalCount++
			if terminalCount > 1 {
				return nil, fmt.Errorf("%w: field %q", ErrMultipleTerminals, f.Name)
			}
		}

		af.configure(name, cardinality, terminal, idx)
		indexes[name] = idx
		compilers[name] = af
	}

	if len(compilers) == 0 {
		return nil, fmt.Errorf("%w: no Action fields found in %T", ErrDefine, schema)
	}

	return &Actions[A, E]{
		schema:    schema,
		fields:    indexes,
		compilers: compilers,
		defined:   true,
	}, nil
}

// parseTag parses a rule struct tag like "label,multi" or "delete,terminal".
func parseTag(tag string) (name string, c Cardinality, terminal bool, err error) {
	parts := strings.Split(tag, ",")
	if len(parts) == 0 || parts[0] == "" {
		return "", 0, false, fmt.Errorf("empty tag")
	}

	name = strings.TrimSpace(parts[0])
	c = Single

	for _, opt := range parts[1:] {
		switch strings.TrimSpace(opt) {
		case "multi":
			c = Multi
		case "terminal":
			terminal = true
		default:
			return "", 0, false, fmt.Errorf("unknown option %q", opt)
		}
	}

	return name, c, terminal, nil
}

// isValidName checks that a name is a valid identifier (letter/underscore
// start, alphanumeric/underscore/hyphen body).
func isValidName(name string) bool {
	if name == "" {
		return false
	}
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
