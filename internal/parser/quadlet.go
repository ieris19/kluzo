package parser

import (
	"fmt"
	"regexp"
	"strings"

	"git.ierislabs.dev/ieris19/kluzo/internal/data"
	"git.ierislabs.dev/ieris19/kluzo/internal/files"
)

func parseQuadletFile(file data.FileEntry) (def data.ContainerDefinition, errR error) {
	// Read the file content
	content, err := data.ReadFileContent(file.Path)
	if err != nil {
		return data.ContainerDefinition{}, fmt.Errorf("could not read %s: %v", file.Path, err)
	}

	// Parse the quadlet content to extract container information
	sections := files.ParseFileKeyValue(content, "=")
	container := sections["Container"]
	update := sections["X-Kluzo"]

	containerName := container["ContainerName"]
	if containerName == "" {
		// ContainerName= is optional. Podman falls back to systemd-<unitname>
		unitName := strings.TrimSuffix(file.Name, file.Extension)
		containerName = "systemd-" + unitName
	}
	// Best-effort to provide at least the name even if the parser fails
	defer func() {
		if errR != nil {
			def.Name = containerName
		}
	}()

	var imagePattern *regexp.Regexp
	if pattern := update["ImagePattern"]; pattern != "" {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return data.ContainerDefinition{}, fmt.Errorf("invalid image pattern for %s: %v", containerName, err)
		}
		if err := data.ValidateImagePattern(re); err != nil {
			return data.ContainerDefinition{}, fmt.Errorf("invalid image pattern for %s: %v", containerName, err)
		}
		imagePattern = re
	}

	imageInfo, err := parseImageInfo(container["Image"], imagePattern)
	if err != nil {
		return data.ContainerDefinition{}, fmt.Errorf("invalid image for %s: %v", containerName, err)
	}

	if imageInfo.Digest != "" || imageInfo.Tag == "" {
		return data.ContainerDefinition{}, fmt.Errorf("no semver tag found for %s", containerName)
	}

	var tagPattern *regexp.Regexp
	if pattern := update["TagPattern"]; pattern != "" {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return data.ContainerDefinition{}, fmt.Errorf("invalid tag pattern for %s: %v", containerName, err)
		}
		if err := data.ValidateTagPattern(re); err != nil {
			return data.ContainerDefinition{}, fmt.Errorf("invalid tag pattern for %s: %v", containerName, err)
		}
		tagPattern = re
	}

	semver, err := data.ParseSemanticVersion(imageInfo.Tag, tagPattern)
	if err != nil {
		return data.ContainerDefinition{}, fmt.Errorf("invalid version for %s: %v", containerName, err)
	}

	pin, err := data.ParsePinLevel(update["VersionPin"])
	if err != nil {
		return data.ContainerDefinition{}, fmt.Errorf("%v for %s", err, containerName)
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
