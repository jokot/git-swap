package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jokot/git-swap/internal/config"
	"github.com/jokot/git-swap/internal/sshcfg"
	"github.com/spf13/cobra"
)

func newUninstallCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Clean up git-swap configuration and SSH blocks before removing the CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Remove the managed SSH block by syncing an empty profile list
			if err := sshcfg.SyncFile(app.SSHPath, nil); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to clean SSH config: %v\n", err)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ Removed git-swap block from %s\n", app.SSHPath)
			}

			// 2. Remove the config directory
			cfgPath, err := config.DefaultPath()
			if err == nil {
				cfgDir := filepath.Dir(cfgPath)
				if err := os.RemoveAll(cfgDir); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to remove config directory: %v\n", err)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "✓ Removed configuration directory %s\n", cfgDir)
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\nState cleaned! To complete uninstallation, remove the binary:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  - If installed via Homebrew: brew uninstall git-swap\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  - If installed via Winget:   winget uninstall jokot.git-swap\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  - If installed via Go:       rm $(which git-swap)\n")
			return nil
		},
	}
}
