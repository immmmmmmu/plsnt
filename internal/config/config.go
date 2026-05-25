package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	CurrentProfile string              `yaml:"current_profile"`
	Profiles       map[string]*Profile `yaml:"profiles"`
}

type Profile struct {
	URL        string `yaml:"url"`
	APIKey     string `yaml:"api_key"`
	APIVersion string `yaml:"api_version"`
}

func DefaultPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "plsnt", "config.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "plsnt", "config.yaml")
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Profiles: make(map[string]*Profile),
			}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]*Profile)
	}

	checkPermissions(path)

	return &cfg, nil
}

func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := ensureSecureWrite(path, data); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

// checkPermissions warns to stderr if the config file has overly permissive
// permissions (anything beyond owner read/write).
func checkPermissions(path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	mode := info.Mode().Perm()
	if mode&0077 != 0 {
		fmt.Fprintf(os.Stderr, "WARNING: config file %s has permissions %04o, should be 0600\n", path, mode)
	}
}

// ensureSecureWrite creates the parent directory with 0700 permissions and
// writes data to the file with 0600 permissions.
func ensureSecureWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (c *Config) ActiveProfile() (*Profile, string, error) {
	return c.ActiveProfileWithOverride("")
}

// ActiveProfileWithOverride returns the active profile. Priority:
// override (from -p flag) > PLSNT_PROFILE env > current_profile in config
func (c *Config) ActiveProfileWithOverride(override string) (*Profile, string, error) {
	name := override
	if name == "" {
		name = os.Getenv("PLSNT_PROFILE")
	}
	if name == "" {
		name = c.CurrentProfile
	}
	if name == "" {
		name = "default"
	}

	p, ok := c.Profiles[name]
	if !ok {
		return nil, name, fmt.Errorf("profile %q not found", name)
	}
	return p, name, nil
}

func (p *Profile) Resolve() (url, apiKey, apiVersion string) {
	url = os.Getenv("PLSNT_URL")
	if url == "" {
		url = p.URL
	}

	apiKey = os.Getenv("PLSNT_API_KEY")
	if apiKey == "" {
		apiKey = p.APIKey
	}

	apiVersion = p.APIVersion
	if apiVersion == "" {
		apiVersion = "1.1"
	}

	return url, apiKey, apiVersion
}

func MaskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return "****..." + key[len(key)-4:]
}
