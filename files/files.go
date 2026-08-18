package files

import (
    "fmt"
    "io/fs"
    "path/filepath"
    "strings"

    "git.ierislabs.dev/ieris19/update-link/config"
    "git.ierislabs.dev/ieris19/update-link/data"
)

func isSupportedExtension(extension string, allowedExtensions []string) bool {
    for _, ext := range allowedExtensions {
        if ext == extension {
            return true
        }
    }
    return false
}

func isExcluded(name string, excludePatterns []string) bool {
    for _, pattern := range excludePatterns {
        if matched, err := filepath.Match(pattern, name); err == nil && matched {
            return true
        }
    }
    return false
}

func GetAllFiles(settings config.ScannerConfig, allowedExtensions []string) []data.FileEntry {
    var containerFiles []data.FileEntry

    for _, dir := range settings.Directories {
        err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
            if err != nil {
                fmt.Printf("Error accessing %s: %v\n", path, err)
                return nil
            }
            if d.IsDir() {
                if path != dir && isExcluded(d.Name(), settings.Exclude) {
                    return fs.SkipDir
                }
                return nil
            }
            if isSupportedExtension(filepath.Ext(d.Name()), allowedExtensions) {
                containerFiles = append(containerFiles, data.NewFileEntry(filepath.Dir(path), d))
            }
            return nil
        })
        if err != nil {
            fmt.Printf("Error walking directory %s: %v\n", dir, err)
        }
    }

    return containerFiles
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
