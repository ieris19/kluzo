package container

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"

	"git.ierislabs.dev/ieris19/kluzo/internal/data"
)

type DistributionResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// Ensure most images return most tags in one page
const tagPageSize = 1000

// A misbehaving registry must not loop forever.
const maxTagPages = 50

type DistributionFetcher struct{}

func (f DistributionFetcher) FetchTags(image data.ImageInfo) ([]string, error) {
	imagePath := image.Name
	if image.Author != "" {
		imagePath = image.Author + "/" + image.Name
	}
	registry := fmt.Sprintf("https://%s", image.Host)
	next := fmt.Sprintf("%s/v2/%s/tags/list?n=%d", registry, imagePath, tagPageSize)

	session := &registrySession{}
	var tags []string
	for page := 0; next != "" && page < maxTagPages; page++ {
		resp, header, err := session.fetch(next)
		if err != nil {
			return nil, fmt.Errorf("could not fetch distribution tags for %s/%s: %v", image.Author, image.Name, err)
		}
		var distTags DistributionResponse
		if err = json.Unmarshal([]byte(resp), &distTags); err != nil {
			return nil, fmt.Errorf("could not parse distribution tags for %s/%s: %v", image.Author, image.Name, err)
		}
		if len(distTags.Tags) == 0 {
			break
		}
		tags = append(tags, distTags.Tags...)

		following, err := nextPageURL(registry, header.Get("Link"))
		if err != nil {
			return nil, fmt.Errorf("could not follow tag pages for %s/%s: %v", image.Author, image.Name, err)
		}
		// A page pointing at itself would loop for as long as the cap allows
		if following == next {
			break
		}
		next = following
	}
	return tags, nil
}

var linkNextRegexp = regexp.MustCompile(`<([^>]+)>\s*;[^,]*rel="?next"?`)

// nextPageURL resolves the next page from a Link header. Registries express it
// as a URL relative to the registry root, and omit the header on the last page.
func nextPageURL(registry string, link string) (string, error) {
	if link == "" {
		return "", nil
	}
	match := linkNextRegexp.FindStringSubmatch(link)
	if match == nil {
		return "", nil
	}
	base, err := url.Parse(registry)
	if err != nil {
		return "", fmt.Errorf("invalid registry URL %s: %v", registry, err)
	}
	ref, err := url.Parse(match[1])
	if err != nil {
		return "", fmt.Errorf("invalid next page link %s: %v", match[1], err)
	}
	return base.ResolveReference(ref).String(), nil
}
