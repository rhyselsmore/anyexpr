package rules

import "testing"

func TestSet_Add_SingleEntry(t *testing.T) {
	t.Parallel()
	s := NewSet()
	s.Add(Entry{Def: Def{Name: "tag", Cardinality: Multi}, Value: "urgent"})
	if len(s.Entries()) != 1 {
		t.Fatalf("got %d entries, want 1", len(s.Entries()))
	}
}

func TestSet_Add_MultipleEntries(t *testing.T) {
	t.Parallel()
	s := NewSet()
	s.Add(Entry{Def: Def{Name: "tag", Cardinality: Multi}, Value: "a"})
	s.Add(Entry{Def: Def{Name: "tag", Cardinality: Multi}, Value: "b"})
	if len(s.Entries()) != 2 {
		t.Fatalf("got %d entries, want 2", len(s.Entries()))
	}
	if s.Entries()[0].Value != "a" || s.Entries()[1].Value != "b" {
		t.Error("entries not in insertion order")
	}
}

func TestSet_HasTerminal_WithTerminal(t *testing.T) {
	t.Parallel()
	s := NewSet()
	s.Add(Entry{Def: Def{Name: "delete", Terminal: true}})
	if !s.HasTerminal() {
		t.Error("expected HasTerminal true")
	}
}

func TestSet_HasTerminal_WithoutTerminal(t *testing.T) {
	t.Parallel()
	s := NewSet()
	s.Add(Entry{Def: Def{Name: "tag", Cardinality: Multi}, Value: "x"})
	if s.HasTerminal() {
		t.Error("expected HasTerminal false")
	}
}

func TestSet_Resolve_MultiAction(t *testing.T) {
	t.Parallel()
	s := NewSet()
	s.Add(Entry{Def: Def{Name: "tag", Cardinality: Multi, Value: StringVal}, Value: "a"})
	s.Add(Entry{Def: Def{Name: "tag", Cardinality: Multi, Value: StringVal}, Value: "b"})
	s.Add(Entry{Def: Def{Name: "tag", Cardinality: Multi, Value: StringVal}, Value: "a"}) // dupe

	ra := s.Resolve()
	vals := ra.ByName["tag"]
	if len(vals) != 2 {
		t.Fatalf("got %d values, want 2: %v", len(vals), vals)
	}
	if vals[0] != "a" || vals[1] != "b" {
		t.Errorf("got %v, want [a b]", vals)
	}
}

func TestSet_Resolve_SingleAction(t *testing.T) {
	t.Parallel()
	s := NewSet()
	s.Add(Entry{Def: Def{Name: "category", Cardinality: Single, Value: StringVal}, Value: "first"})
	s.Add(Entry{Def: Def{Name: "category", Cardinality: Single, Value: StringVal}, Value: "second"})

	ra := s.Resolve()
	vals := ra.ByName["category"]
	if len(vals) != 1 || vals[0] != "second" {
		t.Errorf("got %v, want [second]", vals)
	}
}

func TestSet_Resolve_BoolAction(t *testing.T) {
	t.Parallel()
	s := NewSet()
	tr, fa := true, false
	s.Add(Entry{Def: Def{Name: "read", Cardinality: Single, Value: BoolValue}, BoolVal: &tr})
	s.Add(Entry{Def: Def{Name: "read", Cardinality: Single, Value: BoolValue}, BoolVal: &fa})

	ra := s.Resolve()
	if ra.Flags["read"] == nil || *ra.Flags["read"] != false {
		t.Error("expected read=false (last wins)")
	}
}

func TestSet_Resolve_Terminal(t *testing.T) {
	t.Parallel()
	s := NewSet()
	s.Add(Entry{Def: Def{Name: "delete", Terminal: true, Cardinality: Single, Value: NoValue}})

	ra := s.Resolve()
	if !ra.Terminal {
		t.Error("expected Terminal true")
	}
}

func TestSet_Resolve_Handlers(t *testing.T) {
	t.Parallel()
	s := NewSet()
	s.Add(Entry{Def: Def{Name: "h1", IsHandler: true}, RuleName: "r1"})
	s.Add(Entry{Def: Def{Name: "h2", IsHandler: true}, RuleName: "r2"})
	s.Add(Entry{Def: Def{Name: "h1", IsHandler: true}, RuleName: "r3"}) // dupe

	ra := s.Resolve()
	if len(ra.Handlers) != 2 {
		t.Fatalf("got %d handlers, want 2: %v", len(ra.Handlers), ra.Handlers)
	}
	if ra.Handlers[0] != "h1" || ra.Handlers[1] != "h2" {
		t.Errorf("got %v, want [h1 h2]", ra.Handlers)
	}
}

func TestSet_Resolve_Empty(t *testing.T) {
	t.Parallel()
	s := NewSet()
	ra := s.Resolve()

	if ra.ByName == nil {
		t.Error("ByName is nil, want initialised")
	}
	if ra.Flags == nil {
		t.Error("Flags is nil, want initialised")
	}
	if ra.Handlers == nil {
		t.Error("Handlers is nil, want initialised")
	}
	if ra.Terminal {
		t.Error("expected Terminal false")
	}
}

func TestSet_Resolve_MixedActions(t *testing.T) {
	t.Parallel()
	s := NewSet()
	tr := true
	s.Add(Entry{Def: Def{Name: "tag", Cardinality: Multi, Value: StringVal}, Value: "a"})
	s.Add(Entry{Def: Def{Name: "category", Cardinality: Single, Value: StringVal}, Value: "x"})
	s.Add(Entry{Def: Def{Name: "read", Cardinality: Single, Value: BoolValue}, BoolVal: &tr})
	s.Add(Entry{Def: Def{Name: "delete", Terminal: true, Cardinality: Single, Value: NoValue}})
	s.Add(Entry{Def: Def{Name: "h1", IsHandler: true}})

	ra := s.Resolve()
	if len(ra.ByName["tag"]) != 1 || ra.ByName["tag"][0] != "a" {
		t.Errorf("tag: got %v", ra.ByName["tag"])
	}
	if len(ra.ByName["category"]) != 1 || ra.ByName["category"][0] != "x" {
		t.Errorf("category: got %v", ra.ByName["category"])
	}
	if ra.Flags["read"] == nil || *ra.Flags["read"] != true {
		t.Error("read: expected true")
	}
	if !ra.Terminal {
		t.Error("expected Terminal true")
	}
	if len(ra.Handlers) != 1 || ra.Handlers[0] != "h1" {
		t.Errorf("handlers: got %v", ra.Handlers)
	}
}

func TestSet_Resolve_EmptyStringSkipped(t *testing.T) {
	t.Parallel()
	s := NewSet()
	s.Add(Entry{Def: Def{Name: "tag", Cardinality: Multi, Value: StringVal}, Value: ""})

	ra := s.Resolve()
	if len(ra.ByName["tag"]) != 0 {
		t.Errorf("expected empty, got %v", ra.ByName["tag"])
	}
}
