package cmd

import (
	"github.com/jokot/git-swap/internal/config"
	"github.com/jokot/git-swap/internal/run"
	"github.com/spf13/cobra"
)

// App holds shared dependencies for commands (injectable for tests).
type App struct {
	Runner     run.Runner
	ConfigPath string
	SSHPath    string
}

// Version is injected at build time by GoReleaser
var Version = "dev"

func (a *App) load() (*config.Config, error) { return config.Load(a.ConfigPath) }
func (a *App) save(c *config.Config) error   { return config.Save(a.ConfigPath, c) }

func NewRootCmd(app *App) *cobra.Command {
	root := &cobra.Command{
		Use:     "git-swap",
		Short:   "Quickly switch between git accounts across GitHub, GitLab, Azure DevOps, and more",
		Version: Version,
	}
	root.AddCommand(
		newAddCmd(app),
		newImportCmd(app),
		newListCmd(app),
		newRemoveCmd(app),
		newUninstallCmd(app),
		newUseCmd(app),
		newCurrentCmd(app),
		newDoctorCmd(app),
	)
	return root
}
