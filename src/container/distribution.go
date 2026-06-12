package container

import (
    "encoding/json"
    "fmt"

    "git.ierislabs.dev/update-link/data"
)

type DistributionResponse struct {
    Name string   `json:"name"`
    Tags []string `json:"tags"`
}

type DistributionFetcher struct{}

func (f DistributionFetcher) FetchTags(image data.ImageInfo) ([]string, error) {
    imagePath := image.Name
    if image.Author != "" {
        imagePath = image.Author + "/" + image.Name
    }
    url := fmt.Sprintf("https://%s/v2/%s/tags/list", image.Host, imagePath)
    resp, err := fetch(url)
    if err != nil {
        return nil, fmt.Errorf("could not fetch distribution tags for %s/%s: %v", image.Author, image.Name, err)
    }
    var distTags DistributionResponse
    if err = json.Unmarshal([]byte(resp), &distTags); err != nil {
        return nil, fmt.Errorf("could not parse distribution tags for %s/%s: %v", image.Author, image.Name, err)
    }
    return distTags.Tags, nil
}
