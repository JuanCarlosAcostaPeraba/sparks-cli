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
		Args:  cobra.NoArgs,
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
	cmd.Flags().BoolVar(&opts.all, "all", false, "include completed sparks")
	cmd.Flags().BoolVar(&opts.json, "json", false, "write JSON output")
	return cmd
}
