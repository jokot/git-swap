package sshcfg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jokot/git-swap/internal/config"
)

const (
	beginMarker = "# >>> git-swap managed block >>>"
	endMarker   = "# <<< git-swap managed block <<<"
)

// AliasFor returns the SSH Host alias for a profile, e.g. "github.com-work".
func AliasFor(p config.Profile) string {
	return fmt.Sprintf("%s-%s", p.Host, p.Name)
}

// RenderBlock returns a marker-delimited SSH config block for all profiles
// that declare an SSHKey. Profiles without a key are skipped.
func RenderBlock(profiles []config.Profile) string {
	var b strings.Builder
	var hasSSH bool
	for _, p := range profiles {
		if p.SSHKey != "" && p.Host != "" {
			hasSSH = true
			break
		}
	}
	if !hasSSH {
		return ""
	}

	b.WriteString(beginMarker + "\n")
	for _, p := range profiles {
		if p.SSHKey == "" || p.Host == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("Host %s\n", AliasFor(p)))
		b.WriteString(fmt.Sprintf("    HostName %s\n", p.Host))
		b.WriteString("    User git\n")
		b.WriteString(fmt.Sprintf("    IdentityFile %s\n", p.SSHKey))
		b.WriteString("    IdentitiesOnly yes\n")
	}
	b.WriteString(endMarker + "\n")
	return b.String()
}

// Splice replaces an existing managed block in `existing` with `block`,
// or appends `block` if no managed block is present. User content is preserved.
// If `block` is empty, the managed block is completely removed.
func Splice(existing, block string) string {
	start := strings.Index(existing, beginMarker)
	end := strings.Index(existing, endMarker)
	if start != -1 && end != -1 && end > start {
		end += len(endMarker)
		// consume a trailing newline after the end marker if present
		tail := existing[end:]
		tail = strings.TrimPrefix(tail, "\n")
		
		before := existing[:start]
		// if we are removing the block completely, consume a trailing newline before it too
		if block == "" {
			before = strings.TrimSuffix(before, "\n")
			return before + "\n" + tail
		}
		
		return before + strings.TrimSuffix(block, "\n") + "\n" + tail
	}
	if block == "" {
		return existing
	}
	sep := ""
	if existing != "" && !strings.HasSuffix(existing, "\n") {
		sep = "\n"
	}
	if existing != "" {
		sep += "\n"
	}
	return existing + sep + block
}

// SyncFile reads the ssh config at path (if any), splices in the managed block
// for the given profiles, and writes the result with 0600 perms.
func SyncFile(path string, profiles []config.Profile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	var existing string
	if data, err := os.ReadFile(path); err == nil {
		existing = string(data)
	} else if !os.IsNotExist(err) {
		return err
	}
	merged := Splice(existing, RenderBlock(profiles))
	return os.WriteFile(path, []byte(merged), 0o600)
}

// DefaultPath returns ~/.ssh/config.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh", "config"), nil
}
