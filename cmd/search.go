package cmd

import (
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/app"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/model"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/output"
	"github.com/spf13/cobra"
)

type searchOptions struct {
	all  bool
	json bool
}

func newSearchCommand() *cobra.Command {
	opts := &searchOptions{}
	cmd := &cobra.Command{
		Use:   "search \"query\"",
		Short: "Search sparks by title",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleRun(cmd, func(application *app.App) error {
				sparks, err := application.Search(cmd.Context(), args[0], model.ListOptions{IncludeDone: opts.all})
				if err != nil {
					return err
				}
				return output.Sparks(stdout(cmd), sparks, opts.json)
			})
		},
	}
	cmd.Flags().BoolVar(&opts.all, "all", false, "include completed sparks")
	cmd.Flags().BoolVar(&opts.json, "json", false, "write JSON output")
	return cmd
}
