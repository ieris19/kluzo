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
	Raw   string
}

func (sv SemanticVersion) String() string {
	if sv.Raw != "" {
		return sv.Raw
	}
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

var semverMatcher = NewNamedMatcher(regexp.MustCompile(`^v?(?P<major>\d+)\.(?P<minor>\d+)(?:\.(?P<patch>\d+))?(?:-(?P<extra>.+))?$`))

func ParseSemanticVersion(tag string, re *regexp.Regexp) (SemanticVersion, error) {
	matcher := semverMatcher
	if re != nil {
		matcher = NewNamedMatcher(re)
	}
	match, ok := matcher.Match(tag)
	if !ok {
		return SemanticVersion{}, fmt.Errorf("tag '%s' is not a valid semantic version", tag)
	}
	major, err := strconv.Atoi(match.Get("major"))
	if err != nil {
		return SemanticVersion{}, fmt.Errorf("tag '%s' has a non-numeric major version: %v", tag, err)
	}
	minor, err := strconv.Atoi(match.Get("minor"))
	if err != nil {
		return SemanticVersion{}, fmt.Errorf("tag '%s' has a non-numeric minor version: %v", tag, err)
	}
	// patch is optional: empty capture (no match) means 0, anything else must be numeric
	patch := 0
	if raw := match.Get("patch"); raw != "" {
		patch, err = strconv.Atoi(raw)
		if err != nil {
			return SemanticVersion{}, fmt.Errorf("tag '%s' has a non-numeric patch version: %v", tag, err)
		}
	}
	return SemanticVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
		Extra: match.Get("extra"),
		Raw:   tag,
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
