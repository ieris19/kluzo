package parser

import (
	"fmt"
	"regexp"

	"git.ierislabs.dev/ieris19/kluzo/internal/data"
	"git.ierislabs.dev/ieris19/kluzo/internal/files"
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
	update := sections["X-Kluzo"]

	containerName := container["ContainerName"]

	var imagePattern *regexp.Regexp
	if pattern := update["ImagePattern"]; pattern != "" {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return data.ContainerDefinition{}, fmt.Errorf("invalid image pattern for %s: %w", containerName, err)
		}
		if err := data.ValidateImagePattern(re); err != nil {
			return data.ContainerDefinition{}, fmt.Errorf("invalid image pattern for %s: %w", containerName, err)
		}
		imagePattern = re
	}

	imageInfo, err := parseImageInfo(container["Image"], imagePattern)
	if err != nil {
		fmt.Printf("Error parsing image %s: %v\n", file.Path, err)
		return data.ContainerDefinition{}, err
	}

	if imageInfo.Digest != "" || imageInfo.Tag == "" {
		return data.ContainerDefinition{}, fmt.Errorf("no semver tag found for %s", containerName)
	}

	var tagPattern *regexp.Regexp
	if pattern := update["TagPattern"]; pattern != "" {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return data.ContainerDefinition{}, fmt.Errorf("invalid tag pattern for %s: %w", containerName, err)
		}
		if err := data.ValidateTagPattern(re); err != nil {
			return data.ContainerDefinition{}, fmt.Errorf("invalid tag pattern for %s: %w", containerName, err)
		}
		tagPattern = re
	}

	semver, err := data.ParseSemanticVersion(imageInfo.Tag, tagPattern)
	if err != nil {
		return data.ContainerDefinition{}, err
	}

	pin := data.PinLevel(update["VersionPin"])
	if !pin.Valid() {
		return data.ContainerDefinition{}, fmt.Errorf("invalid pin level %q for %s", pin, containerName)
	}
	if pin == "" {
		pin = data.PinChannel
	}

	return data.ContainerDefinition{
		Name:       containerName,
		Image:      imageInfo,
		Version:    semver,
		File:       file,
		TagPattern: tagPattern,
		Pin:        pin,
	}, nil
}
