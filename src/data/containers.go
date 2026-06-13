package data

type ImageInfo struct {
    Host   string
    Author string
    Name   string
    Tag    string
    Digest string
}
type ContainerDefinition struct {
    Name    string
    Image   ImageInfo
    Version SemanticVersion
    File    FileEntry
}
