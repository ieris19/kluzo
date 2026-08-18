package parser

import (
    "errors"
    "regexp"
    "strings"

    "git.ierislabs.dev/ieris19/update-link/data"
)

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

// looksLikeHost mirrors the convention Docker/Podman use to decide whether a
// single path segment in front of the image name is a registry host: it
// only counts as a host if it contains a '.' or ':', or is "localhost".
// Anything else is a repository path segment on the default registry.
func looksLikeHost(segment string) bool {
    return segment == "localhost" || strings.ContainsAny(segment, ".:")
}

var imageMatcher = data.NewNamedMatcher(regexp.MustCompile(`^(?:(?P<host>[^/\s]*)/)?(?:(?P<user>[^/\s]*)/)?(?P<name>[^\s/:@]*)(?::(?P<tag>[\w.-]*))?(?:@(?P<digest>sha\d{3}:[a-z0-9]*))?$`))

func parseImageInfo(image string, customPattern *regexp.Regexp) (data.ImageInfo, error) {
    matcher := imageMatcher

    if customPattern != nil {
        matcher = data.NewNamedMatcher(customPattern)
    }

    match, ok := matcher.Match(image)
    if !ok {
        return data.ImageInfo{}, errors.New("image format did not match expected pattern")
    }

    host := match.Get("host")
    user := match.Get("user")

    // Remove ambiguity between user/image from host/image
    if host != "" && user == "" && !looksLikeHost(host) {
        user = host
        host = ""
    }

    // Special corrections for compatibility
    aliasedHost := resolveHostAlias(host)
    imageAuthor := determineAuthor(user, aliasedHost)

    return data.ImageInfo{
        Host:   aliasedHost,
        Author: imageAuthor,
        Name:   match.Get("name"),
        Tag:    match.Get("tag"),
        Digest: match.Get("digest"),
    }, nil
}
