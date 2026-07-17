package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRemoveCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an account profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cfg, err := app.load()
			if err != nil {
				return err
			}
			if !cfg.Remove(name) {
				return fmt.Errorf("no profile named %q", name)
			}
			if err := app.save(cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed %q\n", name)
			return nil
		},
	}
}
