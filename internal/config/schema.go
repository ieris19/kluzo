package config

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
