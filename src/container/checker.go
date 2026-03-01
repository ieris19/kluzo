package container

import (
    "errors"
    "fmt"
    "io"
    "net/http"

    "ieris19.com/podman-updater/data"
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
        msg := fmt.Sprintf("could not check updates: %v\n", err)
        return Update{}, errors.New(msg)
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

func dryRun(compared ComparisonResult, update Update) {
    switch compared {
    case GreaterThan:
        fmt.Printf("Update available for %s: %s -> %s\n", update.ContainerDefinition.File.Path, update.ContainerDefinition.Version.String(), update.LatestVersion.String())
    case Equal:
        fmt.Printf("The latest release for %s is %s\n", update.ContainerDefinition.File.Path, update.ContainerDefinition.Version.String())
    case LessThan:
        fmt.Printf("Local version for %s is newer than upstream: %s > %s\n", update.ContainerDefinition.File.Path, update.ContainerDefinition.Version.String(), update.LatestVersion.String())
    }
}

func fetch(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("failed to fetch URL %s: status code %d", url, resp.StatusCode)
    }
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    return string(body), nil
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
