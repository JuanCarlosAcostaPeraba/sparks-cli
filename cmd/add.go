package cmd

import (
	"strings"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/app"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/output"
	"github.com/spf13/cobra"
)

type addOptions struct {
	parent string
}

func newAddCommand() *cobra.Command {
	opts := &addOptions{}
	cmd := &cobra.Command{
		Use:     "add <text>",
		Aliases: []string{"+"},
		Short:   "Add a spark",
		Long: `Add a new spark from a short piece of text.

Use this for ideas, tasks or notes you want to capture quickly. Quotes are
optional, so both sparks add Ship release and sparks add "Ship release" create
the same spark.

To create a sub-idea, pass --parent with the parent spark ID shown by list or
JSON output. Child sparks appear nested under their parent in tree output.`,
		Example: `  sparks add Prepare release notes
  sparks + Fix install docs
	  sparks add --parent 1 "Add Homebrew example"
	  sparks + --parent 1 "Check macOS path"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleRun(cmd, func(application *app.App) error {
				spark, err := application.Add(cmd.Context(), strings.Join(args, " "), app.AddOptions{Parent: opts.parent})
				if err != nil {
					return err
				}
				output.Message(stdout(cmd), "Added spark %d", spark.ID)
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&opts.parent, "parent", "p", "", "add as a child of the given spark ID")
	return cmd
}
