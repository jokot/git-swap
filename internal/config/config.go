package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	Name       string `yaml:"name"`
	Hub        string `yaml:"hub"`            // github | gitlab | azure | custom
	Host       string `yaml:"host"`           // github.com, gitlab.com, ssh.dev.azure.com
	Auth       string `yaml:"auth,omitempty"` // ssh | https (default ssh)
	GitName    string `yaml:"git_name"`
	GitEmail   string `yaml:"git_email"`
	SSHKey     string `yaml:"ssh_key,omitempty"`    // used when Auth==ssh
	Username   string `yaml:"username,omitempty"`   // HTTPS credential username (Auth==https)
	TokenEnv   string `yaml:"token_env,omitempty"`  // env var holding PAT (Auth==https)
	TokenFile  string `yaml:"token_file,omitempty"` // file holding PAT (Auth==https)
	SigningKey string `yaml:"signing_key,omitempty"`
	Sign       bool   `yaml:"sign,omitempty"`
}

// AuthMode returns the effective auth mode, defaulting to "ssh".
func (p Profile) AuthMode() string {
	if p.Auth == "https" {
		return "https"
	}
	return "ssh"
}

type Config struct {
	Profiles []Profile `yaml:"profiles"`
}

// SetDefaults populates omitted fields and expands ~ in paths.
func (p *Profile) SetDefaults() {
	if p.Auth == "" {
		p.Auth = "ssh"
	}
	if p.Host == "" {
		p.Host = p.Hub
		if p.Host != "github" && p.Host != "gitlab" && p.Host != "azure" {
			// keep custom host as is
		} else {
			if p.Host == "azure" {
				p.Host = "dev.azure.com"
			} else {
				p.Host = p.Host + ".com"
			}
		}
	}
	p.SSHKey = expandTilde(p.SSHKey)
	p.SigningKey = expandTilde(p.SigningKey)
}

func expandTilde(path string) string {
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	// ~ or ~/path or ~\path
	if len(path) == 1 {
		return home
	}
	if path[1] == '/' || path[1] == '\\' {
		return filepath.Join(home, path[2:])
	}
	// ~user/path is unsupported, return as is
	return path
}

// Validate checks required fields.
func (p *Profile) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	if p.Hub == "" {
		return fmt.Errorf("hub is required")
	}
	if p.GitEmail == "" {
		return fmt.Errorf("git_email is required")
	}
	if p.Auth != "ssh" && p.Auth != "https" {
		return fmt.Errorf("auth must be 'ssh' or 'https'")
	}
	if p.Auth == "https" {
		if p.Username == "" {
			return fmt.Errorf("username is required for https auth")
		}
		if p.TokenEnv == "" && p.TokenFile == "" {
			return fmt.Errorf("token_env or token_file is required for https auth")
		}
	}
	return nil
}

func (c *Config) Find(name string) (Profile, bool) {
	for _, p := range c.Profiles {
		if p.Name == name {
			return p, true
		}
	}
	return Profile{}, false
}

func (c *Config) Upsert(p Profile) {
	for i := range c.Profiles {
		if c.Profiles[i].Name == p.Name {
			c.Profiles[i] = p
			return
		}
	}
	c.Profiles = append(c.Profiles, p)
}

func (c *Config) Remove(name string) bool {
	for i := range c.Profiles {
		if c.Profiles[i].Name == name {
			c.Profiles = append(c.Profiles[:i], c.Profiles[i+1:]...)
			return true
		}
	}
	return false
}

// DefaultPath returns ~/.config/git-swap/config.yaml (honors XDG_CONFIG_HOME).
func DefaultPath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "git-swap", "config.yaml"), nil
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func Save(path string, c *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
