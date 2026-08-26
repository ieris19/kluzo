package parser

import "strings"

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
