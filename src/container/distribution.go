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

func fetchDistributionTags(definition data.ContainerDefinition) (string, error) {
    url := fmt.Sprintf("https://%s/v2/%s/tags/list", definition.Image.Host, definition.Image.Name)
    resp, err := fetch(url)
    if err != nil {
        return "", err
    }
    return resp, nil
}

func parseDistributionLatestTag(response string, definition data.ContainerDefinition) (data.SemanticVersion, error) {
    var distTags DistributionResponse
    err := json.Unmarshal([]byte(response), &distTags)
    if err != nil {
        return data.SemanticVersion{}, err
    }

    if len(distTags.Tags) == 0 {
        return data.SemanticVersion{}, fmt.Errorf("no tags found")
    }

    // Find the latest tag in semantic versioning format
    var latestTag data.SemanticVersion = data.SemanticVersion{Major: 0, Minor: 0, Patch: 0, Extra: ""}
    for _, tag := range distTags.Tags {
        if data.SemverRegexp.MatchString(tag) {
            newMatch, err := data.ParseSemanticVersion(tag)
            if err != nil {
                continue
            }
            if newMatch.GreaterThan(latestTag) &&
                newMatch.Extra == definition.Version.Extra {
                latestTag = newMatch
            }
        }
    }

    return latestTag, nil
}

func checkDistributionUpdates(definition data.ContainerDefinition) (data.SemanticVersion, error) {
    response, err := fetchDistributionTags(definition)
    if err != nil {
        return data.SemanticVersion{}, err
    }

    latestTag, err := parseDistributionLatestTag(response, definition)
    if err != nil {
        return data.SemanticVersion{}, err
    }

    return latestTag, nil
}
