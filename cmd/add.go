package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jokot/git-swap/internal/config"
	"github.com/spf13/cobra"
)

func newAddCmd(app *App) *cobra.Command {
	var p config.Profile
	c := &cobra.Command{
		Use:   "add [name]",
		Short: "Add or update an account profile",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				p.Name = args[0]
			}

			// Interactive mode if no name is provided (or if called with just `git-swap add`)
			if p.Name == "" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Fprintf(cmd.OutOrStdout(), "Interactive Add (leave blank for defaults)\n")
				
				p.Name = prompt(reader, cmd, "Profile name? (e.g. work, personal): ", "")
				if p.Name == "" {
					return fmt.Errorf("profile name is required")
				}
				
				p.Hub = prompt(reader, cmd, "Which hub? [github/gitlab/azure/custom] (default: github): ", "github")
				p.Auth = prompt(reader, cmd, "Auth mode? [ssh/https] (default: ssh): ", "ssh")
				
				p.GitName = prompt(reader, cmd, "Git user.name (commit name)?: ", "")
				p.GitEmail = prompt(reader, cmd, "Git user.email (commit email)?: ", "")

				if p.Auth == "ssh" {
					p.SSHKey = prompt(reader, cmd, "Path to SSH private key? (e.g. ~/.ssh/id_ed25519): ", "")
					if p.SSHKey == "" {
						return fmt.Errorf("ssh key is required for ssh auth")
					}
					
					signStr := prompt(reader, cmd, "Enable SSH commit signing? [y/N]: ", "N")
					if strings.ToLower(signStr) == "y" || strings.ToLower(signStr) == "yes" {
						p.Sign = true
						p.SigningKey = prompt(reader, cmd, "Path to SSH public key for signing? (e.g. ~/.ssh/id_ed25519.pub): ", p.SSHKey+".pub")
					}
				} else if p.Auth == "https" {
					p.Username = prompt(reader, cmd, "GitHub/GitLab username for credentials?: ", "")
					if p.Username == "" {
						return fmt.Errorf("username is required for https auth")
					}
					p.TokenEnv = prompt(reader, cmd, "Environment variable holding the PAT? (e.g. GH_PAT): ", "")
					if p.TokenEnv == "" {
						return fmt.Errorf("token env is required for https auth")
					}
				}
				fmt.Fprintf(cmd.OutOrStdout(), "\n")
			}

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
