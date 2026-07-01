package app_test

import (
	"testing"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/app"
)

func TestParseID(t *testing.T) {
	id, err := app.ParseID("42")
	if err != nil {
		t.Fatalf("ParseID returned error: %v", err)
	}
	if id != 42 {
		t.Fatalf("expected 42, got %d", id)
	}
}

func TestParseIDRejectsInvalidValues(t *testing.T) {
	for _, raw := range []string{"", "abc", "0", "-1"} {
		if _, err := app.ParseID(raw); err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
}
