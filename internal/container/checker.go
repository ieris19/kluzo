package container

import (
	"fmt"
	"slices"
	"strings"

	"git.ierislabs.dev/ieris19/kluzo/internal/data"
	"git.ierislabs.dev/ieris19/kluzo/internal/semver"
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

	// Would there be updates if the image was unpinned?
	if !update.Upgradeable && definition.Pin > data.PinChannel {
		unpinned := definition
		unpinned.Pin = data.PinChannel
		// Dropping the pin only widens the candidate set.
		// Asserting assumptions anyway, should be unreachable errors
		unconstrainedLatest, err := selectLatestTag(tags, unpinned)
		if err != nil {
			return data.Update{}, fmt.Errorf("selecting latest tag without the version pin: %v", err)
		}
		cmp, err := unconstrainedLatest.Compare(definition.Version)
		if err != nil {
			return data.Update{}, fmt.Errorf("comparing latest tag to current version: %v", err)
		}
		update.Pinned = cmp > 0
	}

	return update, nil
}

func selectLatestTag(tags []string, definition data.ContainerDefinition) (semver.Version, error) {
	if definition.Pin == data.PinFreeze {
		return semver.Version{}, fmt.Errorf("cannot select a tag for a frozen container")
	}
	current := definition.Version
	tagPattern := definition.TagPattern
	var candidates []semver.Version
	for _, tag := range tags {
		if tagPattern == nil {
			if current.Extra == "" {
				if strings.Contains(tag, "-") {
					continue
				}
			} else if !strings.Contains(tag, "-"+current.Extra) {
				continue
			}
		}
		v, err := semver.Parse(tag, tagPattern)
		// A looser rule is always a subset of the restrictions in a stricter rule
		if err != nil || v.Extra != current.Extra {
			continue
		}
		if !v.SharesPrefix(current, int(definition.Pin)) {
			continue
		}
		candidates = append(candidates, v)
	}
	if len(candidates) == 0 {
		return semver.Version{}, fmt.Errorf("no matching tags found")
	}
	slices.SortFunc(candidates, func(a, b semver.Version) int {
		c, err := a.Compare(b)
		if err != nil {
			// Should never error, since Extra should never differ at this point
			// An error here means an assumed invariant here has been broken
			panic("invariant violation: " + err.Error())
		}
		return -c // descending: latest first
	})
	return candidates[0], nil
}
