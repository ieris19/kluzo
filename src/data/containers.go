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
    UpstreamUnknown      ContainerUpstream = "unknown"
)

func CheckContainerUpstream(image ImageInfo) ContainerUpstream {
    switch strings.ToLower(image.Host) {
    case "docker.io":
        return UpstreamDockerHub
    case "ghcr.io", "github.com":
        return UpstreamGitHub
    default:
        return checkDistributionDomains(image)
    }
}

func checkDistributionDomains(image ImageInfo) ContainerUpstream {
    // Additional logic to determine if we can identify the upstream
    host := strings.ToLower(image.Host)
    for _, knownHost := range DistributionDomains {
        if strings.Contains(host, knownHost) {
            return UpstreamDistribution
        }
    }
    return UpstreamUnknown
}

type SemanticVersion struct {
    Major int
    Minor int
    Patch int
    Extra string
}

func (self SemanticVersion) String() string {
    extra := ""
    if self.Extra != "" {
        extra = "-" + self.Extra
    }
    return fmt.Sprintf("%d.%d.%d%s", self.Major, self.Minor, self.Patch, extra)
}

func (self SemanticVersion) LessThan(other SemanticVersion) bool {
    if self.Major != other.Major {
        return self.Major < other.Major
    }
    if self.Minor != other.Minor {
        return self.Minor < other.Minor
    }
    return self.Patch < other.Patch
}

func (self SemanticVersion) GreaterThan(other SemanticVersion) bool {
    if self.Major != other.Major {
        return self.Major > other.Major
    }
    if self.Minor != other.Minor {
        return self.Minor > other.Minor
    }
    return self.Patch > other.Patch
}

func (self SemanticVersion) Equals(other SemanticVersion) bool {
    return self.Major == other.Major &&
        self.Minor == other.Minor &&
        self.Patch == other.Patch &&
        self.Extra == other.Extra
}

var SemverRegexp = regexp.MustCompile(`^v?(\d+)(.(\d+))?(.(\d+))?(-(.+))?$`)

const (
    semverMajorGroup = 1
    semverMinorGroup = 3
    semverPatchGroup = 5
    semverExtraGroup = 7
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
