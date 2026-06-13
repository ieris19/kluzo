package data

import (
    "fmt"
    "regexp"
    "strconv"
)

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

var SemverRegexp = regexp.MustCompile(`^v?(?P<major>\d+)\.(?P<minor>\d+)(?:\.(?P<patch>\d+))?(?:-(?P<extra>.+))?$`)

func ParseSemanticVersion(tag string, re *regexp.Regexp) (SemanticVersion, error) {
    if re == nil {
        re = SemverRegexp
    }
    match := re.FindStringSubmatch(tag)
    if match == nil {
        return SemanticVersion{}, fmt.Errorf("tag %s is not a valid semantic version", tag)
    }
    index := make(map[string]int)
    for i, name := range re.SubexpNames() {
        if name != "" {
            index[name] = i
        }
    }
    get := func(name string) string {
        if i, ok := index[name]; ok {
            return match[i]
        }
        return ""
    }
    major, _ := strconv.Atoi(get("major"))
    minor, _ := strconv.Atoi(get("minor"))
    patch, _ := strconv.Atoi(get("patch"))
    return SemanticVersion{
        Major: major,
        Minor: minor,
        Patch: patch,
        Extra: get("extra"),
    }, nil
}

func ValidateTagPattern(re *regexp.Regexp) error {
    hasMajor, hasMinor := false, false
    for _, name := range re.SubexpNames() {
        switch name {
        case "major":
            hasMajor = true
        case "minor":
            hasMinor = true
        }
    }
    if !hasMajor || !hasMinor {
        return fmt.Errorf("tag pattern must define named groups 'major' and 'minor'")
    }
    return nil
}
