package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/updater"
	"github.com/spf13/cobra"
)

type fakeUpdateClient struct {
	plan       updater.Plan
	prepareErr error
	executeErr error
	executed   bool
}

func (f *fakeUpdateClient) Prepare() (updater.Plan, error) {
	return f.plan, f.prepareErr
}

func (f *fakeUpdateClient) Execute(context.Context, updater.Plan, io.Writer, io.Writer) error {
	f.executed = true
	return f.executeErr
}

func TestUpdateCommandShowsDetectedShellAndCommand(t *testing.T) {
	client := &fakeUpdateClient{plan: updater.Plan{
		Shell:   "zsh",
		Command: "curl -fsSL https://example.test/install.sh | sh",
	}}
	out, _, err := runUpdateCommand(t, client)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Detected zsh", "Running: curl -fsSL", "Update finished"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output %q does not contain %q", out, want)
		}
	}
	if !client.executed {
		t.Fatal("installer was not executed")
	}
}

func TestUpdateCommandReportsDeferredWindowsInstall(t *testing.T) {
	client := &fakeUpdateClient{plan: updater.Plan{
		Shell:    "PowerShell 7",
		Command:  "irm https://example.test/install.ps1 | iex",
		Deferred: true,
	}}
	out, _, err := runUpdateCommand(t, client)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Installer started") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestUpdateCommandReturnsPreparationAndExecutionErrors(t *testing.T) {
	t.Run("prepare", func(t *testing.T) {
		_, _, err := runUpdateCommand(t, &fakeUpdateClient{prepareErr: errors.New("no shell")})
		if err == nil || !strings.Contains(err.Error(), "prepare sparks update: no shell") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("execute", func(t *testing.T) {
		_, _, err := runUpdateCommand(t, &fakeUpdateClient{
			plan:       updater.Plan{Shell: "bash", Command: "curl | sh"},
			executeErr: errors.New("installer failed"),
		})
		if err == nil || !strings.Contains(err.Error(), "update sparks: installer failed") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func runUpdateCommand(t *testing.T, client updateClient) (string, string, error) {
	t.Helper()
	var out bytes.Buffer
	var errOut bytes.Buffer
	root := &cobra.Command{Use: "sparks", SilenceUsage: true, SilenceErrors: true}
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetContext(withRootOptions(context.Background(), &rootOptions{out: &out, errOut: &errOut}))
	root.AddCommand(newUpdateCommandWithClient(client))
	root.SetArgs([]string{"update"})
	err := root.Execute()
	return out.String(), errOut.String(), err
}
