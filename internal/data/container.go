package data

import "regexp"

type ContainerDefinition struct {
	Name       string
	Image      ImageInfo
	Version    SemanticVersion
	File       FileEntry
	TagPattern *regexp.Regexp
	Pin        PinLevel
}
