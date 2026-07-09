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

func runCommand(t *testing.T, dbPath string, args ...string) (string, string, error) {
	t.Helper()
	var out bytes.Buffer
	var errOut bytes.Buffer
	root := NewRootCommand(&out, &errOut)
	root.SetArgs(append([]string{"--db", dbPath}, args...))
	err := root.Execute()
	return out.String(), errOut.String(), err
}
