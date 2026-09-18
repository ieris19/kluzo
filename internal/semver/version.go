package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version holds an arbitrary number of numeric segments.
// Absent trailing segments are simply missing, normalized as 0 on access.
type Version struct {
	Segments []int
	Extra    string
	Raw      string
}

// Segment returns the segment corresponding to the given 1-based index.
// Past the end of the array, segments return 0 as padding.
func (v Version) Segment(n int) int {
	if n < 1 || n > len(v.Segments) {
		return 0
	}
	return v.Segments[n-1]
}

// Depth is simply the amount of segments present.
func (v Version) Depth() int {
	return len(v.Segments)
}

// String returns the original string if available, or a synthetic string.
// Without segments there is no version to render, only the zero value.
func (v Version) String() string {
	if v.Raw != "" {
		return v.Raw
	}
	// Never as digits: "0.0.0" is a real version some container could be on
	if len(v.Segments) == 0 {
		return "<none>"
	}
	var parts = make([]string, len(v.Segments))
	for i, segment := range v.Segments {
		parts[i] = strconv.Itoa(segment)
	}
	extra := ""
	if v.Extra != "" {
		extra = "-" + v.Extra
	}
	return strings.Join(parts, ".") + extra
}

// Compare returns -1(<), 0(=) or 1(>) comparing against another Version
// It will return error if the version's "extra" differs
func (v Version) Compare(other Version) (int, error) {
	if v.Extra != other.Extra {
		return 0, fmt.Errorf("cannot compare versions with different labels: %q vs %q", v.Extra, other.Extra)
	}
	depth := max(v.Depth(), other.Depth())
	for n := 1; n <= depth; n++ {
		if c := cmpInt(v.Segment(n), other.Segment(n)); c != 0 {
			return c, nil
		}
	}
	return 0, nil
}

// SharesPrefix reports whether both versions agree on their first n segments.
// Segments past either version's depth read as 0, so a deeper n than both
// carry simply compares equal.
func (v Version) SharesPrefix(other Version, n int) bool {
	for i := 1; i <= n; i++ {
		if v.Segment(i) != other.Segment(i) {
			return false
		}
	}
	return true
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

func (v Version) Equals(other Version) bool {
	// A differing Extra is reported as an error rather than an ordering, and
	// versions on different channels are never equal
	c, err := v.Compare(other)
	return err == nil && c == 0
}
