package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List saved account profiles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.load()
			if err != nil {
				return err
			}
			if len(cfg.Profiles) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no profiles yet — add one with: git-swap add <name> --email <e> --hub github")
				return nil
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tHUB\tAUTH\tHOST\tEMAIL\tCRED\tSIGN")
			for _, p := range cfg.Profiles {
				cred := p.SSHKey
				if p.AuthMode() == "https" {
					cred = "user:" + p.Username
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%v\n",
					p.Name, p.Hub, p.AuthMode(), p.Host, p.GitEmail, cred, p.Sign)
			}
			return w.Flush()
		},
	}
}
