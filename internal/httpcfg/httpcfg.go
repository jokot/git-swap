package httpcfg

import (
	"fmt"
	"os"
	"strings"

	"github.com/jokot/git-swap/internal/config"
	"github.com/jokot/git-swap/internal/run"
)

// ResolveToken returns the PAT for an HTTPS profile, read from the env var
// named by TokenEnv (preferred) or the file at TokenFile. Never persisted.
func ResolveToken(p config.Profile) (string, error) {
	if p.TokenEnv != "" {
		if v := os.Getenv(p.TokenEnv); v != "" {
			return v, nil
		}
		return "", fmt.Errorf("env var %q is empty or unset", p.TokenEnv)
	}
	if p.TokenFile != "" {
		data, err := os.ReadFile(p.TokenFile)
		if err != nil {
			return "", fmt.Errorf("read token_file: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return "", fmt.Errorf("profile %q has no token_env or token_file", p.Name)
}

// Apply configures the HTTPS credential username for the profile's host and
// seeds the PAT into git's credential helper. scopeLocal targets the repo.
func Apply(r run.Runner, p config.Profile, scopeLocal bool) error {
	token, err := ResolveToken(p)
	if err != nil {
		return err
	}
	scope := "--global"
	if scopeLocal {
		scope = "--local"
	}
	key := fmt.Sprintf("credential.https://%s.username", p.Host)
	if _, err := r.Run("git", "config", scope, key, p.Username); err != nil {
		return fmt.Errorf("set credential username: %w", err)
	}
	// Seed the token into git's credential helper. The payload is piped on
	// stdin so the secret never appears on the process argv.
	payload := fmt.Sprintf("protocol=https\nhost=%s\nusername=%s\npassword=%s\n\n",
		p.Host, p.Username, token)
	if _, err := r.RunInput(payload, "git", "credential", "approve"); err != nil {
		return fmt.Errorf("seed credential: %w", err)
	}
	return nil
}
