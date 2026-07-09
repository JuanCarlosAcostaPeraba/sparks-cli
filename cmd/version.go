package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	commit  = "unknown"
	date    = "unknown"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long: `Display build version information.

The output includes the release version, git commit and build date when those
values are injected by the release build.`,
		Example: `  sparks version`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(stdout(cmd), "sparks %s\ncommit: %s\nbuilt: %s\n", version, commit, date)
			return err
		},
	}
}
