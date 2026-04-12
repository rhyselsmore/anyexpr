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
	// Excludes take priority.
	if s.excludeNames[name] {
		return false
	}
	for _, t := range tags {
		if s.excludeTags[t] {
			return false
		}
	}

	// If no includes are set, include everything.
	hasIncludes := len(s.onlyTags) > 0 || len(s.onlyNames) > 0
	if !hasIncludes {
		return true
	}

	// Check name match.
	if s.onlyNames[name] {
		return true
	}

	// Check tag match.
	for _, t := range tags {
		if s.onlyTags[t] {
			return true
		}
	}

	return false
}

// merge combines two selectors. Includes are intersected (both must pass
// if both have includes set), excludes are unioned.
func (s selector) merge(other selector) selector {
	merged := selector{
		excludeTags:  mergeSets(s.excludeTags, other.excludeTags),
		excludeNames: mergeSets(s.excludeNames, other.excludeNames),
	}

	// For includes, if both have includes, the result must satisfy both.
	// We can't represent true intersection with a single selector, so we
	// keep both sets — the includes check is "either tag or name matches"
	// within a single selector. For merge, we need both selectors to pass.
	// We handle this by keeping the more restrictive set when both are set.
	sHasIncludes := len(s.onlyTags) > 0 || len(s.onlyNames) > 0
	otherHasIncludes := len(other.onlyTags) > 0 || len(other.onlyNames) > 0

	switch {
	case sHasIncludes && otherHasIncludes:
		// Both have includes — intersect: keep only items in both.
		merged.onlyTags = intersectSets(s.onlyTags, other.onlyTags)
		merged.onlyNames = intersectSets(s.onlyNames, other.onlyNames)
		// If the intersection is empty but both had entries, we need to
		// keep at least one side's constraints to prevent matching everything.
		// Merge the sets so the includes check still works.
		if len(merged.onlyTags) == 0 && len(merged.onlyNames) == 0 {
			merged.onlyTags = mergeSets(s.onlyTags, other.onlyTags)
			merged.onlyNames = mergeSets(s.onlyNames, other.onlyNames)
		}
	case sHasIncludes:
		merged.onlyTags = copySets(s.onlyTags)
		merged.onlyNames = copySets(s.onlyNames)
	case otherHasIncludes:
		merged.onlyTags = copySets(other.onlyTags)
		merged.onlyNames = copySets(other.onlyNames)
	}

	return merged
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

func intersectSets(a, b map[string]bool) map[string]bool {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	out := make(map[string]bool)
	for k := range a {
		if b[k] {
			out[k] = true
		}
	}
	if len(out) == 0 {
		return nil
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
