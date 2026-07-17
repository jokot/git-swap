package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindProfile(t *testing.T) {
	c := &Config{Profiles: []Profile{{Name: "work"}, {Name: "personal"}}}
	p, ok := c.Find("personal")
	if !ok || p.Name != "personal" {
		t.Fatalf("expected to find personal, got %+v ok=%v", p, ok)
	}
	if _, ok := c.Find("missing"); ok {
		t.Fatal("should not find missing")
	}
}

func TestUpsertReplacesByName(t *testing.T) {
	c := &Config{}
	c.Upsert(Profile{Name: "work", GitEmail: "old@x.com"})
	c.Upsert(Profile{Name: "work", GitEmail: "new@x.com"})
	if len(c.Profiles) != 1 {
		t.Fatalf("want 1 profile, got %d", len(c.Profiles))
	}
	if c.Profiles[0].GitEmail != "new@x.com" {
		t.Fatalf("upsert did not replace: %+v", c.Profiles[0])
	}
}

func TestSaveThenLoadRoundTrips(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "config.yaml")
	c := &Config{Profiles: []Profile{{Name: "work", GitEmail: "w@x.com"}}}
	if err := Save(path, c); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(got.Profiles) != 1 || got.Profiles[0].GitEmail != "w@x.com" {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
}

func TestLoadMissingReturnsEmpty(t *testing.T) {
	got, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if err != nil {
		t.Fatalf("missing file should not error: %v", err)
	}
	if len(got.Profiles) != 0 {
		t.Fatalf("expected empty config, got %+v", got)
	}
	_ = os.Stat // keep os imported
}
