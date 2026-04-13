package action

import "testing"

// --- String ---

func TestCardinality_String_Single(t *testing.T) {
	t.Parallel()
	if s := Single.String(); s != "single" {
		t.Errorf("got %q, want %q", s, "single")
	}
}

func TestCardinality_String_Multi(t *testing.T) {
	t.Parallel()
	if s := Multi.String(); s != "multi" {
		t.Errorf("got %q, want %q", s, "multi")
	}
}

func TestCardinality_String_Unknown(t *testing.T) {
	t.Parallel()
	s := Cardinality(99).String()
	if s == "single" || s == "multi" {
		t.Errorf("unknown cardinality should not return %q", s)
	}
}

// --- IsValid ---

func TestCardinality_IsValid_Single(t *testing.T) {
	t.Parallel()
	if !Single.IsValid() {
		t.Error("Single should be valid")
	}
}

func TestCardinality_IsValid_Multi(t *testing.T) {
	t.Parallel()
	if !Multi.IsValid() {
		t.Error("Multi should be valid")
	}
}

func TestCardinality_IsValid_Unknown(t *testing.T) {
	t.Parallel()
	if Cardinality(99).IsValid() {
		t.Error("Cardinality(99) should not be valid")
	}
}

func TestCardinality_IsValid_Negative(t *testing.T) {
	t.Parallel()
	if Cardinality(-1).IsValid() {
		t.Error("Cardinality(-1) should not be valid")
	}
}

// --- Validate ---

func TestCardinality_Validate_Single(t *testing.T) {
	t.Parallel()
	if err := Single.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCardinality_Validate_Multi(t *testing.T) {
	t.Parallel()
	if err := Multi.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCardinality_Validate_Unknown(t *testing.T) {
	t.Parallel()
	if err := Cardinality(99).Validate(); err == nil {
		t.Error("expected error for invalid cardinality")
	}
}
