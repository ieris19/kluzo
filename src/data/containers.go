package data

import (
    "fmt"
    "regexp"
)

type ImageInfo struct {
    Host   string
    Author string
    Name   string
    Tag    string
    Digest string
}

type ContainerDefinition struct {
    Name       string
    Image      ImageInfo
    Version    SemanticVersion
    File       FileEntry
    TagPattern *regexp.Regexp
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
    Errors   []ContainerError
}
