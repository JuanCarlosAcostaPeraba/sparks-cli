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
		Long: `Search sparks by title.

Search is case-insensitive and returns active sparks by default. Use --all to
include completed sparks and --json for scripts or integrations.`,
		Example: `  sparks search "release"
  sparks search --all "docs"
  sparks search "homebrew" --json`,
		Args: requireArgs(1),
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
	cmd.Flags().BoolVarP(&opts.all, "all", "a", false, "include completed sparks")
	cmd.Flags().BoolVarP(&opts.json, "json", "j", false, "write JSON output")
	return cmd
}
