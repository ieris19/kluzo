package parser

import (
	"regexp"
	"testing"

	"git.ierislabs.dev/ieris19/update-link/internal/config"
	"git.ierislabs.dev/ieris19/update-link/internal/data"
)

// Formats parseImageInfo is expected to accept without error
func TestParseImageInfoValidFormats(t *testing.T) {
	tests := []struct {
		name  string
		image string
		want  data.ImageInfo
	}{
		{
			name:  "bare name, no host, no tag",
			image: "image-name",
			want:  data.ImageInfo{Name: "image-name"},
		},
		{
			name:  "name with tag, no host",
			image: "image-name:1.21",
			want:  data.ImageInfo{Name: "image-name", Tag: "1.21"},
		},
		{
			name:  "obvious host, user and name with tag",
			image: "registry.example.com/library/image-name:1.21",
			want:  data.ImageInfo{Host: "registry.example.com", Author: "library", Name: "image-name", Tag: "1.21"},
		},
		{
			name:  "non-obvious host with explicit port, no user segment",
			image: "registry:5000/image-name:1.21",
			want:  data.ImageInfo{Host: "registry:5000", Name: "image-name", Tag: "1.21"},
		},
		{
			name:  "localhost, with no port, no user segment",
			image: "localhost:5000/image-name",
			want:  data.ImageInfo{Host: "localhost:5000", Name: "image-name"},
		},
		{
			name:  "digest instead of tag, no host",
			image: "image-name@sha256:abcdef0123456789",
			want:  data.ImageInfo{Name: "image-name", Digest: "sha256:abcdef0123456789"},
		},
		{
			name:  "host and digest",
			image: "registry.example.com/image-name@sha256:abcdef0123456789",
			want:  data.ImageInfo{Host: "registry.example.com", Name: "image-name", Digest: "sha256:abcdef0123456789"},
		},
		{
			name:  "single ambiguous path segment resolves to user, not host",
			image: "someuser/image-name",
			want:  data.ImageInfo{Author: "someuser", Name: "image-name"},
		},
		{
			name:  "single ambiguous path segment resolves to user, not host, with tag",
			image: "someuser/image-name:1.0",
			want:  data.ImageInfo{Author: "someuser", Name: "image-name", Tag: "1.0"},
		},
		{
			name:  "tag and digest combined",
			image: "image-name:1.21@sha256:abcdef0123456789",
			want:  data.ImageInfo{Name: "image-name", Tag: "1.21", Digest: "sha256:abcdef0123456789"},
		},
		{
			name:  "docker.io host with no user defaults author to library",
			image: "docker.io/ubuntu:latest",
			want:  data.ImageInfo{Host: "docker.io", Author: "library", Name: "ubuntu", Tag: "latest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseImageInfo(tt.image, nil)
			if err != nil {
				t.Fatalf("parseImageInfo(%q) returned unexpected error: %v", tt.image, err)
			}
			if got != tt.want {
				t.Errorf("parseImageInfo(%q) = %+v, want %+v", tt.image, got, tt.want)
			}
		})
	}
}

// Formats parseImageInfo is not built to understand.
func TestParseImageInfoInvalidFormats(t *testing.T) {
	tests := []struct {
		name  string
		image string
	}{
		{
			name:  "more than two path segments before the name",
			image: "gcr.io/project/team/image:1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseImageInfo(tt.image, nil)
			if err == nil {
				t.Errorf("parseImageInfo(%q) succeeded, want error", tt.image)
			}
		})
	}
}

// A caller-supplied pattern (the [X-UpdateLink] ImagePattern override
func TestParseImageInfoCustomPattern(t *testing.T) {
	pattern := regexp.MustCompile(`^(?P<host>[^/\s]*)/(?P<user>.*)/(?P<name>[^\s/:@]*):(?P<tag>[\w.-]*)$`)

	t.Run("accepts custom patter for parsing an image", func(t *testing.T) {
		got, err := parseImageInfo("gcr.io/project/team/image:1.0", pattern)
		if err != nil {
			t.Fatalf("parseImageInfo with custom pattern returned unexpected error: %v", err)
		}
		want := data.ImageInfo{Host: "gcr.io", Author: "project/team", Name: "image", Tag: "1.0"}
		if got != want {
			t.Errorf("parseImageInfo with custom pattern = %+v, want %+v", got, want)
		}
	})

	t.Run("rejects input the custom pattern doesn't match", func(t *testing.T) {
		_, err := parseImageInfo("image-name", pattern)
		if err == nil {
			t.Error("parseImageInfo with custom pattern succeeded, want error")
		}
	})
}

// Test that aliases are correctly applied
func TestParseImageInfoResolvesBuiltinHostAlias(t *testing.T) {
	// SetAliases affects package-level state (registryAliases), so restore it
	// once this test is done to avoid leaking into other tests in this package.
	SetAliases(config.BuiltinAliases)
	defer SetAliases(nil)

	image := "docker.io/user/image:tag"
	want := data.ImageInfo{
		Host:   "registry-1.docker.io",
		Author: "user",
		Name:   "image",
		Tag:    "tag",
	}

	got, err := parseImageInfo(image, nil)
	if err != nil {
		t.Fatalf("parseImageInfo returned unexpected error: %v", err)
	}

	if got != want {
		t.Errorf("parseImageInfo(%s) = %+v, want %+v", image, got, want)
	}
}
