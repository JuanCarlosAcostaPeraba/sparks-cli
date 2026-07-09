package cmd

import (
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/app"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/output"
	"github.com/spf13/cobra"
)

func newImportantCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "important <id>",
		Aliases: []string{"!"},
		Short:   "Toggle important status",
		Long: `Toggle whether a spark is marked as important.

Important sparks are listed before normal sparks. Running the command again
removes the important mark.`,
		Example: `  sparks important 3
  sparks ! 3`,
		Args: requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleRun(cmd, func(application *app.App) error {
				spark, err := application.Important(cmd.Context(), args[0])
				if err != nil {
					return err
				}
				if spark.Important {
					output.Message(stdout(cmd), "Marked spark %d as important", spark.ID)
				} else {
					output.Message(stdout(cmd), "Unmarked spark %d as important", spark.ID)
				}
				return nil
			})
		},
	}
}
