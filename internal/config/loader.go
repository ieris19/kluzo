package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

func readConfig(path string) (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, fmt.Errorf("could not load config from %s: %v", path, err)
	}
	return cfg, nil
}

func Load(path string) (Config, error) {
	path, err := resolveConfigPath(path)
	if err != nil {
		return Config{}, err
	}
	cfg, err := readConfig(path)
	if err != nil {
		return Config{}, err
	}
	err = validateConfig(cfg)
	if err != nil {
		return Config{}, err
	}
	cfg = applyDefaults(cfg)
	return cfg, nil
}
