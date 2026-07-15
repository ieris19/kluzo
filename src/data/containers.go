package data

import (
    "fmt"
    "regexp"
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
    for _, name := range re.SubexpNames() {
        if name == "name" {
            return nil
        }
    }
    return fmt.Errorf("image pattern must define a named group 'name'")
}

type PinLevel string

const (
    // PinChannel is the default: only tags matching the current channel (Extra)
    PinChannel PinLevel = "channel"
    // PinMajor additionally only considers tags sharing the current Major version.
    PinMajor PinLevel = "major"
    // PinMinor additionally only considers tags sharing the current Major.Minor version.
    PinMinor PinLevel = "minor"
    // PinFreeze skips the upstream check entirely.
    PinFreeze PinLevel = "freeze"
)

// Empty string represents an omitted Pin key, defaulting to PinChannel.
func (p PinLevel) Valid() bool {
    switch p {
    case "", PinChannel, PinMajor, PinMinor, PinFreeze:
        return true
    default:
        return false
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
