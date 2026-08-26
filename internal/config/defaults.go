package config

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
)

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

func applyDefaults(cfg Config) Config {
	merged := make(map[string]string)
	// Add defaults
	maps.Copy(merged, BuiltinAliases)
	// User-defined aliases overwrite the built-in aliases
	maps.Copy(merged, cfg.Registry.Aliases)
	// Overwrite user defined with merged set
	cfg.Registry.Aliases = merged
	return cfg
}
