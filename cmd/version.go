package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "1.1.0"

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Display the sparks version",
		Long:    `Display the sparks release version.`,
		Example: `  sparks version`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(stdout(cmd), "sparks %s\n", version)
			return err
		},
	}
}
