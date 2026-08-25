package files

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	"git.ierislabs.dev/ieris19/kluzo/internal/config"
	"git.ierislabs.dev/ieris19/kluzo/internal/data"
)

func isSupportedExtension(extension string, allowedExtensions []string) bool {
	return slices.Contains(allowedExtensions, extension)
}

func isExcluded(name string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		if matched, err := filepath.Match(pattern, name); err == nil && matched {
			return true
		}
	}
	return false
}

func GetAllFiles(settings config.ScannerConfig, allowedExtensions []string) ([]data.FileEntry, error) {
	var containerFiles []data.FileEntry
	var errs []error

	for _, dir := range settings.Directories {
		walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			// Returning nil here means continue to the next file
			// Returning an error WalkDir doesn't expect aborts the search
			if err != nil {
				errs = append(errs, fmt.Errorf("accessing %s: %v", path, err))
				return nil
			}
			// Patterns match against the path relative to the scan root
			rel, relErr := filepath.Rel(dir, path)
			if relErr != nil {
				rel = path
			}
			excluded := isExcluded(rel, settings.Exclude)

			if d.IsDir() {
				if path != dir && excluded {
					// WalkDir will swallow this error and won't propagate, safe to use here
					return fs.SkipDir
				}
				return nil
			}
			if excluded || !isSupportedExtension(filepath.Ext(d.Name()), allowedExtensions) {
				return nil
			}
			containerFiles = append(containerFiles, data.NewFileEntry(filepath.Dir(path), d))
			return nil
		})
		// Should not trigger with current code, kept in case a future change aborts the search
		if walkErr != nil {
			errs = append(errs, fmt.Errorf("walking %s: %v", dir, walkErr))
		}
	}

	return containerFiles, errors.Join(errs...)
}

func parseFileLines(content string) []string {
	lines := strings.Split(content, "\n")
	var parsedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parsedLines = append(parsedLines, trimmed)
	}
	return parsedLines
}

func sectionHeader(line string) (string, bool) {
	if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
		return line[1 : len(line)-1], true
	}
	return "", false
}

func ParseFileKeyValue(content string, separator string) map[string]map[string]string {
	lines := parseFileLines(content)
	sections := make(map[string]map[string]string)
	current := ""
	sections[current] = make(map[string]string)
	for _, line := range lines {
		if name, ok := sectionHeader(line); ok {
			current = name
			if _, exists := sections[current]; !exists {
				sections[current] = make(map[string]string)
			}
			continue
		}
		key, value, found := strings.Cut(line, separator)
		if found {
			sections[current][strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	return sections
}
