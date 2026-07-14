package parser

import (
    "errors"

    "git.ierislabs.dev/update-link/data"
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

func ParseContainerFiles(file []data.FileEntry) ([]data.ContainerDefinition, []data.ContainerError) {
    var containers []data.ContainerDefinition
    var errs []data.ContainerError
    for _, f := range file {
        container, err := ParseContainerFile(f)
        if err != nil {
            errs = append(errs, data.ContainerError{
                File:  f,
                Stage: data.ParseStage,
                Err:   err,
            })
            continue
        }
        containers = append(containers, container)
    }
    return containers, errs
}

func ParseContainerFile(file data.FileEntry) (data.ContainerDefinition, error) {
    switch file.Extension {
    case ".container":
        return parseQuadletFile(file)
    default:
        return data.ContainerDefinition{}, errors.New("unsupported file extension")
    }
}
