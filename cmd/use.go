package cmd

import (
	"fmt"

	"github.com/jokot/git-swap/internal/config"
	"github.com/jokot/git-swap/internal/gitcfg"
	"github.com/jokot/git-swap/internal/httpcfg"
	"github.com/jokot/git-swap/internal/sshcfg"
	"github.com/spf13/cobra"
)

func newUseCmd(app *App) *cobra.Command {
	var local bool
	var rewriteRemote string
	c := &cobra.Command{
		Use:   "use <name>",
		Short: "Switch the active git account (global by default, --local for this repo)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.load()
			if err != nil {
				return err
			}
			p, ok := cfg.Find(args[0])
			if !ok {
				return fmt.Errorf("profile %q not found", args[0])
			}
			if err := gitcfg.Apply(app.Runner, p, local); err != nil {
				return fmt.Errorf("apply git config: %w", err)
			}

			switch p.AuthMode() {
			case "ssh":
				if err := sshcfg.SyncFile(app.SSHPath, cfg.Profiles); err != nil {
					return fmt.Errorf("sync ssh config: %w", err)
				}
			case "https":
				// Set credential username (disambiguates same-host accounts) and
				// seed the PAT into git's credential helper via stdin.
				if err := httpcfg.Apply(app.Runner, p, local); err != nil {
					return fmt.Errorf("apply https credentials: %w", err)
				}
			}

			if rewriteRemote != "" {
				url := remoteURL(p, rewriteRemote)
				if url == "" {
					return fmt.Errorf("cannot build remote for profile %q", p.Name)
				}
				if _, err := app.Runner.Run("git", "remote", "set-url", "origin", url); err != nil {
					return fmt.Errorf("rewrite remote: %w", err)
				}
			}
			scope := "global"
			if local {
				scope = "local"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "switched to %q (%s, %s) — %s <%s>\n", p.Name, scope, p.AuthMode(), p.GitName, p.GitEmail)
			return nil
		},
	}
	c.Flags().BoolVar(&local, "local", false, "apply to the current repository only")
	c.Flags().StringVar(&rewriteRemote, "rewrite-remote", "", "rewrite origin for <owner/repo> using this profile's auth mode")
	return c
}

// remoteURL builds the origin URL for a profile+repo per its auth mode.
//
//	ssh:   git@<host>-<name>:<owner/repo>.git
//	https: https://<user>@<host>/<owner/repo>.git
func remoteURL(p config.Profile, ownerRepo string) string {
	switch p.AuthMode() {
	case "ssh":
		if p.SSHKey == "" {
			return ""
		}
		return fmt.Sprintf("git@%s:%s.git", sshcfg.AliasFor(p), ownerRepo)
	case "https":
		if p.Username == "" {
			return ""
		}
		return fmt.Sprintf("https://%s@%s/%s.git", p.Username, p.Host, ownerRepo)
	default:
		return ""
	}
}
