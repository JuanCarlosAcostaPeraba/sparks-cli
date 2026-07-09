package cmd

import (
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/app"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/model"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/output"
	"github.com/spf13/cobra"
)

type listOptions struct {
	all  bool
	json bool
}

func newListCommand() *cobra.Command {
	opts := &listOptions{}
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List active sparks",
		Long: `List sparks in a compact table.

By default, list shows active sparks only. Use --all to include completed
sparks, or --json when another tool needs structured output.`,
		Example: `  sparks list
  sparks ls --all
  sparks list --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListWithOptions(cmd, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.all, "all", false, "include completed sparks")
	cmd.Flags().BoolVar(&opts.json, "json", false, "write JSON output")
	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cobra.NoArgs(cmd, args)
	}
	return runListWithOptions(cmd, &listOptions{})
}

func runListWithOptions(cmd *cobra.Command, opts *listOptions) error {
	return handleRun(cmd, func(application *app.App) error {
		sparks, err := application.List(cmd.Context(), model.ListOptions{IncludeDone: opts.all})
		if err != nil {
			return err
		}
		return output.Sparks(stdout(cmd), sparks, opts.json)
	})
}
