package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootCommandAddAndList(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "sparks.db")

	out, errOut, err := runCommand(t, dbPath, "add", "Prepare Codex prompt")
	if err != nil {
		t.Fatalf("add failed: %v\nstderr: %s", err, errOut)
	}
	if !strings.Contains(out, "Added spark 1") {
		t.Fatalf("unexpected add output: %q", out)
	}

	out, errOut, err = runCommand(t, dbPath, "list")
	if err != nil {
		t.Fatalf("list failed: %v\nstderr: %s", err, errOut)
	}
	if !strings.Contains(out, "STATUS  ID  TITLE") || !strings.Contains(out, "□       1   Prepare Codex prompt") {
		t.Fatalf("unexpected list output: %q", out)
	}
}

func TestRootCommandJSONList(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "sparks.db")
	if _, _, err := runCommand(t, dbPath, "+", "Prepare Codex prompt"); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := runCommand(t, dbPath, "list", "--json")
	if err != nil {
		t.Fatalf("list failed: %v\nstderr: %s", err, errOut)
	}
	if !strings.Contains(out, `"title": "Prepare Codex prompt"`) {
		t.Fatalf("unexpected JSON output: %q", out)
	}
}

func TestRootCommandShortAllFlag(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "sparks.db")
	if _, _, err := runCommand(t, dbPath, "add", "Completed spark"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runCommand(t, dbPath, "done", "1"); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := runCommand(t, dbPath, "list", "-a")
	if err != nil {
		t.Fatalf("list -a failed: %v\nstderr: %s", err, errOut)
	}
	if !strings.Contains(out, "Completed spark") {
		t.Fatalf("expected -a to include completed sparks, got: %q", out)
	}
}

func TestCommandFlagShorthands(t *testing.T) {
	tests := []struct {
		command   string
		flag      string
		shorthand string
	}{
		{"", "db", "d"},
		{"add", "parent", "p"},
		{"list", "all", "a"},
		{"list", "json", "j"},
		{"clear", "all", "a"},
		{"clear", "yes", "y"},
		{"tree", "all", "a"},
		{"tree", "json", "j"},
		{"search", "all", "a"},
		{"search", "json", "j"},
	}

	root := NewRootCommand(&bytes.Buffer{}, &bytes.Buffer{})
	for _, tt := range tests {
		t.Run(tt.command+"/"+tt.flag, func(t *testing.T) {
			cmd := root
			flags := root.PersistentFlags()
			if tt.command != "" {
				var err error
				cmd, _, err = root.Find([]string{tt.command})
				if err != nil {
					t.Fatal(err)
				}
				flags = cmd.Flags()
			}

			flag := flags.Lookup(tt.flag)
			if flag == nil {
				t.Fatalf("flag --%s not found", tt.flag)
			}
			if flag.Shorthand != tt.shorthand {
				t.Fatalf("--%s shorthand = %q, want %q", tt.flag, flag.Shorthand, tt.shorthand)
			}
		})
	}
}

func TestRootCommandAddsChildSpark(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "sparks.db")
	if _, _, err := runCommand(t, dbPath, "add", "Parent idea"); err != nil {
		t.Fatal(err)
	}
	if _, errOut, err := runCommand(t, dbPath, "add", "--parent", "1", "Child idea"); err != nil {
		t.Fatalf("add child failed: %v\nstderr: %s", err, errOut)
	}

	out, errOut, err := runCommand(t, dbPath, "tree")
	if err != nil {
		t.Fatalf("tree failed: %v\nstderr: %s", err, errOut)
	}
	if !strings.Contains(out, "└─ □ 1) Parent idea") || !strings.Contains(out, "   └─ □ 1.1) Child idea") {
		t.Fatalf("unexpected tree output: %q", out)
	}
}

func TestRootCommandHelpIsExplanatory(t *testing.T) {
	out, errOut, err := runCommand(t, "", "--help")
	if err != nil {
		t.Fatalf("help failed: %v\nstderr: %s", err, errOut)
	}
	for _, want := range []string{
		"sparks captures ideas, tasks and nested thoughts",
		"Run sparks with no command to list active items.",
		"Create sub-ideas with add --parent <id>",
		"sparks add --parent 1 \"Document install steps\"",
		"Available Commands:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in help output: %q", want, out)
		}
	}
}

func TestAddAliasHelpIsCommandSpecific(t *testing.T) {
	out, errOut, err := runCommand(t, "", "+", "-h")
	if err != nil {
		t.Fatalf("add help failed: %v\nstderr: %s", err, errOut)
	}
	for _, want := range []string{
		"Add a new spark from a short piece of text.",
		"To create a sub-idea, pass --parent",
		"sparks + \"Fix install docs\"",
		"--parent string   add as a child of the given spark ID",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in add help output: %q", want, out)
		}
	}
	if strings.Contains(out, "Available Commands:") {
		t.Fatalf("expected add help to stay command-specific, got: %q", out)
	}
}

func runCommand(t *testing.T, dbPath string, args ...string) (string, string, error) {
	t.Helper()
	var out bytes.Buffer
	var errOut bytes.Buffer
	root := NewRootCommand(&out, &errOut)
	root.SetArgs(append([]string{"--db", dbPath}, args...))
	err := root.Execute()
	return out.String(), errOut.String(), err
}
