package semver

import (
	"errors"
	"fmt"
	"strconv"

	"git.ierislabs.dev/ieris19/kluzo/internal/matcher"
)

// ErrNotSemver marks a tag that carries no version
var ErrNotSemver = errors.New("not a valid semantic version")

func Parse(tag string, pattern *TagPattern) (Version, error) {
	if pattern == nil {
		pattern = defaultTagPattern
	}
	match, ok := pattern.matcher.Match(tag)
	if !ok {
		// No digits at all: no pattern could ever turn this into a version (e.g. "latest").
		// Mark it so callers can treat as Untrackable rather than a real error.
		if !hasDigit(tag) {
			return Version{}, fmt.Errorf("tag '%s' is not a valid semantic version: %w", tag, ErrNotSemver)
		}
		// A tag with digits might still be a version in an unexpected shape, error to get user's attention
		return Version{}, fmt.Errorf("tag '%s' is not a valid semantic version", tag)
	}
	segments, err := pattern.parseSegments(tag, match)
	if err != nil {
		return Version{}, err
	}
	return Version{
		Segments: segments,
		Extra:    match.At(pattern.extra),
		Raw:      tag,
	}, nil
}

// parseSegments reads the resolved segments. Trailing segments may be absent,
// but a gap leaves the tag ambiguous and is rejected.
func (p *TagPattern) parseSegments(tag string, match matcher.NamedMatch) ([]int, error) {
	segments := make([]int, 0, len(p.groups))
	// gap holds the first absent segment, or 0 while none has been seen
	gap := 0
	for n, group := range p.groups {
		seg := match.At(group)
		if seg == "" {
			if gap == 0 {
				gap = n + 1
			}
			continue
		}
		if gap != 0 {
			return nil, fmt.Errorf("tag '%s' leaves segment %d empty but defines segment %d", tag, gap, n+1)
		}
		value, err := strconv.Atoi(seg)
		if err != nil {
			return nil, fmt.Errorf("tag '%s' has an invalid segment %d: %v", tag, n+1, err)
		}
		segments = append(segments, value)
	}
	if len(segments) == 0 {
		return nil, fmt.Errorf("tag '%s' matched the pattern but yielded no version segments", tag)
	}
	return segments, nil
}

func hasDigit(s string) bool {
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			return true
		}
	}
	return false
}
