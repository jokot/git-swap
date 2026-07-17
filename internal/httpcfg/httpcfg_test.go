package httpcfg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jokot/git-swap/internal/config"
	"github.com/jokot/git-swap/internal/run"
)

func TestResolveTokenFromEnv(t *testing.T) {
	t.Setenv("MY_PAT", "tok-env")
	p := config.Profile{TokenEnv: "MY_PAT"}
	got, err := ResolveToken(p)
	if err != nil || got != "tok-env" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestResolveTokenFromFileTrimsWhitespace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pat")
	if err := os.WriteFile(path, []byte("  tok-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	p := config.Profile{TokenFile: path}
	got, err := ResolveToken(p)
	if err != nil || got != "tok-file" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestResolveTokenEnvBeatsFile(t *testing.T) {
	t.Setenv("MY_PAT", "tok-env")
	p := config.Profile{TokenEnv: "MY_PAT", TokenFile: "/does/not/matter"}
	got, _ := ResolveToken(p)
	if got != "tok-env" {
		t.Fatalf("env should win, got %q", got)
	}
}

func TestResolveTokenMissingErrors(t *testing.T) {
	if _, err := ResolveToken(config.Profile{TokenEnv: "DEFINITELY_UNSET_VAR_XYZ"}); err == nil {
		t.Fatal("expected error when env var is empty/unset")
	}
	if _, err := ResolveToken(config.Profile{}); err == nil {
		t.Fatal("expected error when no token source configured")
	}
}

func TestApplySetsUsernameAndSeedsToken(t *testing.T) {
	t.Setenv("MY_PAT", "ghp_secret")
	f := &run.Fake{}
	p := config.Profile{Name: "work", Host: "github.com", Auth: "https", Username: "jane", TokenEnv: "MY_PAT"}
	if err := Apply(f, p, false); err != nil {
		t.Fatalf("apply: %v", err)
	}
	wantCfg := "git config --global credential.https://github.com.username jane"
	found := false
	for _, c := range f.Calls {
		if c == wantCfg {
			found = true
		}
	}
	if !found {
		t.Fatalf("username not set, calls: %v", f.Calls)
	}
	approvedViaStdin := false
	for _, c := range f.Calls {
		if c == "git credential approve" {
			approvedViaStdin = true
		}
	}
	if !approvedViaStdin {
		t.Fatalf("expected `git credential approve`, calls: %v", f.Calls)
	}
	if len(f.Inputs) == 0 {
		t.Fatalf("no stdin captured for credential approve")
	}
	joinedInputs := strings.Join(f.Inputs, "\n")
	if !strings.Contains(joinedInputs, "password=ghp_secret") ||
		!strings.Contains(joinedInputs, "host=github.com") ||
		!strings.Contains(joinedInputs, "username=jane") {
		t.Fatalf("credential stdin payload wrong: %q", joinedInputs)
	}
	for _, c := range f.Calls {
		if strings.Contains(c, "ghp_secret") {
			t.Fatalf("token leaked into argv: %q", c)
		}
	}
}

func TestApplyLocalScopeUsesLocalFlag(t *testing.T) {
	t.Setenv("MY_PAT", "ghp_secret")
	f := &run.Fake{}
	p := config.Profile{Name: "work", Host: "github.com", Auth: "https", Username: "jane", TokenEnv: "MY_PAT"}
	if err := Apply(f, p, true); err != nil {
		t.Fatalf("apply: %v", err)
	}
	found := false
	for _, c := range f.Calls {
		if c == "git config --local credential.https://github.com.username jane" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected --local scope, calls: %v", f.Calls)
	}
}
