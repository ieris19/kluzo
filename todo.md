# Task List

The following document outlines several different kinds of tasks. Some are
planned features, some are currently WIP features and some are bugs/design
shortcomings that need to be addressed eventually.

## Current Work In Progress

- [ ] Refactor the container update check to be more universal
- [ ] Remove reliance on "Known Domains", instead, try distribution and fail
  if the server does not behave as expected

## Fixes

- [ ] Harden SemVer parsing and handling. There are too many detections that
  are not SemVer tags and the default 0 values cause issues like 1.2 and 1.2.0
  to be considered equivalent when the latter is preferred.
- [ ] Handle containers that might not have semver tags, just notify the
  user that these containers cannot be parsed.
- [ ] Rely on GHCR instead of checking releases, since not all projects
  necessarily make releases for every tag. Be mindful of detecting beta and
  other potentially unstable updates.
- [ ] Improve updates

## Planned features

- [ ] Optional API authentication to avoid rate limits.
