package matcher

import "regexp"

type NamedMatcher struct {
	re    *regexp.Regexp
	index map[string]int
}

// NewNamedMatcher builds the name index once; reuse the result across inputs
// rather than rebuilding it per match.
func NewNamedMatcher(re *regexp.Regexp) *NamedMatcher {
	index := make(map[string]int)
	for i, name := range re.SubexpNames() {
		if name != "" {
			index[name] = i
		}
	}
	return &NamedMatcher{re: re, index: index}
}

// Index reports which submatch the given group name occupies. Callers that
// resolve a name once can then address it by position on every match.
func (nm *NamedMatcher) Index(name string) (int, bool) {
	i, ok := nm.index[name]
	return i, ok
}

func (nm *NamedMatcher) Match(s string) (NamedMatch, bool) {
	sub := nm.re.FindStringSubmatch(s)
	if sub == nil {
		return NamedMatch{}, false
	}
	return NamedMatch{sub: sub, index: nm.index}, true
}

type NamedMatch struct {
	sub   []string
	index map[string]int
}

func (nm NamedMatch) Get(name string) string {
	if i, ok := nm.index[name]; ok {
		return nm.sub[i]
	}
	return ""
}

// At returns the submatch at the given index. A negative index is the
// conventional "group not in this pattern" and yields "". Any other
// out-of-range index is a caller bug, and is left to panic as such.
func (nm NamedMatch) At(i int) string {
	if i < 0 {
		return ""
	}
	return nm.sub[i]
}
