package rules

import "github.com/rhyselsmore/anyexpr"

type selector struct {
	onlyTags     map[string]bool
	onlyNames    map[string]bool
	excludeTags  map[string]bool
	excludeNames map[string]bool
	exprFilter   *anyexpr.Program[RuleMeta]
}

// RuleMeta is the struct that expression-based selectors evaluate
// against. It exposes rule metadata as fields that can be referenced
// in selector expressions.
type RuleMeta struct {
	Name string
	Tags []string
}

// includes returns whether a rule with the given name and tags passes
// the selector's filters.
//
// If onlyTags is non-empty, the rule must have at least one matching tag.
// If onlyNames is non-empty, the rule name must be in the set.
// If both are set, either match suffices (union of includes).
// Excludes take priority over includes.
// If an expression filter is set, it is evaluated last.
func (s selector) includes(name string, tags []string) bool {
	if s.excludeNames[name] {
		return false
	}
	for _, t := range tags {
		if s.excludeTags[t] {
			return false
		}
	}

	hasIncludes := len(s.onlyTags) > 0 || len(s.onlyNames) > 0
	if hasIncludes {
		matched := false
		if s.onlyNames[name] {
			matched = true
		}
		if !matched {
			for _, t := range tags {
				if s.onlyTags[t] {
					matched = true
					break
				}
			}
		}
		if !matched {
			return false
		}
	}

	if s.exprFilter != nil {
		ok, err := s.exprFilter.Match(RuleMeta{Name: name, Tags: tags})
		if err != nil || !ok {
			return false
		}
	}

	return true
}
