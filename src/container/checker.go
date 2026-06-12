package container

import (
    "git.ierislabs.dev/update-link/data"
)

type ComparisonResult int

const (
    GreaterThan ComparisonResult = 1
    LessThan    ComparisonResult = -1
    Equal       ComparisonResult = 0
)

type Update struct {
    ContainerDefinition data.ContainerDefinition
    LatestVersion       data.SemanticVersion
    Upgradeable         bool
    Unsure              bool
}

func CheckUpdate(definition data.ContainerDefinition) (Update, error) {
    var fetcher TagFetcher
    switch definition.Upstream {
    case data.UpstreamGitHub:
        fetcher = GitHubFetcher{}
    case data.UpstreamDockerHub:
        fetcher = DockerHubFetcher{}
    default:
        fetcher = DistributionFetcher{}
    }

    tags, err := fetcher.FetchTags(definition.Image)
    if err != nil {
        return Update{}, err
    }

    latest, err := selectLatestTag(tags, definition.Version.Extra)
    if err != nil {
        return Update{}, err
    }

    return Update{
        ContainerDefinition: definition,
        LatestVersion:       latest,
        Upgradeable:         latest.GreaterThan(definition.Version),
    }, nil
}
