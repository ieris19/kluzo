package parser

import (
    "fmt"

    "git.ierislabs.dev/update-link/data"
    "git.ierislabs.dev/update-link/utils"
)

func parseQuadletFile(file data.FileEntry) (data.ContainerDefinition, error) {
    // Read the file content
    content, err := data.ReadFileContent(file.Path)
    if err != nil {
        return data.ContainerDefinition{}, err
    }

    // Parse the quadlet content to extract container information
    quadletInfo := utils.ParseFileKeyValue(content, "=")

    imageInfo, err := parseImageInfo(quadletInfo["Image"])
    if err != nil {
        fmt.Printf("Error parsing image %s: %v", file.Path, err)
        return data.ContainerDefinition{}, err
    }

    upstream := data.CheckContainerUpstream(imageInfo)

    var semver data.SemanticVersion

    if imageInfo.Digest == "" && imageInfo.Tag != "" {
        semver, err = data.ParseSemanticVersion(imageInfo.Tag)
    }

    return data.ContainerDefinition{
        Name:     quadletInfo["ContainerName"],
        Image:    imageInfo,
        Upstream: upstream,
        Version:  utils.TernaryIf(err == nil, semver, data.SemanticVersion{}),
        File:     file,
    }, nil
}
