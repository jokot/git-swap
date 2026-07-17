package main

import (
	"fmt"
	"os"

	"github.com/jokot/git-swap/cmd"
	"github.com/jokot/git-swap/internal/config"
	"github.com/jokot/git-swap/internal/run"
	"github.com/jokot/git-swap/internal/sshcfg"
)

func main() {
	cfgPath, err := config.DefaultPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	sshPath, err := sshcfg.DefaultPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	app := &cmd.App{Runner: run.Exec{}, ConfigPath: cfgPath, SSHPath: sshPath}
	if err := cmd.NewRootCmd(app).Execute(); err != nil {
		os.Exit(1)
	}
}
