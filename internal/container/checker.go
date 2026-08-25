package container

import (
	"fmt"

	"git.ierislabs.dev/ieris19/kluzo/internal/data"
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

	cmp, err := latest.Compare(definition.Version)
	if err != nil {
		// Should never error, since Extra should never differ
		// An error here means an assumed invariant here has been broken
		return data.Update{}, fmt.Errorf("comparing latest tag to current version: %v", err)
	}

	update := data.Update{
		Definition:    definition,
		LatestVersion: latest,
		Upgradeable:   cmp > 0,
	}

	if !update.Upgradeable && (definition.Pin == data.PinMajor || definition.Pin == data.PinMinor) {
		unpinned := definition
		unpinned.Pin = data.PinChannel
		if unconstrainedLatest, err := selectLatestTag(tags, unpinned); err == nil {
			if cmp, err := unconstrainedLatest.Compare(definition.Version); err == nil {
				update.Pinned = cmp > 0
			} else {
				return data.Update{}, fmt.Errorf("comparing latest tag to current version: %v", err)
			}
		}
	}

	return update, nil
}
