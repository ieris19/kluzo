package parser

import (
    "fmt"

    "git.ierislabs.dev/update-link/data"
    "git.ierislabs.dev/update-link/files"
)

func parseQuadletFile(file data.FileEntry) (data.ContainerDefinition, error) {
    // Read the file content
    content, err := data.ReadFileContent(file.Path)
    if err != nil {
        return data.ContainerDefinition{}, err
    }

    // Parse the quadlet content to extract container information
    quadletInfo := files.ParseFileKeyValue(content, "=")

    imageInfo, err := parseImageInfo(quadletInfo["Image"])
    if err != nil {
        fmt.Printf("Error parsing image %s: %v", file.Path, err)
        return data.ContainerDefinition{}, err
    }

    if imageInfo.Digest != "" || imageInfo.Tag == "" {
        return data.ContainerDefinition{}, fmt.Errorf("no semver tag found for %s", quadletInfo["ContainerName"])
    }
    semver, err := data.ParseSemanticVersion(imageInfo.Tag)
    if err != nil {
        return data.ContainerDefinition{}, err
    }

    return data.ContainerDefinition{
        Name:    quadletInfo["ContainerName"],
        Image:   imageInfo,
        Version: semver,
        File:    file,
    }, nil
}
