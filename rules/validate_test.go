package rules

import (
	"errors"
	"testing"
)

func TestValidateActions_Valid(t *testing.T) {
	t.Parallel()
	err := ValidateActions([]Def{
		{Name: "tag", Cardinality: Multi},
		{Name: "tag", Cardinality: Multi},
		{Name: "category", Cardinality: Single},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateActions_MultipleTerminals(t *testing.T) {
	t.Parallel()
	err := ValidateActions([]Def{
		{Name: "delete", Terminal: true, Cardinality: Single},
		{Name: "archive", Terminal: true, Cardinality: Single},
	})
	if !errors.Is(err, ErrMultipleTerminals) {
		t.Errorf("got %v, want ErrMultipleTerminals", err)
	}
}

func TestValidateActions_DuplicateSingle(t *testing.T) {
	t.Parallel()
	err := ValidateActions([]Def{
		{Name: "category", Cardinality: Single},
		{Name: "category", Cardinality: Single},
	})
	if !errors.Is(err, ErrCardinalityViolation) {
		t.Errorf("got %v, want ErrCardinalityViolation", err)
	}
}

func TestValidateActions_DuplicateMulti(t *testing.T) {
	t.Parallel()
	err := ValidateActions([]Def{
		{Name: "tag", Cardinality: Multi},
		{Name: "tag", Cardinality: Multi},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateActions_SingleTerminal(t *testing.T) {
	t.Parallel()
	err := ValidateActions([]Def{
		{Name: "tag", Cardinality: Multi},
		{Name: "delete", Terminal: true, Cardinality: Single},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateActions_Empty(t *testing.T) {
	t.Parallel()
	err := ValidateActions([]Def{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateActions_ErrorType(t *testing.T) {
	t.Parallel()
	err := ValidateActions([]Def{
		{Name: "a", Terminal: true},
		{Name: "b", Terminal: true},
	})
	if !errors.Is(err, ErrMultipleTerminals) {
		t.Errorf("errors.Is failed for ErrMultipleTerminals")
	}
}
