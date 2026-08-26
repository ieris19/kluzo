package container

import (
	"fmt"
	"slices"
	"strings"

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

func selectLatestTag(tags []string, definition data.ContainerDefinition) (data.SemanticVersion, error) {
	if definition.Pin == data.PinFreeze {
		return data.SemanticVersion{}, fmt.Errorf("cannot select a tag for a frozen container")
	}
	current := definition.Version
	tagPattern := definition.TagPattern
	var candidates []data.SemanticVersion
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
		v, err := data.ParseSemanticVersion(tag, tagPattern)
		// A looser rule is always a subset of the restrictions in a stricter rule
		if err != nil || v.Extra != current.Extra {
			continue
		}
		if definition.Pin >= data.PinMajor && v.Major != current.Major {
			continue
		}
		if definition.Pin >= data.PinMinor && (v.Minor != current.Minor) {
			continue
		}
		candidates = append(candidates, v)
	}
	if len(candidates) == 0 {
		return data.SemanticVersion{}, fmt.Errorf("no matching tags found")
	}
	slices.SortFunc(candidates, func(a, b data.SemanticVersion) int {
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
