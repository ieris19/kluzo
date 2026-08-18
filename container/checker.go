package container

import (
	"git.ierislabs.dev/ieris19/update-link/data"
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

	update := data.Update{
		Definition:    definition,
		LatestVersion: latest,
		Upgradeable:   latest.GreaterThan(definition.Version),
	}

	if !update.Upgradeable && (definition.Pin == data.PinMajor || definition.Pin == data.PinMinor) {
		unpinned := definition
		unpinned.Pin = data.PinChannel
		if unconstrainedLatest, err := selectLatestTag(tags, unpinned); err == nil {
			update.Pinned = unconstrainedLatest.GreaterThan(definition.Version)
		}
	}

	return update, nil
}
