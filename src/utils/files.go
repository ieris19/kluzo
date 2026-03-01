package utils

import "strings"

func ParseFileLines(content string) []string {
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
    lines := ParseFileLines(content)
    keyValueMap := make(map[string]string)
    for _, line := range lines {
        key, value, found := strings.Cut(line, separator)
        if found {
            keyValueMap[strings.TrimSpace(key)] = strings.TrimSpace(value)
        }
    }
    return keyValueMap
}
