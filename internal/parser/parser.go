package parser

import (
	"errors"

	"git.ierislabs.dev/ieris19/kluzo/internal/data"
)

var SupportedExtensions []string = []string{".container"}

var registryAliases map[string]string

func SetAliases(aliases map[string]string) {
	registryAliases = aliases
}

func resolveHostAlias(host string) string {
	if alias, ok := registryAliases[host]; ok {
		return alias
	}
	return host
}

func ParseContainerFiles(file []data.FileEntry, report *data.UpdateReport) []data.ContainerDefinition {
	var containers []data.ContainerDefinition
	for _, f := range file {
		container, err := ParseContainerFile(f)
		if err != nil {
			cErr := data.ContainerError{
				File:  f,
				Name:  container.Name,
				Stage: data.ParseStage,
				Err:   err,
			}
			if errors.Is(err, data.ErrNotSemver) {
				report.Skipped = append(report.Skipped, cErr)
			} else {
				report.Errors = append(report.Errors, cErr)
			}
			continue
		}
		containers = append(containers, container)
	}
	return containers
}

func ParseContainerFile(file data.FileEntry) (data.ContainerDefinition, error) {
	switch file.Extension {
	case ".container":
		return parseQuadletFile(file)
	default:
		return data.ContainerDefinition{}, errors.New("unsupported file extension")
	}
}
