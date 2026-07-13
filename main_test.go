package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestRunPrintsCommandErrors(t *testing.T) {
	var errOut bytes.Buffer
	exitCode := run(func() error { return errors.New("update failed") }, &errOut)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if errOut.String() != "update failed\n" {
		t.Fatalf("stderr = %q", errOut.String())
	}
}

func TestRunSucceedsSilently(t *testing.T) {
	var errOut bytes.Buffer
	exitCode := run(func() error { return nil }, &errOut)
	if exitCode != 0 || errOut.Len() != 0 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, errOut.String())
	}
}
