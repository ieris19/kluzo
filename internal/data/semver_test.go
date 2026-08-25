package data

import (
	"regexp"
	"testing"
)

// Cases the default pattern should accept.
func TestParseSemanticVersionValidFormats(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want SemanticVersion
	}{
		{
			name: "plain major.minor.patch",
			tag:  "1.2.3",
			want: SemanticVersion{Major: 1, Minor: 2, Patch: 3, Raw: "1.2.3"},
		},
		{
			name: "v-prefixed",
			tag:  "v1.2.3",
			want: SemanticVersion{Major: 1, Minor: 2, Patch: 3, Raw: "v1.2.3"},
		},
		{
			name: "no patch defaults to 0",
			tag:  "1.2",
			want: SemanticVersion{Major: 1, Minor: 2, Patch: 0, Raw: "1.2"},
		},
		{
			name: "with extra label",
			tag:  "1.2.3-alpine",
			want: SemanticVersion{Major: 1, Minor: 2, Patch: 3, Extra: "alpine", Raw: "1.2.3-alpine"},
		},
		{
			name: "extra label containing dots",
			tag:  "v2.10.100-rc.1",
			want: SemanticVersion{Major: 2, Minor: 10, Patch: 100, Extra: "rc.1", Raw: "v2.10.100-rc.1"},
		},
		{
			name: "extra label with dashes",
			tag:  "v2.10.100-rc1-alpine",
			want: SemanticVersion{Major: 2, Minor: 10, Patch: 100, Extra: "rc1-alpine", Raw: "v2.10.100-rc1-alpine"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSemanticVersion(tt.tag, nil)
			if err != nil {
				t.Fatalf("ParseSemanticVersion(%q) returned unexpected error: %v", tt.tag, err)
			}
			if got != tt.want {
				t.Errorf("ParseSemanticVersion(%q) = %+v, want %+v", tt.tag, got, tt.want)
			}
		})
	}
}

// Cases the default pattern should reject.
func TestParseSemanticVersionInvalidFormats(t *testing.T) {
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
			_, err := ParseSemanticVersion(tt.tag, nil)
			if err == nil {
				t.Errorf("ParseSemanticVersion(%q) succeeded, want error", tt.tag)
			}
		})
	}
}

// Regression coverage for the Atoi-discard bug: a tag pattern whose
// major/minor named groups can capture non-digits, or a digit run too large
// for int, must fail loudly instead of silently collapsing to 0.
func TestParseSemanticVersionNonNumericCaptures(t *testing.T) {
	t.Run("custom pattern with non-numeric major/minor", func(t *testing.T) {
		re := regexp.MustCompile(`^(?P<major>[a-z]+)\.(?P<minor>[a-z]+)$`)
		_, err := ParseSemanticVersion("abc.def", re)
		if err == nil {
			t.Fatal("ParseSemanticVersion succeeded on non-numeric major/minor, want error")
		}
	})

	t.Run("custom pattern with non-numeric patch", func(t *testing.T) {
		re := regexp.MustCompile(`^(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>[a-z]+)$`)
		_, err := ParseSemanticVersion("1.2.abc", re)
		if err == nil {
			t.Fatal("ParseSemanticVersion succeeded on non-numeric patch, want error")
		}
	})

	t.Run("major overflows int", func(t *testing.T) {
		_, err := ParseSemanticVersion("99999999999999999999.1.0", nil)
		if err == nil {
			t.Fatal("ParseSemanticVersion succeeded on overflowing major, want error")
		}
	})
}

// Validate custom patterns enforce group presence correctly
func TestValidateTagPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		// Patch is optional
		{name: "has major and minor", pattern: `(?P<major>\d+)\.(?P<minor>\d+)`, wantErr: false},
		// Major is mandatory
		{name: "missing major", pattern: `(?P<minor>\d+)`, wantErr: true},
		// Minor is mandatory
		{name: "missing minor", pattern: `(?P<major>\d+)`, wantErr: true},
		// Major and Minor are mandatory
		{name: "missing both", pattern: `\d+\.\d+`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTagPattern(regexp.MustCompile(tt.pattern))
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTagPattern(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

// Comparison logic test
func TestSemanticVersionCompare(t *testing.T) {
	tests := []struct {
		name string
		a, b SemanticVersion
		want int
	}{
		{name: "equal", a: SemanticVersion{1, 2, 3, "", ""}, b: SemanticVersion{1, 2, 3, "", ""}, want: 0},
		{name: "major less", a: SemanticVersion{1, 9, 9, "", ""}, b: SemanticVersion{2, 0, 0, "", ""}, want: -1},
		{name: "major greater", a: SemanticVersion{2, 0, 0, "", ""}, b: SemanticVersion{1, 9, 9, "", ""}, want: 1},
		{name: "minor less", a: SemanticVersion{1, 1, 9, "", ""}, b: SemanticVersion{1, 2, 0, "", ""}, want: -1},
		{name: "minor greater", a: SemanticVersion{1, 2, 0, "", ""}, b: SemanticVersion{1, 1, 9, "", ""}, want: 1},
		{name: "patch less", a: SemanticVersion{1, 2, 3, "", ""}, b: SemanticVersion{1, 2, 4, "", ""}, want: -1},
		{name: "patch greater", a: SemanticVersion{1, 2, 4, "", ""}, b: SemanticVersion{1, 2, 3, "", ""}, want: 1},
		{name: "same extra label", a: SemanticVersion{1, 0, 0, "alpine", ""}, b: SemanticVersion{2, 0, 0, "alpine", ""}, want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.Compare(tt.b)
			if err != nil {
				t.Fatalf("Compare returned unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Compare(%+v, %+v) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}

	// Extra is an arbitrary label, two differing extras are incomparable
	t.Run("differing extra label is incomparable", func(t *testing.T) {
		a := SemanticVersion{1, 0, 0, "trixie", ""}
		b := SemanticVersion{1, 0, 0, "alpine", ""}
		if _, err := a.Compare(b); err == nil {
			t.Error("Compare succeeded across differing extra labels, want error")
		}
	})
}

// Test equality logic
func TestSemanticVersionEquals(t *testing.T) {
	tests := []struct {
		name string
		a, b SemanticVersion
		want bool
	}{
		{name: "identical", a: SemanticVersion{1, 2, 3, "alpine", ""}, b: SemanticVersion{1, 2, 3, "alpine", ""}, want: true},
		{name: "differing major", a: SemanticVersion{1, 2, 3, "", ""}, b: SemanticVersion{2, 2, 3, "", ""}, want: false},
		{name: "differing minor", a: SemanticVersion{1, 2, 3, "", ""}, b: SemanticVersion{1, 3, 3, "", ""}, want: false},
		{name: "differing patch", a: SemanticVersion{1, 2, 3, "", ""}, b: SemanticVersion{1, 2, 4, "", ""}, want: false},
		// Unlike Compare, Equals can just compare Extra, it's not incomparable here
		{name: "differing extra only", a: SemanticVersion{1, 0, 0, "trixie", ""}, b: SemanticVersion{1, 0, 0, "alpine", ""}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Equals(tt.b); got != tt.want {
				t.Errorf("Equals(%+v, %+v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// String reconstruction test
func TestSemanticVersionString(t *testing.T) {
	tests := []struct {
		name string
		sv   SemanticVersion
		want string
	}{
		{name: "raw is preferred when set", sv: SemanticVersion{Major: 1, Minor: 0, Patch: 0, Raw: "v1.0.0"}, want: "v1.0.0"},
		{name: "no extra, no raw", sv: SemanticVersion{Major: 1, Minor: 2, Patch: 3}, want: "1.2.3"},
		{name: "with extra, no raw", sv: SemanticVersion{Major: 1, Minor: 2, Patch: 3, Extra: "alpine"}, want: "1.2.3-alpine"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sv.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
