package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jokot/git-swap/internal/httpcfg"
	"github.com/spf13/cobra"
)

func newDoctorCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check profiles for missing SSH keys / unresolvable tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.load()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(cfg.Profiles) == 0 {
				fmt.Fprintln(out, "no profiles configured")
				return nil
			}
			for _, p := range cfg.Profiles {
				status := "ok"
				switch p.AuthMode() {
				case "ssh":
					if p.SSHKey != "" {
						path := p.SSHKey
						if strings.HasPrefix(path, "~/") {
							home, _ := os.UserHomeDir()
							path = filepath.Join(home, path[2:])
						}
						if _, err := os.Stat(path); err != nil {
							status = "MISSING ssh key: " + p.SSHKey
						}
					} else {
						status = "no ssh_key set"
					}
				case "https":
					if _, err := httpcfg.ResolveToken(p); err != nil {
						status = "TOKEN unresolved: " + err.Error()
					} else {
						status = "ok (token resolvable)"
					}
				}
				fmt.Fprintf(out, "%-16s [%s] %s\n", p.Name, p.AuthMode(), status)
			}
			return nil
		},
	}
}
