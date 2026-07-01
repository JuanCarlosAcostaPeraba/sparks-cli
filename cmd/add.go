package cmd

import (
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/app"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/output"
	"github.com/spf13/cobra"
)

func newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add \"text\"",
		Aliases: []string{"+"},
		Short:   "Add a spark",
		Args:    requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleRun(cmd, func(application *app.App) error {
				spark, err := application.Add(cmd.Context(), args[0])
				if err != nil {
					return err
				}
				output.Message(stdout(cmd), "Added spark %d", spark.ID)
				return nil
			})
		},
	}
	return cmd
}
