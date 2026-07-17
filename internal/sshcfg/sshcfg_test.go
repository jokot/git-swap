package sshcfg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jokot/git-swap/internal/config"
)

func TestRenderBlockIncludesAliasAndKey(t *testing.T) {
	profiles := []config.Profile{
		{Name: "work", Host: "github.com", SSHKey: "~/.ssh/id_work"},
		{Name: "home", Host: "gitlab.com", SSHKey: "~/.ssh/id_home"},
	}
	block := RenderBlock(profiles)
	if !strings.Contains(block, beginMarker) || !strings.Contains(block, endMarker) {
		t.Fatal("block must be delimited by markers")
	}
	if !strings.Contains(block, "Host github.com-work") {
		t.Fatalf("missing work alias:\n%s", block)
	}
	if !strings.Contains(block, "IdentityFile ~/.ssh/id_work") {
		t.Fatalf("missing work identity file:\n%s", block)
	}
	if !strings.Contains(block, "HostName gitlab.com") {
		t.Fatalf("missing home hostname:\n%s", block)
	}
}

func TestRenderBlockSkipsProfilesWithoutKey(t *testing.T) {
	block := RenderBlock([]config.Profile{{Name: "nokeys", Host: "github.com"}})
	if strings.Contains(block, "Host github.com-nokeys") {
		t.Fatal("profiles without ssh_key must be skipped")
	}
}

func TestAliasFor(t *testing.T) {
	p := config.Profile{Name: "work", Host: "github.com"}
	if got := AliasFor(p); got != "github.com-work" {
		t.Fatalf("got %q", got)
	}
}

func TestSpliceAppendsWhenAbsent(t *testing.T) {
	existing := "Host example\n    HostName example.com\n"
	block := "# >>> git-swap managed block >>>\nHost x\n# <<< git-swap managed block <<<\n"
	out := Splice(existing, block)
	if !strings.Contains(out, "Host example") {
		t.Fatal("must preserve user content")
	}
	if !strings.Contains(out, "Host x") {
		t.Fatal("must include new block")
	}
}

func TestSpliceReplacesExisting(t *testing.T) {
	existing := "top\n" + beginMarker + "\nOLD\n" + endMarker + "\nbottom\n"
	block := beginMarker + "\nNEW\n" + endMarker + "\n"
	out := Splice(existing, block)
	if strings.Contains(out, "OLD") {
		t.Fatalf("old block not replaced:\n%s", out)
	}
	if !strings.Contains(out, "NEW") || !strings.Contains(out, "top") || !strings.Contains(out, "bottom") {
		t.Fatalf("splice lost content:\n%s", out)
	}
	if strings.Count(out, beginMarker) != 1 {
		t.Fatalf("expected exactly one managed block:\n%s", out)
	}
}

func TestSyncFileCreatesAndUpdates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	profiles := []config.Profile{{Name: "work", Host: "github.com", SSHKey: "~/.ssh/id_work"}}
	if err := SyncFile(path, profiles); err != nil {
		t.Fatalf("sync: %v", err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "Host github.com-work") {
		t.Fatalf("alias not written:\n%s", data)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("perms = %v, want 0600", info.Mode().Perm())
	}
	_ = SyncFile(path, profiles)
	data, _ = os.ReadFile(path)
	if strings.Count(string(data), beginMarker) != 1 {
		t.Fatalf("duplicate managed block after re-sync:\n%s", data)
	}
}
