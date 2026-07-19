package gitcfg

import (
	"strings"

	"github.com/jokot/git-swap/internal/config"
	"github.com/jokot/git-swap/internal/run"
)

// Apply writes identity and signing settings for the profile.
// scopeLocal=true targets the current repo (.git/config); otherwise global.
func Apply(r run.Runner, p config.Profile, scopeLocal bool) error {
	scope := "--global"
	if scopeLocal {
		scope = "--local"
	}
	sets := [][2]string{
		{"user.name", p.GitName},
		{"user.email", p.GitEmail},
	}
	if p.Sign {
		if p.SigningKey != "" {
			sets = append(sets, [2]string{"user.signingkey", p.SigningKey})
		}
		sets = append(sets, [2]string{"commit.gpgsign", "true"})
	}
	for _, kv := range sets {
		if _, err := r.Run("git", "config", scope, kv[0], kv[1]); err != nil {
			return err
		}
	}
	return nil
}

// Current reads the effective identity for the given scope.
func Current(r run.Runner, scopeLocal bool) (name, email string) {
	scope := "--global"
	if scopeLocal {
		scope = "--local"
	}
	name, errName := r.Run("git", "config", scope, "user.name")
	if errName != nil {
		name = ""
	}
	email, errEmail := r.Run("git", "config", scope, "user.email")
	if errEmail != nil {
		email = ""
	}
	return strings.TrimSpace(name), strings.TrimSpace(email)
}
