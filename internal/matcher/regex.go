package matcher

import "regexp"

type NamedMatcher struct {
	re    *regexp.Regexp
	index map[string]int
}

func NewNamedMatcher(re *regexp.Regexp) NamedMatcher {
	index := make(map[string]int)
	for i, name := range re.SubexpNames() {
		if name != "" {
			index[name] = i
		}
	}
	return NamedMatcher{re: re, index: index}
}

func (nm NamedMatcher) Match(s string) (NamedMatch, bool) {
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
