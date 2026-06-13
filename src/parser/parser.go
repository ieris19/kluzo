package parser

import (
    "errors"
    "fmt"
    "regexp"
    "strings"

    "git.ierislabs.dev/update-link/data"
)

var SupportedExtensions []string = []string{".container"}

var registryAliases map[string]string

func SetAliases(aliases map[string]string) {
    registryAliases = aliases
}

func resolveHostAlias(host string) string {
    if alias, ok := registryAliases[host]; ok {
        return alias
    }
    return host
}

func determineAuthor(user string, host string) string {
    if !strings.Contains(host, "docker.io") {
        return user
    }
    author := user
    if author == "" {
        author = "library"
    }
    return author
}

func ParseContainerFiles(file []data.FileEntry) []data.ContainerDefinition {
    var containers []data.ContainerDefinition
    for _, f := range file {
        container, err := ParseContainerFile(f)
        if err != nil {
            fmt.Printf("Error parsing file %s: %v\n", f.Path, err)
            continue
        }
        containers = append(containers, container)
    }
    return containers
}

func ParseContainerFile(file data.FileEntry) (data.ContainerDefinition, error) {
    switch file.Extension {
    case ".container":
        return parseQuadletFile(file)
    default:
        return data.ContainerDefinition{}, errors.New("unsupported file extension")
    }
}

var imageMatcher = data.NewNamedMatcher(regexp.MustCompile(`^(?:(?P<host>[^/\s]*)/)?(?:(?P<user>[^/\s]*)/)?(?P<name>[^\s/:@]*)(?::(?P<tag>[\w.-]*)|@(?P<digest>sha\d{3}:[a-z0-9]*))?$`))

func parseImageInfo(image string) (data.ImageInfo, error) {
    match, ok := imageMatcher.Match(image)
    if !ok {
        return data.ImageInfo{}, errors.New("image format did not match expected pattern")
    }

    // Special corrections for compatibility
    aliasedHost := resolveHostAlias(match.Get("host"))
    imageAuthor := determineAuthor(match.Get("user"), aliasedHost)

    return data.ImageInfo{
        Host:   aliasedHost,
        Author: imageAuthor,
        Name:   match.Get("name"),
        Tag:    match.Get("tag"),
        Digest: match.Get("digest"),
    }, nil
}
