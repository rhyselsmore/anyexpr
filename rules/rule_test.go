package rules

import "testing"

func TestDefinition_IsEnabled_Default(t *testing.T) {
	t.Parallel()
	d := Definition{Name: "r"}
	if !d.IsEnabled() {
		t.Error("nil Enabled should default to true")
	}
}

func TestDefinition_IsEnabled_True(t *testing.T) {
	t.Parallel()
	b := true
	d := Definition{Enabled: &b}
	if !d.IsEnabled() {
		t.Error("expected true")
	}
}

func TestDefinition_IsEnabled_False(t *testing.T) {
	t.Parallel()
	b := false
	d := Definition{Enabled: &b}
	if d.IsEnabled() {
		t.Error("expected false")
	}
}
