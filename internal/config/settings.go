package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Scanner  ScannerConfig  `toml:"scanner"`
	Registry RegistryConfig `toml:"registry"`
}

type ScannerConfig struct {
	Directories []string `toml:"directories"`
	Exclude     []string `toml:"exclude"`
}

type RegistryConfig struct {
	Aliases map[string]string `toml:"aliases"`
}

var BuiltinAliases = map[string]string{
	"docker.io": "registry-1.docker.io",
}

// lists config paths in priority order: XDG_CONFIG_HOME, $HOME/.config, /etc
func defaultConfigCandidates() []string {
	var candidates []string
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		candidates = append(candidates, filepath.Join(xdg, "kluzo", "config.toml"))
	} else if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".config", "kluzo", "config.toml"))
	}
	return append(candidates, "/etc/kluzo/config.toml")
}

// Pick the config to use, path if provided, or first existing default path
func resolveConfigPath(path string) (string, error) {
	var candidates []string
	if path == "" {
		candidates = defaultConfigCandidates()
	} else {
		candidates = []string{path}
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no config file found in %s", strings.Join(candidates, ", "))
}

func Load(path string) (Config, error) {
	path, err := resolveConfigPath(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, fmt.Errorf("could not load config from %s: %v", path, err)
	}
	// Validate the patterns on load so we can safely ignore ErrBadPattern later
	for _, pattern := range cfg.Scanner.Exclude {
		if _, err := filepath.Match(pattern, ""); err != nil {
			return Config{}, fmt.Errorf("invalid exclude pattern %q: %v", pattern, err)
		}
	}
	merged := make(map[string]string)
	for k, v := range BuiltinAliases {
		merged[k] = v
	}
	// User-defined aliases overwrite the built-in aliases
	for k, v := range cfg.Registry.Aliases {
		merged[k] = v
	}
	cfg.Registry.Aliases = merged
	return cfg, nil
}
