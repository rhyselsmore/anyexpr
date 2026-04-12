package rules

import "testing"

func TestSelector_IncludesByTag(t *testing.T) {
	t.Parallel()
	s := selector{onlyTags: map[string]bool{"urgent": true}}
	if !s.includes("rule1", []string{"urgent", "billing"}) {
		t.Error("expected included by tag")
	}
	if s.includes("rule1", []string{"billing"}) {
		t.Error("expected excluded, no matching tag")
	}
}

func TestSelector_IncludesByName(t *testing.T) {
	t.Parallel()
	s := selector{onlyNames: map[string]bool{"rule1": true}}
	if !s.includes("rule1", nil) {
		t.Error("expected included by name")
	}
	if s.includes("rule2", nil) {
		t.Error("expected excluded, wrong name")
	}
}

func TestSelector_ExcludesByTag(t *testing.T) {
	t.Parallel()
	s := selector{excludeTags: map[string]bool{"internal": true}}
	if s.includes("rule1", []string{"internal"}) {
		t.Error("expected excluded by tag")
	}
	if !s.includes("rule1", []string{"external"}) {
		t.Error("expected included")
	}
}

func TestSelector_ExcludesByName(t *testing.T) {
	t.Parallel()
	s := selector{excludeNames: map[string]bool{"rule1": true}}
	if s.includes("rule1", nil) {
		t.Error("expected excluded by name")
	}
	if !s.includes("rule2", nil) {
		t.Error("expected included")
	}
}

func TestSelector_ExcludeOverridesInclude(t *testing.T) {
	t.Parallel()
	s := selector{
		onlyNames:    map[string]bool{"rule1": true},
		excludeNames: map[string]bool{"rule1": true},
	}
	if s.includes("rule1", nil) {
		t.Error("exclude should override include")
	}
}

func TestSelector_EmptyIncludesAll(t *testing.T) {
	t.Parallel()
	s := selector{}
	if !s.includes("anything", []string{"any", "tag"}) {
		t.Error("empty selector should include everything")
	}
}

func TestSelector_BothTagAndNameIncludes(t *testing.T) {
	t.Parallel()
	s := selector{
		onlyTags:  map[string]bool{"urgent": true},
		onlyNames: map[string]bool{"rule1": true},
	}
	// Either match suffices.
	if !s.includes("rule1", nil) {
		t.Error("should match by name")
	}
	if !s.includes("rule2", []string{"urgent"}) {
		t.Error("should match by tag")
	}
	if s.includes("rule2", []string{"billing"}) {
		t.Error("should not match, neither name nor tag")
	}
}

func TestSelector_MergeUnionsExcludes(t *testing.T) {
	t.Parallel()
	a := selector{excludeNames: map[string]bool{"r1": true}}
	b := selector{excludeNames: map[string]bool{"r2": true}}
	m := a.merge(b)
	if m.includes("r1", nil) || m.includes("r2", nil) {
		t.Error("merged excludes should union")
	}
	if !m.includes("r3", nil) {
		t.Error("r3 should be included")
	}
}

func TestSelector_MergeIntersectsIncludes(t *testing.T) {
	t.Parallel()
	a := selector{onlyTags: map[string]bool{"urgent": true, "billing": true}}
	b := selector{onlyTags: map[string]bool{"urgent": true, "shipping": true}}
	m := a.merge(b)
	if !m.includes("r1", []string{"urgent"}) {
		t.Error("urgent is in both, should be included")
	}
}
