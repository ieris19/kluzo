package parser

import (
    "errors"
    "fmt"
    "regexp"

    "git.ierislabs.dev/update-link/data"
)

var supportedExtensions []string = []string{".container"}

func IsSupportedExtension(extension string) bool {
    for _, ext := range supportedExtensions {
        if ext == extension {
            return true
        }
    }
    return false
}

func ParseContainerFiles(file []data.FileEntry) []data.ContainerDefinition {
    var containers []data.ContainerDefinition
    for _, f := range file {
        container, err := ParseContainerFile(f)
        if err != nil {
            fmt.Printf("Error parsing file %s: %v\n", f.Path, err)
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

var imageRegexp = regexp.MustCompile(`^(([^/\s]*)/)?(([^/\s]*)/)?([^\s/:@]*)((:([\w.-]*))|(@(sha\d{3}:[a-z0-9]*)))?$`)

const (
    hostGroup   = 2
    userGroup   = 4
    nameGroup   = 5
    tagGroup    = 8
    digestGroup = 10
)

func parseImageInfo(image string) (data.ImageInfo, error) {
    matches := imageRegexp.FindStringSubmatch(image)
    if matches == nil {
        return data.ImageInfo{}, errors.New("image format did not match expected pattern")
    }

    return data.ImageInfo{
        Host:   matches[hostGroup],
        Author: matches[userGroup],
        Name:   matches[nameGroup],
        Tag:    matches[tagGroup],
        Digest: matches[digestGroup],
    }, nil
}

func mustParseImageInfo(image string) data.ImageInfo {
    info, err := parseImageInfo(image)
    if err != nil {
        panic(err)
    }
    return info
}
