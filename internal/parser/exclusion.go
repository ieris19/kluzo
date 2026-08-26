package parser

import (
	"fmt"
	"path/filepath"
	"slices"
)

func isSupportedExtension(extension string, allowedExtensions []string) bool {
	return slices.Contains(allowedExtensions, extension)
}

func isExcluded(name string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		matched, err := filepath.Match(pattern, name)
		if err != nil {
			// config.Load already rejects malformed patterns
			// An error here is a violated invariant
			panic(fmt.Sprintf("exclude pattern %q passed config validation but failed to match: %v", pattern, err))
		}
		if matched {
			return true
		}
	}
	return false
}
