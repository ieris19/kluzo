package data

import (
	"fmt"
	"regexp"
	"strconv"

	"git.ierislabs.dev/ieris19/kluzo/internal/matcher"
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

// Compare returns -1(<), 0(=) or 1(>) comparing against another SemanticVersion
// It will return error if the semantic version's "extra" differs
func (sv SemanticVersion) Compare(other SemanticVersion) (int, error) {
	if sv.Extra != other.Extra {
		return 0, fmt.Errorf("cannot compare versions with different labels: %q vs %q", sv.Extra, other.Extra)
	}
	if sv.Major != other.Major {
		return cmpInt(sv.Major, other.Major), nil
	}
	if sv.Minor != other.Minor {
		return cmpInt(sv.Minor, other.Minor), nil
	}
	return cmpInt(sv.Patch, other.Patch), nil
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func (sv SemanticVersion) Equals(other SemanticVersion) bool {
	return sv.Major == other.Major &&
		sv.Minor == other.Minor &&
		sv.Patch == other.Patch &&
		sv.Extra == other.Extra
}

var semverMatcher = matcher.NewNamedMatcher(regexp.MustCompile(`^v?(?P<major>\d+)\.(?P<minor>\d+)(?:\.(?P<patch>\d+))?(?:-(?P<extra>.+))?$`))

func ParseSemanticVersion(tag string, re *regexp.Regexp) (SemanticVersion, error) {
	regexMatcher := semverMatcher
	if re != nil {
		regexMatcher = matcher.NewNamedMatcher(re)
	}
	match, ok := regexMatcher.Match(tag)
	if !ok {
		return SemanticVersion{}, fmt.Errorf("tag '%s' is not a valid semantic version", tag)
	}
	major, err := strconv.Atoi(match.Get("major"))
	if err != nil {
		return SemanticVersion{}, fmt.Errorf("tag '%s' has an invalid major version: %v", tag, err)
	}
	minor, err := strconv.Atoi(match.Get("minor"))
	if err != nil {
		return SemanticVersion{}, fmt.Errorf("tag '%s' has an invalid minor version: %v", tag, err)
	}
	// patch is optional: empty capture (no match) means 0, anything else must be numeric
	patch := 0
	if raw := match.Get("patch"); raw != "" {
		patch, err = strconv.Atoi(raw)
		if err != nil {
			return SemanticVersion{}, fmt.Errorf("tag '%s' has an invalid patch version: %v", tag, err)
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
