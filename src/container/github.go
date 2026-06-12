package container

import (
    "encoding/json"
    "fmt"

    "git.ierislabs.dev/update-link/data"
)

type GitHubRelease struct {
    Owner       string `json:"repo_owner"`
    TagName     string `json:"tag_name"`
    PublishedAt string `json:"created_at"`
}

type GitHubFetcher struct{}

func (f GitHubFetcher) FetchTags(image data.ImageInfo) ([]string, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=10", image.Author, image.Name)
    resp, err := fetch(url)
    if err != nil {
        return nil, fmt.Errorf("could not fetch GitHub releases for %s/%s: %v", image.Author, image.Name, err)
    }
    var releases []GitHubRelease
    if err = json.Unmarshal([]byte(resp), &releases); err != nil {
        return nil, fmt.Errorf("could not parse GitHub releases for %s/%s: %v", image.Author, image.Name, err)
    }
    tags := make([]string, len(releases))
    for i, r := range releases {
        tags[i] = r.TagName
    }
    return tags, nil
}
