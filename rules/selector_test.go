package rules

import "testing"

func TestSelector_NoFilters(t *testing.T) {
	t.Parallel()
	s := selector{}
	if !s.includes("anything", nil) {
		t.Error("no filters should include everything")
	}
}

func TestSelector_OnlyTags_Match(t *testing.T) {
	t.Parallel()
	s := selector{onlyTags: map[string]bool{"billing": true}}
	if !s.includes("r1", []string{"billing", "auto"}) {
		t.Error("should match on tag")
	}
}

func TestSelector_OnlyTags_NoMatch(t *testing.T) {
	t.Parallel()
	s := selector{onlyTags: map[string]bool{"billing": true}}
	if s.includes("r1", []string{"shipping"}) {
		t.Error("should not match")
	}
}

func TestSelector_OnlyNames_Match(t *testing.T) {
	t.Parallel()
	s := selector{onlyNames: map[string]bool{"r1": true}}
	if !s.includes("r1", nil) {
		t.Error("should match on name")
	}
}

func TestSelector_OnlyNames_NoMatch(t *testing.T) {
	t.Parallel()
	s := selector{onlyNames: map[string]bool{"r1": true}}
	if s.includes("r2", nil) {
		t.Error("should not match")
	}
}

func TestSelector_ExcludeTags(t *testing.T) {
	t.Parallel()
	s := selector{excludeTags: map[string]bool{"archived": true}}
	if s.includes("r1", []string{"archived"}) {
		t.Error("should be excluded")
	}
	if !s.includes("r2", []string{"active"}) {
		t.Error("should be included")
	}
}

func TestSelector_ExcludeNames(t *testing.T) {
	t.Parallel()
	s := selector{excludeNames: map[string]bool{"r1": true}}
	if s.includes("r1", nil) {
		t.Error("should be excluded")
	}
	if !s.includes("r2", nil) {
		t.Error("should be included")
	}
}

func TestSelector_ExcludeOverridesInclude(t *testing.T) {
	t.Parallel()
	s := selector{
		onlyTags:    map[string]bool{"billing": true},
		excludeNames: map[string]bool{"r1": true},
	}
	if s.includes("r1", []string{"billing"}) {
		t.Error("exclude should override include")
	}
}

func TestSelector_OnlyTagsAndNames_Union(t *testing.T) {
	t.Parallel()
	s := selector{
		onlyTags:  map[string]bool{"billing": true},
		onlyNames: map[string]bool{"r2": true},
	}
	// Matches by tag.
	if !s.includes("r1", []string{"billing"}) {
		t.Error("should match by tag")
	}
	// Matches by name.
	if !s.includes("r2", []string{"shipping"}) {
		t.Error("should match by name")
	}
	// Neither.
	if s.includes("r3", []string{"shipping"}) {
		t.Error("should not match")
	}
}
