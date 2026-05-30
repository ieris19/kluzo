package config

import (
    "fmt"

    "github.com/BurntSushi/toml"
)

type Config struct {
    Scanner ScannerConfig `toml:"scanner"`
}

type ScannerConfig struct {
    Dirs []string `toml:"dirs"`
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
    return cfg, nil
}
