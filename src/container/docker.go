package container

import (
    "encoding/json"
    "fmt"
    "slices"
    "time"

    "git.ierislabs.dev/update-link/data"
)

type dockerTagList struct {
    Count   int                 `json:"count"`
    Results []dockerTagResponse `json:"results"`
}

type dockerTagResponse struct {
    Id         int       `json:"id"`
    Name       string    `json:"name"`
    LastPushed time.Time `json:"tag_last_pushed"`
}

func fetchDockerHub(author string, imageName string) (string, error) {
    url := fmt.Sprintf("https://hub.docker.com/v2/namespaces/%s/repositories/%s/tags?page_size=100", author, imageName)
    resp, err := fetch(url)
    if err != nil {
        return "", err
    }
    return resp, nil
}

func parseDockerHubLatestTag(response string, definition data.ContainerDefinition) (data.SemanticVersion, error) {
    var tagList dockerTagList
    err := json.Unmarshal([]byte(response), &tagList)
    if err != nil {
        return data.SemanticVersion{}, err
    }

    if len(tagList.Results) == 0 {
        return data.SemanticVersion{}, fmt.Errorf("no tags found")
    }

    // Sort tags by LastPushed
    slices.SortFunc(tagList.Results, func(previous, next dockerTagResponse) int {
        if previous.LastPushed.Before(next.LastPushed) {
            return 1
        }
        if previous.LastPushed.After(next.LastPushed) {
            return -1
        }
        return 0
    })

    // Traverse each tag and find the latest in semantic versioning format
    var latestTag data.SemanticVersion = data.SemanticVersion{Major: 0, Minor: 0, Patch: 0, Extra: ""}
    latestTagIdx := len(tagList.Results) - 1
    for idx, tag := range tagList.Results {
        if data.SemverRegexp.MatchString(tag.Name) {
            newMatch, err := data.ParseSemanticVersion(tag.Name)
            if err != nil {
                continue
            }
            if newMatch.GreaterThan(latestTag) &&
                newMatch.Extra == definition.Version.Extra &&
                tag.LastPushed.AddDate(0, 0, 1).After(tagList.Results[latestTagIdx].LastPushed) {
                latestTag = newMatch
                latestTagIdx = idx
            }
        }
    }

    return latestTag, nil
}

func checkDockerHubUpdates(definition data.ContainerDefinition) (data.SemanticVersion, error) {
    resp, err := fetchDockerHub(definition.Image.Author, definition.Image.Name)
    if err != nil {
        err = fmt.Errorf("could not fetch Docker Hub tags for %s/%s: %v\n", definition.Image.Author, definition.Image.Name, err)
        return data.SemanticVersion{}, err
    }

    latestTag, err := parseDockerHubLatestTag(resp, definition)
    if err != nil {
        err = fmt.Errorf("could not parse Docker Hub tags for %s/%s: %v\n", definition.Image.Author, definition.Image.Name, err)
        return data.SemanticVersion{}, err
    }

    return latestTag, nil
}
