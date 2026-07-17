package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jokot/git-swap/internal/config"
	"github.com/jokot/git-swap/internal/run"
)

func TestCurrentMatchesProfileByEmail(t *testing.T) {
	app := newTestApp(t, config.Profile{Name: "work", GitEmail: "jane@work.com"})
	app.Runner = &run.Fake{Outputs: map[string]string{
		"git config --global user.name":  "Jane",
		"git config --global user.email": "jane@work.com",
	}}
	var out bytes.Buffer
	root := NewRootCmd(app)
	root.SetOut(&out)
	root.SetArgs([]string{"current"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out.String(), "work") || !strings.Contains(out.String(), "jane@work.com") {
		t.Fatalf("output missing match: %q", out.String())
	}
}
