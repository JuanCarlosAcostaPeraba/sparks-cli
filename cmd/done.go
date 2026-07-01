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
		Args:    requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleRun(cmd, func(application *app.App) error {
				spark, err := application.Done(cmd.Context(), args[0])
				if err != nil {
					return err
				}
				output.Message(stdout(cmd), "Completed spark %d", spark.ID)
				return nil
			})
		},
	}
}
