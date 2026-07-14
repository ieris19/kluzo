package data

import "regexp"

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

type ContainerError struct {
    Message    string
    Definition ContainerDefinition
}

type UpdateReport struct {
    Outdated []Update
    Updated  []Update
    Errors   []ContainerError
}
