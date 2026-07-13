package presentation

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestPalettePaintsRolesWhenEnabled(t *testing.T) {
	palette := Palette{Enabled: true}
	for _, role := range []Role{ID, Important, Completed, Selected, Success, Warning, Error, Muted, Key, Logo} {
		got := palette.Paint(role, "text")
		if !strings.HasPrefix(got, "\x1b[") || !strings.HasSuffix(got, "\x1b[0m") {
			t.Fatalf("role %d did not add ANSI styling: %q", role, got)
		}
	}
}

func TestPaletteLeavesTextPlainWhenDisabled(t *testing.T) {
	if got := (Palette{}).Paint(Important, "important"); got != "important" {
		t.Fatalf("disabled palette returned %q", got)
	}
}

func TestForWriterDisablesColorForNonTerminal(t *testing.T) {
	if ForWriter(&bytes.Buffer{}).Enabled {
		t.Fatal("buffer should not enable color")
	}
}

func TestAllowedRespectsNoColorAndDumbTerminal(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	if Allowed() {
		t.Fatal("NO_COLOR should disable color")
	}

	if err := os.Unsetenv("NO_COLOR"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TERM", "dumb")
	if Allowed() {
		t.Fatal("TERM=dumb should disable color")
	}
}
