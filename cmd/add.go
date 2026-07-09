package cmd

import (
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
		Use:     "add \"text\"",
		Aliases: []string{"+"},
		Short:   "Add a spark",
		Args:    requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleRun(cmd, func(application *app.App) error {
				spark, err := application.Add(cmd.Context(), args[0], app.AddOptions{Parent: opts.parent})
				if err != nil {
					return err
				}
				output.Message(stdout(cmd), "Added spark %d", spark.ID)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&opts.parent, "parent", "", "add as a child of the spark id")
	return cmd
}
