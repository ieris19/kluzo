package config

import (
    "fmt"

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

func defaultConfigPath() string {
    return "/etc/update-link/config.toml"
}

func Load(path string) (Config, error) {
    if path == "" {
        path = defaultConfigPath()
    }

    var cfg Config
    if _, err := toml.DecodeFile(path, &cfg); err != nil {
        return Config{}, fmt.Errorf("could not load config from %s: %w", path, err)
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
