package data

import (
	"fmt"
	"regexp"
	"slices"
)

// Remember, author/name must be the "repository" segment of an OCI image name
type ImageInfo struct {
	Host   string
	Author string
	Name   string
	Tag    string
	Digest string
}

func ValidateImagePattern(re *regexp.Regexp) error {
	if slices.Contains(re.SubexpNames(), "name") {
		return nil
	}
	return fmt.Errorf("image pattern must define a named group 'name'")
}
