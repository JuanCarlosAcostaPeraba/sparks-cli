package updater

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	installScriptURL           = "https://raw.githubusercontent.com/JuanCarlosAcostaPeraba/sparks-cli/main/scripts/install.sh"
	installPowerShellScriptURL = "https://raw.githubusercontent.com/JuanCarlosAcostaPeraba/sparks-cli/main/scripts/install.ps1"
)

type Plan struct {
	Shell    string
	Command  string
	Deferred bool

	executable string
	args       []string
	env        []string
}

type runFunc func(context.Context, Plan, io.Writer, io.Writer) error
type startFunc func(Plan, io.Writer, io.Writer) error

type Updater struct {
	goos           string
	executablePath string
	pid            int
	getenv         func(string) string
	parentShell    func() string
	lookPath       func(string) (string, error)
	run            runFunc
	start          startFunc
}

func New() *Updater {
	return &Updater{
		goos:        runtime.GOOS,
		pid:         os.Getpid(),
		getenv:      os.Getenv,
		parentShell: detectParentShell,
		lookPath:    exec.LookPath,
		run:         runPlan,
		start:       startPlan,
	}
}

func (u *Updater) Prepare() (Plan, error) {
	target := u.executablePath
	if target == "" {
		var err error
		target, err = os.Executable()
		if err != nil {
			return Plan{}, fmt.Errorf("locate executable: %w", err)
		}
		if resolved, resolveErr := filepath.EvalSymlinks(target); resolveErr == nil {
			target = resolved
		}
	}
	if _, err := os.Stat(target); err != nil {
		return Plan{}, fmt.Errorf("inspect executable: %w", err)
	}
	installDir := filepath.Dir(target)

	switch u.goos {
	case "darwin", "linux":
		return u.prepareUnix(installDir)
	case "windows":
		return u.prepareWindows(installDir)
	default:
		return Plan{}, fmt.Errorf("updates are not available for %s", u.goos)
	}
}

func (u *Updater) Execute(ctx context.Context, plan Plan, out, errOut io.Writer) error {
	if strings.TrimSpace(plan.executable) == "" {
		return errors.New("update plan has no shell executable")
	}
	if plan.Deferred {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := u.start(plan, out, errOut); err != nil {
			return fmt.Errorf("start %s installer: %w", plan.Shell, err)
		}
		return nil
	}
	if err := u.run(ctx, plan, out, errOut); err != nil {
		return fmt.Errorf("run %s installer: %w", plan.Shell, err)
	}
	return nil
}

func (u *Updater) prepareUnix(installDir string) (Plan, error) {
	shellName, shellPath := normalizeShell(u.getenv("SPARKS_SHELL"))
	if shellName == "" {
		shellName, shellPath = normalizeShell(u.parentShell())
	}
	if shellName == "" || !isSupportedUnixShell(shellName) {
		shellName, shellPath = normalizeShell(u.getenv("SHELL"))
	}
	if !isSupportedUnixShell(shellName) {
		shellName = "sh"
		shellPath = "sh"
	}
	resolved, err := u.lookPath(shellPath)
	if err != nil && shellName != "sh" {
		shellName = "sh"
		resolved, err = u.lookPath("sh")
	}
	if err != nil {
		return Plan{}, fmt.Errorf("locate shell: %w", err)
	}

	command := "curl -fsSL " + installScriptURL + " | sh"
	environment := withEnvironment(os.Environ(), "SPARKS_INSTALL_DIR", installDir, false)
	environment = withEnvironment(environment, "SPARKS_SKIP_PATH_UPDATE", "1", false)
	return Plan{
		Shell:      shellName,
		Command:    "SPARKS_INSTALL_DIR=" + quotePOSIX(installDir) + " " + command,
		executable: resolved,
		args:       []string{"-c", command},
		env:        environment,
	}, nil
}

func (u *Updater) prepareWindows(installDir string) (Plan, error) {
	shellName := "PowerShell"
	shellPath, err := u.lookPath("pwsh")
	if err == nil {
		shellName = "PowerShell 7"
	} else {
		shellPath, err = u.lookPath("powershell.exe")
		if err != nil {
			return Plan{}, errors.New("locate PowerShell: neither pwsh nor powershell.exe is available")
		}
	}

	waitAndInstall := "$parent = Get-Process -Id " + strconv.Itoa(u.pid) +
		" -ErrorAction SilentlyContinue; if ($parent) { $parent.WaitForExit() }; " +
		"irm '" + installPowerShellScriptURL + "' | iex"
	display := "$env:SPARKS_INSTALL_DIR=" + quotePowerShell(installDir) + "; irm '" +
		installPowerShellScriptURL + "' | iex"
	environment := withEnvironment(os.Environ(), "SPARKS_INSTALL_DIR", installDir, true)
	environment = withEnvironment(environment, "SPARKS_SKIP_PATH_UPDATE", "1", true)
	return Plan{
		Shell:      shellName,
		Command:    display,
		Deferred:   true,
		executable: shellPath,
		args: []string{
			"-NoLogo",
			"-NoProfile",
			"-ExecutionPolicy", "Bypass",
			"-Command", waitAndInstall,
		},
		env: environment,
	}, nil
}

func normalizeShell(value string) (string, string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ""
	}
	normalized := strings.ReplaceAll(value, "\\", "/")
	parts := strings.Split(normalized, "/")
	rawName := strings.TrimSuffix(strings.ToLower(parts[len(parts)-1]), ".exe")
	name := strings.TrimPrefix(rawName, "-")
	if strings.HasPrefix(rawName, "-") {
		value = name
	}
	return name, value
}

func detectParentShell() string {
	parentPID := os.Getppid()
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(parentPID), "comm"))
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	if runtime.GOOS == "darwin" {
		output, err := exec.Command("ps", "-p", strconv.Itoa(parentPID), "-o", "comm=").Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	}
	return ""
}

func isSupportedUnixShell(shell string) bool {
	switch shell {
	case "bash", "zsh", "fish", "sh", "dash", "ksh":
		return true
	default:
		return false
	}
}

func withEnvironment(environment []string, key, value string, caseInsensitive bool) []string {
	prefix := key + "="
	result := make([]string, 0, len(environment)+1)
	for _, entry := range environment {
		matches := strings.HasPrefix(entry, prefix)
		if caseInsensitive {
			matches = strings.EqualFold(entry[:min(len(entry), len(prefix))], prefix)
		}
		if !matches {
			result = append(result, entry)
		}
	}
	return append(result, prefix+value)
}

func quotePOSIX(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func quotePowerShell(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func runPlan(ctx context.Context, plan Plan, out, errOut io.Writer) error {
	command := exec.CommandContext(ctx, plan.executable, plan.args...)
	command.Env = plan.env
	command.Stdin = os.Stdin
	command.Stdout = out
	command.Stderr = errOut
	return command.Run()
}

func startPlan(plan Plan, out, errOut io.Writer) error {
	command := exec.Command(plan.executable, plan.args...)
	command.Env = plan.env
	command.Stdin = os.Stdin
	command.Stdout = out
	command.Stderr = errOut
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}
