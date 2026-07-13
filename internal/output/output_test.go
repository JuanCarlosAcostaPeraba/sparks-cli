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
	for _, want := range []string{"STATUS  ID  TITLE", "[ ]     1   Prepare Codex prompt", "[!]     2   Publish Homebrew tap", "[x]     3   Initial README"} {
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

func TestColoredSparksStylesIDsAndStatesWithoutBreakingAlignment(t *testing.T) {
	sparks := []model.Spark{
		{ID: 1, Title: "Active"},
		{ID: 22, Title: "Important", Important: true},
		{ID: 333, Title: "Completed", Done: true},
	}
	var buf bytes.Buffer
	renderer := output.NewRenderer(&buf, true)
	if err := renderer.Sparks(sparks, false); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{
		"\x1b[2mSTATUS  ID   TITLE\x1b[0m",
		"\x1b[36m1\x1b[0m",
		"\x1b[1;33m[!]\x1b[0m",
		"\x1b[1;33mImportant\x1b[0m",
		"\x1b[32m[x]\x1b[0m",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in colored output %q", want, got)
		}
	}
	plain := stripANSI(got)
	for _, want := range []string{"STATUS  ID   TITLE", "[ ]     1    Active", "[!]     22   Important", "[x]     333  Completed"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected aligned %q in stripped output %q", want, plain)
		}
	}
}

func TestJSONNeverContainsColor(t *testing.T) {
	var buf bytes.Buffer
	renderer := output.NewRenderer(&buf, true)
	if err := renderer.Sparks([]model.Spark{{ID: 1, Title: "Plain JSON"}}, true); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "\x1b[") {
		t.Fatalf("JSON contains ANSI escapes: %q", buf.String())
	}
}

func TestColoredTreeAndMessageProvideVisualFeedback(t *testing.T) {
	var buf bytes.Buffer
	renderer := output.NewRenderer(&buf, true)
	if err := renderer.Tree([]model.Spark{{ID: 1, Title: "Important", Important: true}}, false); err != nil {
		t.Fatal(err)
	}
	renderer.Message("Added spark %s", renderer.ID(1))
	got := buf.String()
	for _, want := range []string{"\x1b[36m1\x1b[0m", "\x1b[1;33mImportant\x1b[0m", "\x1b[32m✓\x1b[0m"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in feedback %q", want, got)
		}
	}
}

func TestTreeUsesHierarchicalNumbers(t *testing.T) {
	parentID := int64(10)
	childID := int64(20)
	sparks := []model.Spark{
		{ID: parentID, Title: "Parent idea"},
		{ID: childID, Title: "Child idea", ParentID: &parentID},
		{ID: 30, Title: "Nested idea", ParentID: &childID},
	}
	var buf bytes.Buffer
	if err := output.Tree(&buf, sparks, false); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	for _, want := range []string{"└─ [ ] 1) Parent idea", "   └─ [ ] 1.1) Child idea", "      └─ [ ] 1.1.1) Nested idea"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in output %q", want, got)
		}
	}
}

func stripANSI(value string) string {
	for {
		start := strings.Index(value, "\x1b[")
		if start < 0 {
			return value
		}
		end := strings.IndexByte(value[start:], 'm')
		if end < 0 {
			return value
		}
		value = value[:start] + value[start+end+1:]
	}
}
