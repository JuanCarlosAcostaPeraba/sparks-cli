package cmd

import (
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/app"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/output"
	"github.com/spf13/cobra"
)

func newRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <id>",
		Aliases: []string{"rm", "-"},
		Short:   "Remove a spark",
		Long: `Remove a spark from active views.

Removal is a soft delete: the spark is marked deleted in the local database
instead of being physically removed immediately.`,
		Example: `  sparks remove 3
  sparks rm 3
  sparks - 3`,
		Args: requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleRun(cmd, func(application *app.App) error {
				if err := application.Remove(cmd.Context(), args[0]); err != nil {
					return err
				}
				output.Message(stdout(cmd), "Removed spark %s", output.ID(stdout(cmd), args[0]))
				return nil
			})
		},
	}
}
