package config

import (
	"fmt"
	"path/filepath"
)

func validateExclusionPatterns(patterns []string) error {
	// Validate the patterns on load so we can safely ignore ErrBadPattern later
	for _, pattern := range patterns {
		if _, err := filepath.Match(pattern, ""); err != nil {
			return fmt.Errorf("invalid exclude pattern %q: %v", pattern, err)
		}
	}
	return nil
}

func validateConfig(cfg Config) error {
	if err := validateExclusionPatterns(cfg.Scanner.Exclude); err != nil {
		return err
	}
	return nil
}
