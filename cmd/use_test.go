package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jokot/git-swap/internal/config"
	"github.com/jokot/git-swap/internal/run"
)

func newTestApp(t *testing.T, profiles ...config.Profile) *App {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := config.Save(cfgPath, &config.Config{Profiles: profiles}); err != nil {
		t.Fatal(err)
	}
	return &App{
		Runner:     &run.Fake{},
		ConfigPath: cfgPath,
		SSHPath:    filepath.Join(dir, "ssh_config"),
	}
}

func TestUseAppliesGlobalIdentity(t *testing.T) {
	app := newTestApp(t, config.Profile{Name: "work", Host: "github.com", GitName: "Jane", GitEmail: "jane@work.com"})
	var out bytes.Buffer
	root := NewRootCmd(app)
	root.SetOut(&out)
	root.SetArgs([]string{"use", "work"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	f := app.Runner.(*run.Fake)
	found := false
	for _, c := range f.Calls {
		if c == "git config --global user.email jane@work.com" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected global email set, calls: %v", f.Calls)
	}
}

func TestUseUnknownProfileErrors(t *testing.T) {
	app := newTestApp(t)
	root := NewRootCmd(app)
	root.SetArgs([]string{"use", "ghost"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for unknown profile")
	}
}

func TestUseHTTPSSetsCredentialUsername(t *testing.T) {
	t.Setenv("WORK_PAT", "ghp_testtoken")
	app := newTestApp(t, config.Profile{
		Name: "work", Host: "github.com", Auth: "https",
		GitName: "Jane", GitEmail: "jane@work.com",
		Username: "jane-work", TokenEnv: "WORK_PAT",
	})
	root := NewRootCmd(app)
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"use", "work"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	f := app.Runner.(*run.Fake)
	found := false
	for _, c := range f.Calls {
		if c == "git config --global credential.https://github.com.username jane-work" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected credential username set, calls: %v", f.Calls)
	}
	// The token must NOT appear as a bare Run arg (it is piped via stdin, not argv).
	for _, c := range f.Calls {
		if strings.Contains(c, "ghp_testtoken") {
			t.Fatalf("token leaked into argv: %v", c)
		}
	}
}
