package container

import (
    "encoding/json"
    "errors"
    "fmt"

    "git.ierislabs.dev/update-link/data"
)

type GitHubRelease struct {
    Owner       string `json:repo_owner`
    TagName     string `json:"tag_name"`
    PublishedAt string `json:"created_at"`
}

func fetchGitHubReleases(author string, imageName string) (string, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", author, imageName)
    resp, err := fetch(url)
    if err != nil {
        return "", err
    }
    return resp, nil
}

func parseLatestGitHubReleases(response string) (GitHubRelease, error) {
    var releases []GitHubRelease
    err := json.Unmarshal([]byte(response), &releases)
    if err != nil {
        return GitHubRelease{}, err
    }

    if len(releases) == 0 {
        return GitHubRelease{}, errors.New("no releases found")
    }
    return releases[0], nil
}

func parseGitHubReleaseVersion(tag string) (data.SemanticVersion, error) {
    semver, err := data.ParseSemanticVersion(tag)
    if err != nil {
        return data.SemanticVersion{}, err
    }
    return semver, nil
}

func checkGitHubUpdates(definition data.ContainerDefinition) (data.SemanticVersion, error) {
    resp, err := fetchGitHubReleases(definition.Image.Author, definition.Image.Name)
    if err != nil {
        fmt.Printf("Error fetching releases for %s/%s: %v\n", definition.Image.Author, definition.Image.Name, err)
        return data.SemanticVersion{}, err
    }

    latestRelease, err := parseLatestGitHubReleases(resp)
    if err != nil {
        fmt.Printf("Error parsing releases for %s/%s: %v\n", definition.Image.Author, definition.Image.Name, err)
        return data.SemanticVersion{}, err
    }

    latestVersion, err := parseGitHubReleaseVersion(latestRelease.TagName)
    if err != nil {
        fmt.Printf("Error parsing version for %s/%s: %v\n", definition.Image.Author, definition.Image.Name, err)
        return data.SemanticVersion{}, err
    }

    return latestVersion, nil
}
