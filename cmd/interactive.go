package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/app"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/presentation"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func runInteractive(cmd *cobra.Command, _ []string) error {
	if isInteractiveTerminal(cmd) {
		return runTUI(cmd)
	}
	return runLineInteractive(cmd)
}

func isInteractiveTerminal(cmd *cobra.Command) bool {
	in, inputIsFile := cmd.InOrStdin().(*os.File)
	out, outputIsFile := stdout(cmd).(*os.File)
	return inputIsFile && outputIsFile &&
		term.IsTerminal(int(in.Fd())) && term.IsTerminal(int(out.Fd()))
}

func runTUI(cmd *cobra.Command) error {
	application, closeFn, err := newApp(cmd)
	if err != nil {
		return err
	}
	defer closeFn()

	program := tea.NewProgram(
		tui.New(cmd.Context(), application, tui.WithColor(presentation.Allowed())),
		tea.WithInput(cmd.InOrStdin()),
		tea.WithOutput(stdout(cmd)),
		tea.WithAltScreen(),
	)
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run interactive interface: %w", err)
	}
	return nil
}

func runLineInteractive(cmd *cobra.Command) error {
	if err := executeInteractiveCommand(cmd, []string{"list"}); err != nil {
		return err
	}
	fmt.Fprintln(stdout(cmd), "Interactive mode. Type help for commands, or exit to quit.")

	scanner := bufio.NewScanner(cmd.InOrStdin())
	for {
		fmt.Fprint(stdout(cmd), "sparks> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("read interactive command: %w", err)
			}
			return nil
		}

		args, err := parseInteractiveLine(scanner.Text())
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			continue
		}
		if len(args) == 0 {
			continue
		}
		switch args[0] {
		case "exit", "quit":
			return nil
		case "?":
			args = []string{"--help"}
		}

		if err := executeInteractiveCommand(cmd, args); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), app.FriendlyError(err))
		}
	}
}

func executeInteractiveCommand(parent *cobra.Command, args []string) error {
	opts := getRootOptions(parent)
	root := NewRootCommand(opts.out, opts.errOut)
	root.SetIn(parent.InOrStdin())
	if opts.dbPath != "" {
		args = append([]string{"--db", opts.dbPath}, args...)
	}
	root.SetArgs(args)
	return root.ExecuteContext(parent.Context())
}

func parseInteractiveLine(line string) ([]string, error) {
	var args []string
	var current strings.Builder
	var quote rune
	tokenStarted := false

	flush := func() {
		if tokenStarted {
			args = append(args, current.String())
			current.Reset()
			tokenStarted = false
		}
	}

	for _, char := range line {
		if quote != 0 {
			if char == quote {
				quote = 0
			} else {
				current.WriteRune(char)
			}
			continue
		}

		switch {
		case (char == '\'' || char == '"') && !tokenStarted:
			quote = char
			tokenStarted = true
		case unicode.IsSpace(char):
			flush()
		default:
			current.WriteRune(char)
			tokenStarted = true
		}
	}
	if quote != 0 {
		return nil, errors.New("unterminated quote")
	}
	flush()
	return args, nil
}
