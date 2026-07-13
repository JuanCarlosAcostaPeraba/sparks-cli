package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/output"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/updater"
	"github.com/spf13/cobra"
)

type updateClient interface {
	Prepare() (updater.Plan, error)
	Execute(context.Context, updater.Plan, io.Writer, io.Writer) error
}

func newUpdateCommand() *cobra.Command {
	return newUpdateCommandWithClient(updater.New())
}

func newUpdateCommandWithClient(client updateClient) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update sparks to the latest release",
		Long: `Detect the active shell and run the official sparks installer.

On macOS and Linux, sparks uses SPARKS_SHELL or SHELL and supports bash, zsh,
fish and POSIX-compatible shells. On Windows, PowerShell waits for the running
sparks process to exit before replacing it. The installer targets the directory
of the active executable and verifies the GoReleaser SHA-256 checksum.`,
		Example: `  sparks update`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			plan, err := client.Prepare()
			if err != nil {
				return fmt.Errorf("prepare sparks update: %w", err)
			}
			output.Message(stdout(cmd), "Detected %s", plan.Shell)
			output.Message(stdout(cmd), "Running: %s", plan.Command)
			if err := client.Execute(cmd.Context(), plan, stdout(cmd), cmd.ErrOrStderr()); err != nil {
				return fmt.Errorf("update sparks: %w", err)
			}
			if plan.Deferred {
				output.Message(stdout(cmd), "Installer started. It will replace sparks after this process exits")
				return nil
			}
			output.Message(stdout(cmd), "Update finished. Restart sparks to use the installed version")
			return nil
		},
	}
}
