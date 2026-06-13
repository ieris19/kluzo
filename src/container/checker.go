package container

import (
    "git.ierislabs.dev/update-link/data"
)

type Update struct {
    ContainerDefinition data.ContainerDefinition
    LatestVersion       data.SemanticVersion
    Upgradeable         bool
}

func CheckUpdate(definition data.ContainerDefinition) (Update, error) {
    fetcher := DistributionFetcher{}

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
