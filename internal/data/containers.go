package data

import (
	"fmt"
	"regexp"
	"slices"
)

// Remember, author/name must be the "repository" segment of an OCI image name
type ImageInfo struct {
	Host   string
	Author string
	Name   string
	Tag    string
	Digest string
}

func ValidateImagePattern(re *regexp.Regexp) error {
	if slices.Contains(re.SubexpNames(), "name") {
			return nil
		}
	return fmt.Errorf("image pattern must define a named group 'name'")
}

type PinLevel int

const (
	// PinFreeze skips the upstream check entirely.
	// Deliberately outside the order below: frozen containers should never compare
	PinFreeze PinLevel = -1
	// PinChannel is the default: only tags matching the current channel (Extra)
	PinChannel PinLevel = 0
	// PinMajor additionally only considers tags sharing the current Major version.
	PinMajor PinLevel = 1
	// PinMinor additionally only considers tags sharing the current Major.Minor version.
	PinMinor PinLevel = 2
)

func (p PinLevel) String() string {
	switch p {
	case PinFreeze:
		return "freeze"
	case PinChannel:
		return "channel"
	case PinMajor:
		return "major"
	case PinMinor:
		return "minor"
	default:
		return "invalid"
	}
}

// Parses the PinLevel string into its corresponding enum value.
func ParsePinLevel(s string) (PinLevel, error) {
	switch s {
	// Empty string represents an omitted Pin key, defaulting to PinChannel.
	case "", "channel":
		return PinChannel, nil
	case "major":
		return PinMajor, nil
	case "minor":
		return PinMinor, nil
	case "freeze":
		return PinFreeze, nil
	default:
		return 0, fmt.Errorf("invalid pin level %q", s)
	}
}

type ContainerDefinition struct {
	Name       string
	Image      ImageInfo
	Version    SemanticVersion
	File       FileEntry
	TagPattern *regexp.Regexp
	Pin        PinLevel
}

type Update struct {
	Definition    ContainerDefinition
	LatestVersion SemanticVersion
	Upgradeable   bool
	// Pinned is true when no update is available within the version pin,
	// but a newer version exists outside of it (e.g. a new major release).
	Pinned bool
}

type Stage string

const (
	ParseStage Stage = "parse"
	CheckStage Stage = "check"
)

type ContainerError struct {
	File  FileEntry
	Name  string
	Stage Stage
	Err   error
}

func (e ContainerError) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("%s (%s): %v", e.Name, e.File.Path, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.File.Path, e.Err)
}

func (e ContainerError) Unwrap() error {
	return e.Err
}

type UpdateReport struct {
	Outdated []Update
	Updated  []Update
	Frozen   []ContainerDefinition
	Errors   []ContainerError
}
