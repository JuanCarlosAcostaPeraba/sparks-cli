package cmd

import (
	"errors"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/app"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/output"
	"github.com/spf13/cobra"
)

type clearOptions struct {
	all bool
	yes bool
}

func newClearCommand() *cobra.Command {
	opts := &clearOptions{}
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear completed sparks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.all && !opts.yes {
				return errors.New("clearing all sparks requires --yes")
			}
			return handleRun(cmd, func(application *app.App) error {
				count, err := application.Clear(cmd.Context(), opts.all)
				if err != nil {
					return err
				}
				output.Message(stdout(cmd), "Cleared %d spark(s)", count)
				return nil
			})
		},
	}
	cmd.Flags().BoolVar(&opts.all, "all", false, "clear all sparks")
	cmd.Flags().BoolVar(&opts.yes, "yes", false, "confirm clearing all sparks")
	return cmd
}
