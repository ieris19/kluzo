package container

import (
    "git.ierislabs.dev/update-link/data"
)

func CheckUpdate(definition data.ContainerDefinition) (data.Update, error) {
    fetcher := DistributionFetcher{}

    tags, err := fetcher.FetchTags(definition.Image)
    if err != nil {
        return data.Update{}, err
    }

    latest, err := selectLatestTag(tags, definition)
    if err != nil {
        return data.Update{}, err
    }

    return data.Update{
        Definition:    definition,
        LatestVersion: latest,
        Upgradeable:   latest.GreaterThan(definition.Version),
    }, nil
}
