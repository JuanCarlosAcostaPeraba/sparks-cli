package updater

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestPrepareDetectsUnixShells(t *testing.T) {
	tests := []struct {
		name     string
		shell    string
		want     string
		wantPath string
	}{
		{name: "bash", shell: "/bin/bash", want: "bash", wantPath: "/bin/bash"},
		{name: "zsh", shell: "/bin/zsh", want: "zsh", wantPath: "/bin/zsh"},
		{name: "login zsh", shell: "-zsh", want: "zsh", wantPath: "zsh"},
		{name: "fish", shell: "/usr/local/bin/fish", want: "fish", wantPath: "/usr/local/bin/fish"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updater, target := testUpdater(t, "darwin")
			updater.getenv = func(key string) string {
				if key == "SHELL" {
					return test.shell
				}
				return ""
			}
			updater.lookPath = func(file string) (string, error) { return file, nil }

			plan, err := updater.Prepare()
			if err != nil {
				t.Fatal(err)
			}
			if plan.Shell != test.want || plan.executable != test.wantPath || plan.Deferred {
				t.Fatalf("unexpected plan: %#v", plan)
			}
			if !strings.Contains(plan.Command, "curl -fsSL") || !strings.Contains(plan.Command, filepath.Dir(target)) {
				t.Fatalf("command does not describe installer and target: %q", plan.Command)
			}
			if got := environmentValue(plan.env, "SPARKS_INSTALL_DIR"); got != filepath.Dir(target) {
				t.Fatalf("SPARKS_INSTALL_DIR = %q", got)
			}
			if got := environmentValue(plan.env, "SPARKS_SKIP_PATH_UPDATE"); got != "1" {
				t.Fatalf("SPARKS_SKIP_PATH_UPDATE = %q", got)
			}
		})
	}
}

func TestPrepareUsesShellOverrideAndPOSIXFallback(t *testing.T) {
	updater, _ := testUpdater(t, "linux")
	updater.getenv = func(key string) string {
		switch key {
		case "SPARKS_SHELL":
			return "zsh"
		case "SHELL":
			return "/bin/bash"
		default:
			return ""
		}
	}
	updater.lookPath = func(file string) (string, error) {
		if file == "zsh" {
			return "", errors.New("missing")
		}
		return "/bin/sh", nil
	}

	plan, err := updater.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	if plan.Shell != "sh" || plan.executable != "/bin/sh" {
		t.Fatalf("unexpected fallback plan: %#v", plan)
	}
}

func TestPreparePrefersActiveParentShellOverLoginShell(t *testing.T) {
	updater, _ := testUpdater(t, "darwin")
	updater.parentShell = func() string { return "/bin/bash" }
	updater.getenv = func(key string) string {
		if key == "SHELL" {
			return "/bin/zsh"
		}
		return ""
	}
	updater.lookPath = func(file string) (string, error) { return file, nil }

	plan, err := updater.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	if plan.Shell != "bash" || plan.executable != "/bin/bash" {
		t.Fatalf("active parent shell was not selected: %#v", plan)
	}
}

func TestPrepareDefersWindowsInstallerUntilCurrentProcessExits(t *testing.T) {
	updater, target := testUpdater(t, "windows")
	updater.pid = 4242
	updater.lookPath = func(file string) (string, error) {
		if file == "pwsh" {
			return "C:\\Program Files\\PowerShell\\7\\pwsh.exe", nil
		}
		return "", errors.New("not found")
	}

	plan, err := updater.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	if plan.Shell != "PowerShell 7" || !plan.Deferred {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	if !containsSequence(plan.args, "-Command") || !strings.Contains(strings.Join(plan.args, " "), "4242") {
		t.Fatalf("PowerShell arguments do not wait for the parent: %v", plan.args)
	}
	if got := environmentValue(plan.env, "SPARKS_INSTALL_DIR"); got != filepath.Dir(target) {
		t.Fatalf("SPARKS_INSTALL_DIR = %q", got)
	}
	if got := environmentValue(plan.env, "SPARKS_SKIP_PATH_UPDATE"); got != "1" {
		t.Fatalf("SPARKS_SKIP_PATH_UPDATE = %q", got)
	}
}

func TestExecuteRunsUnixPlanSynchronously(t *testing.T) {
	updater, _ := testUpdater(t, "linux")
	plan := Plan{Shell: "bash", executable: "/bin/bash"}
	called := false
	updater.run = func(_ context.Context, got Plan, out, errOut io.Writer) error {
		called = true
		if !reflect.DeepEqual(got, plan) || out == nil || errOut == nil {
			t.Fatal("run received unexpected arguments")
		}
		return nil
	}
	updater.start = func(Plan, io.Writer, io.Writer) error {
		t.Fatal("deferred runner should not be called")
		return nil
	}

	if err := updater.Execute(context.Background(), plan, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("synchronous runner was not called")
	}
}

func TestExecuteStartsWindowsPlanAndHonorsCancellation(t *testing.T) {
	updater, _ := testUpdater(t, "windows")
	plan := Plan{Shell: "PowerShell", Deferred: true, executable: "powershell.exe"}
	starts := 0
	updater.start = func(Plan, io.Writer, io.Writer) error {
		starts++
		return nil
	}

	if err := updater.Execute(context.Background(), plan, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := updater.Execute(ctx, plan, &bytes.Buffer{}, &bytes.Buffer{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled update returned %v", err)
	}
	if starts != 1 {
		t.Fatalf("installer starts = %d, want 1", starts)
	}
}

func TestPrepareRejectsUnsupportedOSAndMissingExecutable(t *testing.T) {
	updater, _ := testUpdater(t, "plan9")
	if _, err := updater.Prepare(); err == nil || !strings.Contains(err.Error(), "plan9") {
		t.Fatalf("unexpected unsupported OS error: %v", err)
	}
	updater.goos = "linux"
	updater.executablePath = filepath.Join(t.TempDir(), "missing")
	if _, err := updater.Prepare(); err == nil || !strings.Contains(err.Error(), "inspect executable") {
		t.Fatalf("unexpected missing executable error: %v", err)
	}
}

func testUpdater(t *testing.T, goos string) (*Updater, string) {
	t.Helper()
	target := filepath.Join(t.TempDir(), "sparks")
	if goos == "windows" {
		target += ".exe"
	}
	if err := os.WriteFile(target, []byte("test"), 0o755); err != nil {
		t.Fatal(err)
	}
	updater := New()
	updater.goos = goos
	updater.executablePath = target
	updater.parentShell = func() string { return "" }
	return updater, target
}

func environmentValue(environment []string, key string) string {
	prefix := key + "="
	for _, entry := range environment {
		if strings.HasPrefix(entry, prefix) {
			return strings.TrimPrefix(entry, prefix)
		}
	}
	return ""
}

func containsSequence(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
