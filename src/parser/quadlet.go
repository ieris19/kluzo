package parser

import (
    "fmt"
    "regexp"

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
    sections := files.ParseFileKeyValue(content, "=")
    container := sections["Container"]

    imageInfo, err := parseImageInfo(container["Image"])
    if err != nil {
        fmt.Printf("Error parsing image %s: %v\n", file.Path, err)
        return data.ContainerDefinition{}, err
    }

    if imageInfo.Digest != "" || imageInfo.Tag == "" {
        return data.ContainerDefinition{}, fmt.Errorf("no semver tag found for %s", container["ContainerName"])
    }

    var tagPattern *regexp.Regexp
    if pattern := sections["X-UpdateLink"]["TagPattern"]; pattern != "" {
        re, err := regexp.Compile(pattern)
        if err != nil {
            return data.ContainerDefinition{}, fmt.Errorf("invalid tag pattern for %s: %w", container["ContainerName"], err)
        }
        if err := data.ValidateTagPattern(re); err != nil {
            return data.ContainerDefinition{}, fmt.Errorf("invalid tag pattern for %s: %w", container["ContainerName"], err)
        }
        tagPattern = re
    }

    semver, err := data.ParseSemanticVersion(imageInfo.Tag, tagPattern)
    if err != nil {
        return data.ContainerDefinition{}, err
    }

    return data.ContainerDefinition{
        Name:       container["ContainerName"],
        Image:      imageInfo,
        Version:    semver,
        File:       file,
        TagPattern: tagPattern,
    }, nil
}
