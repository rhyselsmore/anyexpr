package rules

type selector struct {
	onlyTags     map[string]bool
	onlyNames    map[string]bool
	excludeTags  map[string]bool
	excludeNames map[string]bool
}

// includes returns whether a rule with the given name and tags passes
// the selector's filters.
//
// If onlyTags is non-empty, the rule must have at least one matching tag.
// If onlyNames is non-empty, the rule name must be in the set.
// If both are set, either match suffices (union of includes).
// Excludes take priority over includes.
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
	if !hasIncludes {
		return true
	}

	if s.onlyNames[name] {
		return true
	}

	for _, t := range tags {
		if s.onlyTags[t] {
			return true
		}
	}

	return false
}

func mergeSets(a, b map[string]bool) map[string]bool {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	out := make(map[string]bool, len(a)+len(b))
	for k := range a {
		out[k] = true
	}
	for k := range b {
		out[k] = true
	}
	return out
}

func copySets(m map[string]bool) map[string]bool {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]bool, len(m))
	for k := range m {
		out[k] = true
	}
	return out
}
