package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/model"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/output"
)

func TestSparksTextOutput(t *testing.T) {
	sparks := []model.Spark{
		{ID: 1, Title: "Prepare Codex prompt"},
		{ID: 2, Title: "Publish Homebrew tap", Important: true},
		{ID: 3, Title: "Initial README", Done: true},
	}
	var buf bytes.Buffer
	if err := output.Sparks(&buf, sparks, false); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	for _, want := range []string{"STATUS  ID  TITLE", "□       1   Prepare Codex prompt", "❗       2   Publish Homebrew tap", "☑       3   Initial README"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in output %q", want, got)
		}
	}
}

func TestSparksJSONOutput(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	sparks := []model.Spark{{ID: 1, Title: "Prepare Codex prompt", CreatedAt: now, UpdatedAt: now}}
	var buf bytes.Buffer
	if err := output.Sparks(&buf, sparks, true); err != nil {
		t.Fatal(err)
	}

	var decoded []model.Spark
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if len(decoded) != 1 || decoded[0].Title != "Prepare Codex prompt" {
		t.Fatalf("unexpected JSON output: %#v", decoded)
	}
}
