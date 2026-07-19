package cmd

import (
	"fmt"

	"github.com/jokot/git-swap/internal/config"
	"github.com/spf13/cobra"
)

func newAddCmd(app *App) *cobra.Command {
	var p config.Profile
	c := &cobra.Command{
		Use:   "add <name>",
		Short: "Add or update an account profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p.Name = args[0]
			if p.GitEmail == "" {
				return fmt.Errorf("--email is required")
			}
			p.SetDefaults()

			if err := p.Validate(); err != nil {
				return err
			}

			cfg, err := app.load()
			if err != nil {
				return err
			}
			cfg.Upsert(p)
			if err := app.save(cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "saved profile %q (%s, %s)\n", p.Name, p.Host, p.Auth)
			return nil
		},
	}
	c.Flags().StringVar(&p.Hub, "hub", "", "hub type: github|gitlab|azure|custom")
	c.Flags().StringVar(&p.Host, "host", "", "override host (e.g. github.com)")
	c.Flags().StringVar(&p.Auth, "auth", "ssh", "auth mode: ssh|https")
	c.Flags().StringVar(&p.GitName, "name", "", "git user.name")
	c.Flags().StringVar(&p.GitEmail, "email", "", "git user.email (required)")
	c.Flags().StringVar(&p.SSHKey, "ssh-key", "", "path to ssh private key (auth=ssh)")
	c.Flags().StringVar(&p.Username, "username", "", "HTTPS credential username (auth=https)")
	c.Flags().StringVar(&p.TokenEnv, "token-env", "", "env var holding the PAT (auth=https)")
	c.Flags().StringVar(&p.TokenFile, "token-file", "", "file holding the PAT (auth=https)")
	c.Flags().StringVar(&p.SigningKey, "signing-key", "", "commit signing key id/path")
	c.Flags().BoolVar(&p.Sign, "sign", false, "enable commit signing")
	return c
}

// defaultHost picks a sensible host per hub. Azure differs by auth mode
// (ssh.dev.azure.com for SSH vs dev.azure.com for HTTPS).
func defaultHost(hub, auth string) string {
	switch hub {
	case "github":
		return "github.com"
	case "gitlab":
		return "gitlab.com"
	case "azure":
		if auth == "https" {
			return "dev.azure.com"
		}
		return "ssh.dev.azure.com"
	default:
		return ""
	}
}
