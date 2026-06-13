package files

import (
    "fmt"
    "os"
    "strings"

    "git.ierislabs.dev/update-link/data"
)

func isSupportedExtension(extension string, allowedExtensions []string) bool {
    for _, ext := range allowedExtensions {
        if ext == extension {
            return true
        }
    }
    return false
}

func GetAllFiles(dirs []string, allowedExtensions []string) []data.FileEntry {
    var files []data.FileEntry

    // Traverse each root directory for container definitions
    for _, dir := range dirs {
        entries, err := os.ReadDir(dir)
        if err != nil {
            fmt.Printf("Error reading directory %s: %v\n", dir, err)
            continue
        }
        files = append(files, data.NewFileEntries(dir, entries)...)
    }

    // Filter files by supported extensions
    var containerFiles []data.FileEntry
    for _, file := range files {

        if isSupportedExtension(file.Extension, allowedExtensions) {
            containerFiles = append(containerFiles, file)
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

func ParseFileKeyValue(content string, separator string) map[string]string {
    lines := parseFileLines(content)
    keyValueMap := make(map[string]string)
    for _, line := range lines {
        key, value, found := strings.Cut(line, separator)
        if found {
            keyValueMap[strings.TrimSpace(key)] = strings.TrimSpace(value)
        }
    }
    return keyValueMap
}
