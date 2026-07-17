package cmd

import (
	"fmt"

	"github.com/jokot/git-swap/internal/gitcfg"
	"github.com/spf13/cobra"
)

func newCurrentCmd(app *App) *cobra.Command {
	var local bool
	c := &cobra.Command{
		Use:   "current",
		Short: "Show the active git identity and matching profile",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.load()
			if err != nil {
				return err
			}
			name, email := gitcfg.Current(app.Runner, local)
			match := "(no matching profile)"
			for _, p := range cfg.Profiles {
				if p.GitEmail == email && email != "" {
					match = p.Name
					break
				}
			}
			scope := "global"
			if local {
				scope = "local"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %s <%s> -> profile %s\n", scope, name, email, match)
			return nil
		},
	}
	c.Flags().BoolVar(&local, "local", false, "read the current repository scope")
	return c
}
