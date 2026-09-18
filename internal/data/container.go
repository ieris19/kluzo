package data

import "git.ierislabs.dev/ieris19/kluzo/internal/semver"

type ContainerDefinition struct {
	Name    string
	Image   ImageInfo
	Version semver.Version
	File    FileEntry
	// nil falls back to the default pattern
	TagPattern *semver.TagPattern
	Pin        PinLevel
}
