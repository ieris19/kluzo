package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"git.ierislabs.dev/ieris19/kluzo/internal/matcher"
)

// segmentAliases holds the original name for each of the first segments.
// This enables common patterns to be semantic, and most importantly,
// allows parsing patterns from before the arbitrary 'level+number' format.
var segmentAliases = []string{"major", "minor", "patch"}

// segmentName is the capture group name every segment answers to.
func segmentName(n int) string {
	return "level" + strconv.Itoa(n)
}

// segmentAlias is the conventional name for segment n, or "" past the table.
func segmentAlias(n int) string {
	if n < 1 || n > len(segmentAliases) {
		return ""
	}
	return segmentAliases[n-1]
}

// segmentIndex is the inverse of the two above: the segment a group name
// addresses, or 0 if the name addresses no segment at all.
func segmentIndex(name string) int {
	for i, alias := range segmentAliases {
		if name == alias {
			return i + 1
		}
	}
	rest, ok := strings.CutPrefix(name, "level")
	if !ok {
		return 0
	}
	n, err := strconv.Atoi(rest)
	if err != nil || n < 1 {
		return 0
	}
	return n
}

// TagPattern is a compiled tag regex whose version segments have already been
// resolved to submatch positions, so matching a tag costs no name lookups.
type TagPattern struct {
	matcher *matcher.NamedMatcher
	// groups[n-1] is the capture group holding segment n
	groups []int
	// extra is the capture group holding the label, or -1 when the pattern has none
	extra int
}

// NewTagPattern resolves a tag pattern's groups, rejecting the shapes the
// parser cannot make sense of. Validation and resolution are a single walk on
// purpose: a pattern that parses is a pattern that was checked.
func NewTagPattern(re *regexp.Regexp) (*TagPattern, error) {
	nm := matcher.NewNamedMatcher(re)
	var groups []int
	// Walk segments from the bottom; the first undefined one ends the pattern
	for n := 1; ; n++ {
		idx, found := -1, ""
		if alias := segmentAlias(n); alias != "" {
			if i, ok := nm.Index(alias); ok {
				idx, found = i, alias
			}
		}
		name := segmentName(n)
		if i, ok := nm.Index(name); ok {
			if found != "" {
				return nil, fmt.Errorf("tag pattern defines both %q and %q for the same version segment", found, name)
			}
			idx, found = i, name
		}
		if found == "" {
			break
		}
		groups = append(groups, idx)
	}
	if len(groups) < 2 {
		return nil, fmt.Errorf("tag pattern must define at least two version segments ('major'/'level1' and 'minor'/'level2')")
	}
	// Segments are collected until one is missing, warn on gaps
	for _, name := range re.SubexpNames() {
		if n := segmentIndex(name); n > len(groups) {
			return nil, fmt.Errorf("tag pattern defines %q but leaves level %d undefined", name, len(groups)+1)
		}
	}
	extra := -1
	if i, ok := nm.Index("extra"); ok {
		extra = i
	}
	return &TagPattern{matcher: nm, groups: groups, extra: extra}, nil
}

// mustTagPattern builds a pattern known good at compile time.
func mustTagPattern(expr string) *TagPattern {
	pattern, err := NewTagPattern(regexp.MustCompile(expr))
	if err != nil {
		panic("invalid built-in tag pattern: " + err.Error())
	}
	return pattern
}

// defaultTagPattern is the shape assumed when a container declares none.
var defaultTagPattern = mustTagPattern(`^v?(?P<major>\d+)\.(?P<minor>\d+)(?:\.(?P<patch>\d+))?(?:-(?P<extra>.+))?$`)
