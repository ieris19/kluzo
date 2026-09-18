package semver

import "testing"

// buildVersion builds a Version carrying only segments, the common case in these tables.
func buildVersion(segments ...int) Version {
	return Version{Segments: segments}
}

// Segments past the end read as 0, and so do out-of-range requests.
func TestVersionSegment(t *testing.T) {
	v := buildVersion(1, 2, 3)
	tests := []struct {
		name string
		n    int
		want int
	}{
		{name: "first", n: 1, want: 1},
		{name: "last", n: 3, want: 3},
		{name: "past the end pads", n: 4, want: 0},
		{name: "zero is not a segment", n: 0, want: 0},
		{name: "negative is not a segment", n: -1, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := v.Segment(tt.n); got != tt.want {
				t.Errorf("Segment(%d) = %d, want %d", tt.n, got, tt.want)
			}
		})
	}
}

func TestVersionDepth(t *testing.T) {
	tests := []struct {
		name string
		v    Version
		want int
	}{
		{name: "zero value", v: Version{}, want: 0},
		{name: "two segments", v: buildVersion(1, 2), want: 2},
		{name: "four segments", v: buildVersion(1, 2, 3, 4), want: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.Depth(); got != tt.want {
				t.Errorf("Depth() = %d, want %d", got, tt.want)
			}
		})
	}
}

// Comparison logic test
func TestVersionCompare(t *testing.T) {
	tests := []struct {
		name string
		a, b Version
		want int
	}{
		{name: "equal", a: buildVersion(1, 2, 3), b: buildVersion(1, 2, 3), want: 0},
		{name: "major less", a: buildVersion(1, 9, 9), b: buildVersion(2, 0, 0), want: -1},
		{name: "major greater", a: buildVersion(2, 0, 0), b: buildVersion(1, 9, 9), want: 1},
		{name: "minor less", a: buildVersion(1, 1, 9), b: buildVersion(1, 2, 0), want: -1},
		{name: "minor greater", a: buildVersion(1, 2, 0), b: buildVersion(1, 1, 9), want: 1},
		{name: "patch less", a: buildVersion(1, 2, 3), b: buildVersion(1, 2, 4), want: -1},
		{name: "patch greater", a: buildVersion(1, 2, 4), b: buildVersion(1, 2, 3), want: 1},
		// Differing depths pad with 0 rather than ordering by length
		{name: "missing segment pads to equal", a: buildVersion(1, 2), b: buildVersion(1, 2, 0), want: 0},
		{name: "missing segment still orders", a: buildVersion(1, 2), b: buildVersion(1, 2, 1), want: -1},
		{name: "deeper is not automatically greater", a: buildVersion(1, 2, 0, 0), b: buildVersion(1, 3), want: -1},
		{
			name: "same extra label",
			a:    Version{Segments: []int{1, 0, 0}, Extra: "alpine"},
			b:    Version{Segments: []int{2, 0, 0}, Extra: "alpine"},
			want: -1,
		},
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
		a := Version{Segments: []int{1, 0, 0}, Extra: "trixie"}
		b := Version{Segments: []int{1, 0, 0}, Extra: "alpine"}
		if _, err := a.Compare(b); err == nil {
			t.Error("Compare succeeded across differing extra labels, want error")
		}
	})
}

// Test equality logic
func TestVersionEquals(t *testing.T) {
	tests := []struct {
		name string
		a, b Version
		want bool
	}{
		{
			name: "identical",
			a:    Version{Segments: []int{1, 2, 3}, Extra: "alpine"},
			b:    Version{Segments: []int{1, 2, 3}, Extra: "alpine"},
			want: true,
		},
		{name: "differing major", a: buildVersion(1, 2, 3), b: buildVersion(2, 2, 3), want: false},
		{name: "differing minor", a: buildVersion(1, 2, 3), b: buildVersion(1, 3, 3), want: false},
		{name: "differing patch", a: buildVersion(1, 2, 3), b: buildVersion(1, 2, 4), want: false},
		{name: "padded depths are equal", a: buildVersion(1, 2), b: buildVersion(1, 2, 0), want: true},
		// Unlike Compare, Equals can just compare Extra, it's not incomparable here
		{
			name: "differing extra only",
			a:    Version{Segments: []int{1, 0, 0}, Extra: "trixie"},
			b:    Version{Segments: []int{1, 0, 0}, Extra: "alpine"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Equals(tt.b); got != tt.want {
				t.Errorf("Equals(%+v, %+v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// Pin depth is expressed as a count of leading segments that must agree.
func TestVersionSharesPrefix(t *testing.T) {
	tests := []struct {
		name string
		a, b Version
		n    int
		want bool
	}{
		{name: "no segments pinned", a: buildVersion(1, 2, 3), b: buildVersion(9, 9, 9), n: 0, want: true},
		{name: "first segment agrees", a: buildVersion(1, 2, 3), b: buildVersion(1, 9, 9), n: 1, want: true},
		{name: "first segment differs", a: buildVersion(1, 2, 3), b: buildVersion(2, 2, 3), n: 1, want: false},
		{name: "two segments agree", a: buildVersion(1, 2, 3), b: buildVersion(1, 2, 9), n: 2, want: true},
		{name: "second segment differs", a: buildVersion(1, 2, 3), b: buildVersion(1, 3, 3), n: 2, want: false},
		{name: "deeper pin than either carries", a: buildVersion(1, 2), b: buildVersion(1, 2), n: 5, want: true},
		// Padding makes an absent segment agree with an explicit zero
		{name: "absent segment matches zero", a: buildVersion(1, 2), b: buildVersion(1, 2, 0), n: 3, want: true},
		{name: "absent segment differs from nonzero", a: buildVersion(1, 2), b: buildVersion(1, 2, 1), n: 3, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.SharesPrefix(tt.b, tt.n); got != tt.want {
				t.Errorf("SharesPrefix(%+v, %d) = %v, want %v", tt.b, tt.n, got, tt.want)
			}
		})
	}
}

// String reconstruction test
func TestVersionString(t *testing.T) {
	tests := []struct {
		name string
		v    Version
		want string
	}{
		{name: "raw is preferred when set", v: Version{Segments: []int{1, 0, 0}, Raw: "v1.0.0"}, want: "v1.0.0"},
		{name: "no extra, no raw", v: buildVersion(1, 2, 3), want: "1.2.3"},
		{name: "with extra, no raw", v: Version{Segments: []int{1, 2, 3}, Extra: "alpine"}, want: "1.2.3-alpine"},
		{name: "arbitrary depth", v: buildVersion(1, 2, 3, 4, 5), want: "1.2.3.4.5"},
		// Without segments there is no version to render, only the zero value
		{name: "zero value", v: Version{}, want: "<none>"},
		{name: "extra without segments", v: Version{Extra: "alpine"}, want: "<none>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
