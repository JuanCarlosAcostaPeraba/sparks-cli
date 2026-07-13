package cmd

import (
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/app"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/model"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/output"
	"github.com/spf13/cobra"
)

type treeOptions struct {
	all  bool
	json bool
}

func newTreeCommand() *cobra.Command {
	opts := &treeOptions{}
	cmd := &cobra.Command{
		Use:   "tree",
		Short: "Display sparks as a tree",
		Long: `Display sparks in their parent-child structure.

Tree output is useful for nested thoughts. Visible tree numbers are based on
position, such as 1, 1.1 and 1.1.1. Command arguments still use the real spark
ID shown by list and JSON output.`,
		Example: `  sparks tree
  sparks tree --all
  sparks tree --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleRun(cmd, func(application *app.App) error {
				sparks, err := application.List(cmd.Context(), model.ListOptions{IncludeDone: opts.all})
				if err != nil {
					return err
				}
				return output.Tree(stdout(cmd), sparks, opts.json)
			})
		},
	}
	cmd.Flags().BoolVarP(&opts.all, "all", "a", false, "include completed sparks")
	cmd.Flags().BoolVarP(&opts.json, "json", "j", false, "write JSON output")
	return cmd
}
