package container

import (
    "errors"

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
    var (
        latest data.SemanticVersion
        err    error
    )
    switch definition.Upstream {
    case data.UpstreamGitHub:
        latest, err = checkGitHubUpdates(definition)
    case data.UpstreamDockerHub:
        latest, err = checkDockerHubUpdates(definition)
    case data.UpstreamDistribution:
        latest, err = checkDistributionUpdates(definition)
    default:
        return Update{}, errors.New("unsupported upstream: " + string(definition.Upstream))
    }
    if err != nil {
        return Update{}, err
    }

    compared := checkUpdates(latest, definition)
    update := Update{
        ContainerDefinition: definition,
        LatestVersion:       latest,
    }

    switch compared {
    case GreaterThan:
        update.Upgradeable = true
    case Equal, LessThan:
        update.Upgradeable = false
    }
    return update, nil
}

func checkUpdates(latestRelease data.SemanticVersion, definition data.ContainerDefinition) ComparisonResult {
    if latestRelease.GreaterThan(definition.Version) {
        return GreaterThan
    }
    if latestRelease.Equals(definition.Version) {
        return Equal
    }
    return LessThan
}
