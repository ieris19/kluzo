# Task List

The following document outlines several different kinds of tasks. Some are
planned features, some are currently WIP features and some are bugs/design
shortcomings that need to be addressed eventually.

## Current Work In Progress

- [X] Refactor the container update check to be more universal
- [X] Remove reliance on "Known Domains", instead, try distribution and fail
  if the server does not behave as expected
- [ ] Delete unused code paths after the new universal update checker
- [ ] Bring parity to the quadlet parser by recursively searching for files

## Fixes

- [X] Harden SemVer parsing and handling. There are too many detections that
  are not SemVer tags and the default 0 values cause issues like 1.2 and 1.2.0
  to be considered equivalent when the latter is preferred.
- [X] Handle containers that might not have semver tags, just notify the
  user that these containers cannot be parsed.
- [X] Rely on GHCR instead of checking releases, since not all projects
  necessarily make releases for every tag. Be mindful of detecting beta and
  other potentially unstable updates.

## Planned Features

- [ ] Add support for compose files
- [ ] Add support for comparing digests on non-semver tags
- [ ] Add additional management features beyond checks, such as actual upgrades
