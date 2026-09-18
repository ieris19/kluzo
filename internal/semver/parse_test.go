package semver

import (
	"errors"
	"reflect"
	"testing"
)

// Cases the default pattern should accept.
func TestParseValidFormats(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want Version
	}{
		{
			name: "plain major.minor.patch",
			tag:  "1.2.3",
			want: Version{Segments: []int{1, 2, 3}, Raw: "1.2.3"},
		},
		{
			name: "v-prefixed",
			tag:  "v1.2.3",
			want: Version{Segments: []int{1, 2, 3}, Raw: "v1.2.3"},
		},
		{
			// A trailing segment may simply be absent, rather than defaulting to 0
			name: "no patch yields two segments",
			tag:  "1.2",
			want: Version{Segments: []int{1, 2}, Raw: "1.2"},
		},
		{
			name: "with extra label",
			tag:  "1.2.3-alpine",
			want: Version{Segments: []int{1, 2, 3}, Extra: "alpine", Raw: "1.2.3-alpine"},
		},
		{
			name: "extra label containing dots",
			tag:  "v2.10.100-rc.1",
			want: Version{Segments: []int{2, 10, 100}, Extra: "rc.1", Raw: "v2.10.100-rc.1"},
		},
		{
			name: "extra label with dashes",
			tag:  "v2.10.100-rc1-alpine",
			want: Version{Segments: []int{2, 10, 100}, Extra: "rc1-alpine", Raw: "v2.10.100-rc1-alpine"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.tag, nil)
			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.tag, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse(%q) = %+v, want %+v", tt.tag, got, tt.want)
			}
		})
	}
}

// Cases the default pattern should reject.
func TestParseInvalidFormats(t *testing.T) {
	tests := []struct {
		name string
		tag  string
	}{
		{name: "not a semantic version at all", tag: "latest"},
		{name: "empty string", tag: ""},
		{name: "major only, no minor", tag: "1"},
		{name: "trailing segment past patch", tag: "1.2.3.4"},
		{name: "dash with no label", tag: "1.2.3-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.tag, nil)
			if err == nil {
				t.Errorf("Parse(%q) succeeded, want error", tt.tag)
			}
		})
	}
}

// A tag carrying no digits can never be a version under any pattern, and is
// marked so callers can skip it rather than report it as a failure.
func TestParseUntrackableTags(t *testing.T) {
	tests := []struct {
		name          string
		tag           string
		wantNotSemver bool
	}{
		{name: "no digits is untrackable", tag: "latest", wantNotSemver: true},
		{name: "empty is untrackable", tag: "", wantNotSemver: true},
		// Digits mean it might be a version in an unexpected shape, worth reporting
		{name: "digits in an odd shape is an error", tag: "20240101", wantNotSemver: false},
		{name: "partial version is an error", tag: "1.2.3.4", wantNotSemver: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.tag, nil)
			if err == nil {
				t.Fatalf("Parse(%q) succeeded, want error", tt.tag)
			}
			if got := errors.Is(err, ErrNotSemver); got != tt.wantNotSemver {
				t.Errorf("errors.Is(err, ErrNotSemver) = %v, want %v (err: %v)", got, tt.wantNotSemver, err)
			}
		})
	}
}

// Regression coverage for the Atoi-discard bug: a tag pattern whose
// major/minor named groups can capture non-digits, or a digit run too large
// for int, must fail loudly instead of silently collapsing to 0.
func TestParseNonNumericCaptures(t *testing.T) {
	t.Run("custom pattern with non-numeric major/minor", func(t *testing.T) {
		p := mustTagPattern(`^(?P<major>[a-z]+)\.(?P<minor>[a-z]+)$`)
		if _, err := Parse("abc.def", p); err == nil {
			t.Fatal("Parse succeeded on non-numeric major/minor, want error")
		}
	})

	t.Run("custom pattern with non-numeric patch", func(t *testing.T) {
		p := mustTagPattern(`^(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>[a-z]+)$`)
		if _, err := Parse("1.2.abc", p); err == nil {
			t.Fatal("Parse succeeded on non-numeric patch, want error")
		}
	})

	t.Run("major overflows int", func(t *testing.T) {
		if _, err := Parse("99999999999999999999.1.0", nil); err == nil {
			t.Fatal("Parse succeeded on overflowing major, want error")
		}
	})
}

// A pattern may leave trailing segments optional, but an interior one absent
// makes the tag ambiguous: it must not silently truncate the version.
func TestParseInteriorGap(t *testing.T) {
	p := mustTagPattern(`^(?P<major>\d+)(?:\.(?P<minor>\d+))?\.(?P<patch>\d+)$`)

	t.Run("all segments present", func(t *testing.T) {
		got, err := Parse("1.2.3", p)
		if err != nil {
			t.Fatalf("Parse returned unexpected error: %v", err)
		}
		want := Version{Segments: []int{1, 2, 3}, Raw: "1.2.3"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Parse = %+v, want %+v", got, want)
		}
	})

	t.Run("interior segment absent", func(t *testing.T) {
		got, err := Parse("1.5", p)
		if err == nil {
			t.Fatalf("Parse succeeded with an interior gap, yielding %+v, want error", got)
		}
	})
}

// A pattern whose segments are all optional can match input that fills none of
// them. Matching is not the same as carrying a version.
func TestParseNoSegmentsCaptured(t *testing.T) {
	p := mustTagPattern(`^v(?P<major>\d+)?(?P<minor>\d+)?$`)

	got, err := Parse("v", p)
	if err == nil {
		t.Fatalf("Parse succeeded capturing no segments, yielding %+v, want error", got)
	}
}

// Patterns of arbitrary depth parse every segment they declare.
func TestParseCustomDepth(t *testing.T) {
	p := mustTagPattern(`^(?P<level1>\d+)\.(?P<level2>\d+)\.(?P<level3>\d+)\.(?P<level4>\d+)$`)

	got, err := Parse("1.2.3.4", p)
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}
	want := Version{Segments: []int{1, 2, 3, 4}, Raw: "1.2.3.4"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse = %+v, want %+v", got, want)
	}
}

// A custom pattern replaces the default outright, rather than extending it.
func TestParseCustomPatternReplacesDefault(t *testing.T) {
	p := mustTagPattern(`^release-(?P<major>\d+)_(?P<minor>\d+)$`)

	if _, err := Parse("1.2.3", p); err == nil {
		t.Error("Parse accepted a default-shaped tag under a custom pattern, want error")
	}

	got, err := Parse("release-4_7", p)
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}
	want := Version{Segments: []int{4, 7}, Raw: "release-4_7"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse = %+v, want %+v", got, want)
	}
}

// Group names carry the meaning, so a pattern is free to order them as it likes.
func TestParseGroupOrderIndependence(t *testing.T) {
	p := mustTagPattern(`^(?P<minor>\d+)-(?P<major>\d+)$`)

	got, err := Parse("7-2", p)
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}
	want := Version{Segments: []int{2, 7}, Raw: "7-2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse = %+v, want %+v", got, want)
	}
}

// Validating at construction keeps regexp.Compile errors separate from
// pattern-shape errors.
func TestMustTagPatternPanicsOnInvalidShape(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("mustTagPattern accepted a pattern with no segments, want panic")
		}
	}()
	mustTagPattern(`^(?P<major>\d+)$`)
}
