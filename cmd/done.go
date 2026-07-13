package cmd

import (
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/app"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/output"
	"github.com/spf13/cobra"
)

func newDoneCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "done <id>",
		Aliases: []string{"ok"},
		Short:   "Mark a spark as completed",
		Long: `Mark a spark as completed.

The ID is the real spark ID shown by list or JSON output. Completed sparks are
hidden from the default list and tree views, but can be shown with --all on
commands that support it.`,
		Example: `  sparks done 3
  sparks ok 3`,
		Args: requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleRun(cmd, func(application *app.App) error {
				spark, err := application.Done(cmd.Context(), args[0])
				if err != nil {
					return err
				}
				output.Message(stdout(cmd), "Completed spark %s", output.ID(stdout(cmd), spark.ID))
				return nil
			})
		},
	}
}
