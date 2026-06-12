package data

import (
    _ "embed"
    "fmt"
    "regexp"
    "strconv"
    "strings"
)

type ImageInfo struct {
    Host   string
    Author string
    Name   string
    Tag    string
    Digest string
}
type ContainerDefinition struct {
    Name     string
    Image    ImageInfo
    Version  SemanticVersion
    Upstream ContainerUpstream
    File     FileEntry
}

type ContainerUpstream string

const (
    UpstreamGitHub       ContainerUpstream = "github"
    UpstreamDockerHub    ContainerUpstream = "docker"
    UpstreamDistribution ContainerUpstream = "distribution"
)

func CheckContainerUpstream(image ImageInfo) ContainerUpstream {
    switch strings.ToLower(image.Host) {
    case "docker.io":
        return UpstreamDockerHub
    case "ghcr.io", "github.com":
        return UpstreamGitHub
    default:
        return UpstreamDistribution
    }
}

type SemanticVersion struct {
    Major int
    Minor int
    Patch int
    Extra string
}

func (sv SemanticVersion) String() string {
    extra := ""
    if sv.Extra != "" {
        extra = "-" + sv.Extra
    }
    return fmt.Sprintf("%d.%d.%d%s", sv.Major, sv.Minor, sv.Patch, extra)
}

func (sv SemanticVersion) LessThan(other SemanticVersion) bool {
    if sv.Major != other.Major {
        return sv.Major < other.Major
    }
    if sv.Minor != other.Minor {
        return sv.Minor < other.Minor
    }
    return sv.Patch < other.Patch
}

func (sv SemanticVersion) GreaterThan(other SemanticVersion) bool {
    if sv.Major != other.Major {
        return sv.Major > other.Major
    }
    if sv.Minor != other.Minor {
        return sv.Minor > other.Minor
    }
    return sv.Patch > other.Patch
}

func (sv SemanticVersion) Equals(other SemanticVersion) bool {
    return sv.Major == other.Major &&
        sv.Minor == other.Minor &&
        sv.Patch == other.Patch &&
        sv.Extra == other.Extra
}

var SemverRegexp = regexp.MustCompile(`^v?(\d+)\.(\d+)(?:\.(\d+))?(?:-(.+))?$`)

const (
    semverMajorGroup = 1
    semverMinorGroup = 2
    semverPatchGroup = 3
    semverExtraGroup = 4
)

func ParseSemanticVersion(tag string) (SemanticVersion, error) {
    matches := SemverRegexp.FindStringSubmatch(tag)
    if matches == nil {
        return SemanticVersion{}, fmt.Errorf("tag %s is not a valid semantic version", tag)
    }

    major, _ := strconv.Atoi(matches[semverMajorGroup])
    minor, _ := strconv.Atoi(matches[semverMinorGroup])
    patch, _ := strconv.Atoi(matches[semverPatchGroup])

    return SemanticVersion{
        Major: major,
        Minor: minor,
        Patch: patch,
        Extra: matches[semverExtraGroup],
    }, nil
}
