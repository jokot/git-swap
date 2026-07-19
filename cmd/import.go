package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jokot/git-swap/internal/config"
	"github.com/jokot/git-swap/internal/gitcfg"
	"github.com/spf13/cobra"
)

func newImportCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "import <name>",
		Short: "Capture the currently active git identity into a new profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cfg, err := app.load()
			if err != nil {
				return err
			}
			if _, ok := cfg.Find(name); ok {
				return fmt.Errorf("profile %q already exists", name)
			}

			// Read identities
			globalName, globalEmail := gitcfg.Current(app.Runner, false)
			localName, localEmail := gitcfg.Current(app.Runner, true)

			var pName, pEmail string

			if localName != "" && localName != globalName || localEmail != "" && localEmail != globalEmail {
				fmt.Fprintf(cmd.OutOrStdout(), "Found two distinct git identities:\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  [1] Local (repo): %s <%s>\n", localName, localEmail)
				fmt.Fprintf(cmd.OutOrStdout(), "  [2] Global:       %s <%s>\n", globalName, globalEmail)
				fmt.Fprintf(cmd.OutOrStdout(), "Which one to import? [1/2, default 2]: ")
				
				var choice string
				fmt.Scanln(&choice)
				if choice == "1" {
					pName = localName
					pEmail = localEmail
				} else {
					pName = globalName
					pEmail = globalEmail
				}
			} else {
				if globalName == "" && globalEmail == "" {
					return fmt.Errorf("no git identity found in config")
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Found global git identity:\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  Name:  %s\n", globalName)
				fmt.Fprintf(cmd.OutOrStdout(), "  Email: %s\n", globalEmail)
				pName = globalName
				pEmail = globalEmail
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n")

			reader := bufio.NewReader(os.Stdin)

			hub := prompt(reader, cmd, "Which hub? [github/gitlab/azure/custom] (default: github): ", "github")
			auth := prompt(reader, cmd, "Auth mode? [ssh/https] (default: ssh): ", "ssh")
			
			var sshKey, username, tokenEnv string
			if auth == "ssh" {
				sshKey = prompt(reader, cmd, "Path to SSH private key? (e.g. ~/.ssh/id_ed25519): ", "")
				if sshKey == "" {
					return fmt.Errorf("ssh key is required for ssh auth")
				}
			} else if auth == "https" {
				username = prompt(reader, cmd, "GitHub/GitLab username for credentials?: ", "")
				if username == "" {
					return fmt.Errorf("username is required for https auth")
				}
				tokenEnv = prompt(reader, cmd, "Environment variable holding the PAT? (e.g. GH_PAT): ", "")
				if tokenEnv == "" {
					return fmt.Errorf("token env is required for https auth")
				}
			} else {
				return fmt.Errorf("invalid auth mode: %q", auth)
			}

			p := config.Profile{
				Name:     name,
				Hub:      hub,
				Auth:     auth,
				GitName:  pName,
				GitEmail: pEmail,
				SSHKey:   sshKey,
				Username: username,
				TokenEnv: tokenEnv,
			}

			if p.Auth == "" {
				p.Auth = "ssh"
			}
			p.Host = defaultHost(p.Hub, p.Auth)

			cfg.Profiles = append(cfg.Profiles, p)
			if err := app.save(cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\nProfile %q saved successfully!\n", name)
			return nil
		},
	}
}

func prompt(r *bufio.Reader, cmd *cobra.Command, text, def string) string {
	fmt.Fprintf(cmd.OutOrStdout(), "%s", text)
	input, _ := r.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return def
	}
	return input
}