package semver

import (
	"regexp"
	"testing"
)

// Validate custom patterns enforce group presence correctly
func TestNewTagPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		// Patch is optional
		{name: "has major and minor", pattern: `(?P<major>\d+)\.(?P<minor>\d+)`, wantErr: false},
		{name: "has major, minor and patch", pattern: `(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)`, wantErr: false},
		// Major is mandatory
		{name: "missing major", pattern: `(?P<minor>\d+)`, wantErr: true},
		// Minor is mandatory
		{name: "missing minor", pattern: `(?P<major>\d+)`, wantErr: true},
		// Major and Minor are mandatory
		{name: "missing both", pattern: `\d+\.\d+`, wantErr: true},
		// The level form is equivalent to the aliases
		{name: "level form", pattern: `(?P<level1>\d+)\.(?P<level2>\d+)`, wantErr: false},
		{name: "level form past the aliases", pattern: `(?P<level1>\d+)\.(?P<level2>\d+)\.(?P<level3>\d+)\.(?P<level4>\d+)`, wantErr: false},
		{name: "aliases and levels mixed", pattern: `(?P<major>\d+)\.(?P<minor>\d+)\.(?P<level3>\d+)`, wantErr: false},
		// One segment cannot answer to two names at once
		{name: "major and level1 both defined", pattern: `(?P<major>\d+)(?P<level1>\d+)\.(?P<minor>\d+)`, wantErr: true},
		{name: "patch and level3 both defined", pattern: `(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)(?P<level3>x)`, wantErr: true},
		// A gap would silently drop every segment above it
		{name: "gap leaves level 3 undefined", pattern: `(?P<major>\d+)\.(?P<minor>\d+)\.(?P<level4>\d+)`, wantErr: true},
		// A label group is optional and addresses no segment
		{name: "extra is not a segment", pattern: `(?P<major>\d+)\.(?P<minor>\d+)-(?P<extra>.+)`, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTagPattern(regexp.MustCompile(tt.pattern))
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewTagPattern(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
			if tt.wantErr && got != nil {
				t.Errorf("NewTagPattern(%q) returned a pattern alongside its error", tt.pattern)
			}
		})
	}
}

// segmentIndex is the inverse of the naming helpers: it must answer 0 for
// every name that addresses no segment, including near-misses on the prefix.
func TestSegmentIndex(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{name: "major", want: 1},
		{name: "minor", want: 2},
		{name: "patch", want: 3},
		{name: "level1", want: 1},
		{name: "level3", want: 3},
		{name: "level42", want: 42},
		// Names that address no segment at all
		{name: "extra", want: 0},
		{name: "", want: 0},
		{name: "host", want: 0},
		// The prefix alone is not a segment
		{name: "level", want: 0},
		{name: "levelfoo", want: 0},
		{name: "level0", want: 0},
		{name: "level-1", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := segmentIndex(tt.name); got != tt.want {
				t.Errorf("segmentIndex(%q) = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}

// Group positions are resolved once, at construction.
func TestNewTagPatternResolvesGroups(t *testing.T) {
	t.Run("segments in order", func(t *testing.T) {
		p, err := NewTagPattern(regexp.MustCompile(`^(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)$`))
		if err != nil {
			t.Fatalf("NewTagPattern returned unexpected error: %v", err)
		}
		if len(p.groups) != 3 {
			t.Fatalf("resolved %d segments, want 3", len(p.groups))
		}
		// Group 0 is the whole match, so a named group always resolves past it
		for n, group := range p.groups {
			if group < 1 {
				t.Errorf("segment %d resolved to group %d, want >= 1", n+1, group)
			}
		}
	})

	// Lookup is by name, so declaration order in the pattern is irrelevant
	t.Run("segments declared out of order", func(t *testing.T) {
		p, err := NewTagPattern(regexp.MustCompile(`^(?P<minor>\d+)-(?P<major>\d+)$`))
		if err != nil {
			t.Fatalf("NewTagPattern returned unexpected error: %v", err)
		}
		if p.groups[0] <= p.groups[1] {
			t.Errorf("expected major to resolve after minor, got %v", p.groups)
		}
	})

	t.Run("absent extra resolves to the sentinel", func(t *testing.T) {
		p, err := NewTagPattern(regexp.MustCompile(`^(?P<major>\d+)\.(?P<minor>\d+)$`))
		if err != nil {
			t.Fatalf("NewTagPattern returned unexpected error: %v", err)
		}
		if p.extra != -1 {
			t.Errorf("extra = %d, want -1 when the pattern defines no label", p.extra)
		}
	})
}

// The built-in pattern goes through the same construction as any custom one.
func TestDefaultTagPattern(t *testing.T) {
	if defaultTagPattern == nil {
		t.Fatal("defaultTagPattern is nil")
	}
	if len(defaultTagPattern.groups) != 3 {
		t.Errorf("defaultTagPattern resolved %d segments, want 3", len(defaultTagPattern.groups))
	}
	if defaultTagPattern.extra < 1 {
		t.Errorf("defaultTagPattern.extra = %d, want a resolved group", defaultTagPattern.extra)
	}
}
