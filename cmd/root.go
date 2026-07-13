package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/app"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/storage"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	dbPath string
	out    io.Writer
	errOut io.Writer
}

func Execute() error {
	root := NewRootCommand(os.Stdout, os.Stderr)
	return root.Execute()
}

func NewRootCommand(out, errOut io.Writer) *cobra.Command {
	opts := &rootOptions{out: out, errOut: errOut}

	root := &cobra.Command{
		Use:   "sparks",
		Short: "A tiny, fast CLI to capture ideas, tasks and nested thoughts.",
		Long: `sparks captures ideas, tasks and nested thoughts without leaving the terminal.

Run sparks with no command to list active items. Use add, done, important,
remove, search and tree to keep lightweight notes organized in a local SQLite
database stored in your application data directory.

Create sub-ideas with add --parent <id>, where <id> is the parent spark ID
shown by list or JSON output.`,
		Example: `  sparks
  sparks add "Prepare release notes"
  sparks add --parent 1 "Document install steps"
  sparks tree
  sparks done 2`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, args)
		},
	}
	root.SetOut(out)
	root.SetErr(errOut)
	root.PersistentFlags().StringVarP(&opts.dbPath, "db", "d", "", "use a specific SQLite database path")
	root.SetContext(withRootOptions(context.Background(), opts))

	root.AddCommand(newListCommand())
	root.AddCommand(newAddCommand())
	root.AddCommand(newDoneCommand())
	root.AddCommand(newImportantCommand())
	root.AddCommand(newRemoveCommand())
	root.AddCommand(newClearCommand())
	root.AddCommand(newTreeCommand())
	root.AddCommand(newSearchCommand())
	root.AddCommand(newVersionCommand())
	return root
}

func newApp(cmd *cobra.Command) (*app.App, func() error, error) {
	opts := getRootOptions(cmd)
	dbPath := opts.dbPath
	if dbPath == "" {
		path, err := storage.DefaultDBPath()
		if err != nil {
			return nil, nil, err
		}
		dbPath = path
	}

	store, err := storage.Open(dbPath)
	if err != nil {
		return nil, nil, err
	}
	return app.New(store), store.Close, nil
}

func handleRun(cmd *cobra.Command, fn func(*app.App) error) error {
	application, closeFn, err := newApp(cmd)
	if err != nil {
		return err
	}
	defer closeFn()

	if err := fn(application); err != nil {
		return app.FriendlyError(err)
	}
	return nil
}

func stdout(cmd *cobra.Command) io.Writer {
	return getRootOptions(cmd).out
}

func requireArgs(count int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != count {
			return fmt.Errorf("expected %d argument(s), got %d", count, len(args))
		}
		return nil
	}
}
