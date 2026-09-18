package data

import "git.ierislabs.dev/ieris19/kluzo/internal/matcher"

type ContainerDefinition struct {
	Name       string
	Image      ImageInfo
	Version    SemanticVersion
	File       FileEntry
	// nil falls back to the default matcher
	TagPattern *matcher.NamedMatcher
	Pin        PinLevel
}
