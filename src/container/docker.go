package container

import (
    "encoding/json"
    "fmt"

    "git.ierislabs.dev/update-link/data"
)

type dockerTagList struct {
    Count   int                 `json:"count"`
    Results []dockerTagResponse `json:"results"`
}

type dockerTagResponse struct {
    Id   int    `json:"id"`
    Name string `json:"name"`
}

type DockerHubFetcher struct{}

func (f DockerHubFetcher) FetchTags(image data.ImageInfo) ([]string, error) {
    namespace := image.Author
    if namespace == "" {
        namespace = "library"
    }
    url := fmt.Sprintf("https://hub.docker.com/v2/namespaces/%s/repositories/%s/tags?page_size=10", namespace, image.Name)
    resp, err := fetch(url)
    if err != nil {
        return nil, fmt.Errorf("could not fetch Docker Hub tags for %s/%s: %v", image.Author, image.Name, err)
    }
    var tagList dockerTagList
    if err = json.Unmarshal([]byte(resp), &tagList); err != nil {
        return nil, fmt.Errorf("could not parse Docker Hub tags for %s/%s: %v", image.Author, image.Name, err)
    }
    tags := make([]string, len(tagList.Results))
    for i, t := range tagList.Results {
        tags[i] = t.Name
    }
    return tags, nil
}
