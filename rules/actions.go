package rules

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/rhyselsmore/anyexpr/rules2/action"
)

// Actions holds the action schema for type A, bound to environment
// type E. Created via DefineActions.
//
//   - E is the environment type — the struct that expressions evaluate
//     against (e.g. Email, Transaction).
//   - A is the actions struct containing Action[V, E] fields with
//     `rule` tags.
type Actions[E any, A any] struct {
	schema  A
	fields  map[string]int
	binders map[string]actionBinder[E]
	infos   []ActionInfo
	defined bool
}

// IsZero returns true if the registry was not created via DefineActions.
func (ac *Actions[E, A]) IsZero() bool { return ac == nil || !ac.defined }

// Describe returns metadata for all defined actions, in struct field
// order. Useful for introspection — an agent or UI can discover what
// actions are available, their types, and descriptions.
func (ac *Actions[E, A]) Describe() []ActionInfo { return ac.infos }

// DefineActions reflects over A to build the action schema.
//
//   - E is the environment type (e.g. Email).
//   - A is the actions struct (e.g. EmailActions).
//
// It walks exported fields of A, looking for Action[V, E] types with
// a `rule` struct tag. Each field is configured with its name,
// cardinality, and terminal flag parsed from the tag. Values are
// bound at compile time via Compile.
func DefineActions[E any, A any]() (*Actions[E, A], error) {
	var schema A
	v := reflect.ValueOf(&schema).Elem()
	t := v.Type()

	names := make(map[string]string) // action name → field name (for errors)
	fields := make(map[string]int)   // action name → field index (for setting)
	binders := make(map[string]actionBinder[E])
	var infos []ActionInfo

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		// Get the Field Pointer
		fieldPtr := v.Field(i).Addr().Interface()
		bf, ok := fieldPtr.(actionDefiner[E])
		if !ok {
			continue
		}

		// Get Tags, Define Action
		tag, hasTag := f.Tag.Lookup("rule")
		if !hasTag {
			return nil, fmt.Errorf("%w: field %q has no rule tag", ErrDefine, f.Name)
		}

		name, cardinality, terminal, err := parseTag(tag)
		if err != nil {
			return nil, fmt.Errorf("%w: field %q: %v", ErrDefine, f.Name, err)
		}

		// Description tag (optional)
		description, _ := f.Tag.Lookup("description")

		if err := bf.define(name, description, cardinality, terminal); err != nil {
			return nil, err
		}

		// Name Uniqueness
		if prev, exists := names[name]; exists {
			return nil, fmt.Errorf("%w: name %q used by both %q and %q", ErrDuplicateRegistration, name, prev, f.Name)
		}
		names[name] = f.Name

		// Set Index
		fields[name] = i
		binders[name] = bf
		infos = append(infos, bf.describe())
	}

	if len(binders) == 0 {
		return nil, fmt.Errorf("%w: no Action fields found in %T", ErrDefine, schema)
	}

	return &Actions[E, A]{
		schema:  schema,
		fields:  fields,
		binders: binders,
		infos:   infos,
		defined: true,
	}, nil
}

// parseTag parses a rule struct tag like "label,multi" or "delete,terminal".
func parseTag(tag string) (name string, c action.Cardinality, terminal bool, err error) {
	parts := strings.Split(tag, ",")
	if len(parts) == 0 || parts[0] == "" {
		return "", 0, false, fmt.Errorf("%w: tag is empty", ErrDefine)
	}

	name = strings.TrimSpace(parts[0])
	c = action.Single

	for _, opt := range parts[1:] {
		opt = strings.TrimSpace(opt)
		if opt == "terminal" {
			terminal = true
			continue
		}
		if parsed, parseErr := action.ParseCardinality(opt); parseErr == nil {
			c = parsed
			continue
		}
		return "", 0, false, fmt.Errorf("%w: unknown tag option %q", ErrDefine, opt)
	}

	return name, c, terminal, nil
}
