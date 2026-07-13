package cmd

import (
	"fmt"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/output"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/updater"
	"github.com/spf13/cobra"
)

func newUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update sparks to the latest release",
		Long: `Download and install the latest sparks release from GitHub.

The update is installed only after its SHA-256 checksum matches the checksum
published by GoReleaser. The executable directory must be writable.`,
		Example: `  sparks update`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client := updater.New(version).OnProgress(func(stage updater.ProgressStage) {
				messages := map[updater.ProgressStage]string{
					updater.ProgressChecking:    "Checking for updates...",
					updater.ProgressDownloading: "Downloading update...",
					updater.ProgressVerifying:   "Verifying checksum...",
					updater.ProgressInstalling:  "Installing update...",
				}
				output.Message(stdout(cmd), "%s", messages[stage])
			})
			result, err := client.Update(cmd.Context())
			if err != nil {
				return fmt.Errorf("update sparks: %w", err)
			}
			if !result.Updated {
				output.Message(stdout(cmd), "sparks %s is already up to date", result.CurrentVersion)
				return nil
			}
			output.Message(stdout(cmd), "Updated sparks from %s to %s. Restart sparks to use the new version", result.CurrentVersion, result.LatestVersion)
			return nil
		},
	}
}
