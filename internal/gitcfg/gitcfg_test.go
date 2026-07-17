package gitcfg

import (
	"testing"

	"github.com/jokot/git-swap/internal/config"
	"github.com/jokot/git-swap/internal/run"
)

func TestApplyGlobalIdentity(t *testing.T) {
	f := &run.Fake{}
	p := config.Profile{GitName: "Jane", GitEmail: "jane@work.com"}
	if err := Apply(f, p, false); err != nil {
		t.Fatalf("apply: %v", err)
	}
	assertContains(t, f.Calls, "git config --global user.name Jane")
	assertContains(t, f.Calls, "git config --global user.email jane@work.com")
}

func TestApplyLocalScope(t *testing.T) {
	f := &run.Fake{}
	p := config.Profile{GitName: "Jane", GitEmail: "jane@work.com"}
	if err := Apply(f, p, true); err != nil {
		t.Fatalf("apply: %v", err)
	}
	assertContains(t, f.Calls, "git config --local user.name Jane")
}

func TestApplySigningWhenEnabled(t *testing.T) {
	f := &run.Fake{}
	p := config.Profile{GitName: "J", GitEmail: "j@x.com", Sign: true, SigningKey: "ABC123"}
	if err := Apply(f, p, false); err != nil {
		t.Fatalf("apply: %v", err)
	}
	assertContains(t, f.Calls, "git config --global user.signingkey ABC123")
	assertContains(t, f.Calls, "git config --global commit.gpgsign true")
}

func TestApplyNoSigningWhenDisabled(t *testing.T) {
	f := &run.Fake{}
	p := config.Profile{GitName: "J", GitEmail: "j@x.com", Sign: false}
	_ = Apply(f, p, false)
	for _, c := range f.Calls {
		if c == "git config --global commit.gpgsign true" {
			t.Fatal("should not enable signing when Sign=false")
		}
	}
}

func assertContains(t *testing.T, calls []string, want string) {
	t.Helper()
	for _, c := range calls {
		if c == want {
			return
		}
	}
	t.Fatalf("call %q not found in %v", want, calls)
}
