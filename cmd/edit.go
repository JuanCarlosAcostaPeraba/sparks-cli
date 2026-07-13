package cmd

import (
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/app"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/output"
	"github.com/spf13/cobra"
)

func newEditCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "edit <id> \"text\"",
		Aliases: []string{"e"},
		Short:   "Edit a spark title",
		Long: `Edit the title of an existing spark.

The ID is the real spark ID shown by list or JSON output. Editing preserves the
spark's status, importance and parent-child relationships.`,
		Example: `  sparks edit 3 "Ship v1.0.0"
  sparks e 3 "Ship v1.0.0"`,
		Args: requireArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleRun(cmd, func(application *app.App) error {
				spark, err := application.Edit(cmd.Context(), args[0], args[1])
				if err != nil {
					return err
				}
				output.Message(stdout(cmd), "Updated spark %s", output.ID(stdout(cmd), spark.ID))
				return nil
			})
		},
	}
}
