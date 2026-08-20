# Kluzo

A CLI tool that checks container tags in file definitions against the upstream
registries and reports which containers have updates available.

The tool focuses mostly on Podman Quadlets for now (`.container`). Support for
docker compose and other formats is planned in the future.

## Usage

```
kluzo [--config /path/to/config.toml]
```

## Configuration

The configuration file is the source of many important settings, such as what
directories to check for. The file is mandatory and is read by default from
`/etc/kluzo/config.toml`. This path can be overridden using `--config`
which allows you to choose an arbitrary file.

A sample configuration can be found at `config/sample.toml`. You can use it as a
starting point.

For further reference, please refer to the `config` package.

## Customized behavior

A tool trying to manage container configuration is bound to hit an image that
does not conform to standards or needs some sort of special treatment. This is
achieved by adding custom attributes to the container definitions, allowing
certain behaviors to be altered on a per-image basis.

As of right now, the following keys are respected under the `[X-Update]`
in Quadlet files:

- `TagPattern`: Must define 2 named groups `major` and `minor`; `patch` and
  `extra` are optional, the rest correspond to semantic versioning, `extra`
  will be used for channel pinning. This only fixes tags that are oddly
  formatted or don't parse with the standard regex, and still requires the
  images to be tagged using semantic version.
    - For example, an image tagged with dates can use
      `^(?P<major>\d+)-(?P<minor>\d+)(?:-(?P<patch>\d+))?$` and match
      `2026-08-18` as a semantic version, `18-08-2026` and `08-18-2026`
      could also match with small tweaks.
- `ImagePattern`: the base assumption is that images will have a stable form,
  that is `host/user/name:tag@digest`, when that assumption does not apply, you
  can supply your own regex, recognizing the following named groups:
  `host`, `user`, `name`, `tag`, `digest`. Only `name` is mandatory, but without
  `tag`, an error will be reported. Some repositories nest images in deeper or
  shallower URLs, so feel free to adjust to match your registries.
- `VersionPin`: defines a cumulative pinning policy. The base strategy applies
  at all times, stricter formats also check the looser formats. In order, from
  looser to stricter, the following are the allowed policies:
    - `channel` **(default)**: only the "channel" (`extra`) is pinned; any newer
      version within that channel is considered an update. E.g. `1.0.0-alpine`
      will only match other tags ending in `-alpine`.
    - `major`: also restricts tags to the same major version, useful for
      projects where crossing a major version requires manual intervention.
    - `minor`: also restricts tags to the same minor version, only patch
      releases are suggested as possible updates.
    - `freeze`: the upstream registry is not checked at all; the container is
      always reported as frozen at its current version.

For `ImagePattern`, by convention, the sections are called `user` and `name`,
however, OCI image names don't make such distinction. It's all a "repository" to
the container runtime. Thus, when `user` is absent, name must be the full
repository path, if `user` is present, then this project constructs the
repository path as `user/name`. If this convention does not apply to your
registry, you can simply use the name capture group for the whole repository
segment.

## Limitations

- Only Quadlet `.container` files are supported, for now.
- Upstream registries must expose the [OCI Distribution][spec]
  tag-listing API.
- Aliases may be needed for registries whose public hostname differs from their
  API endpoint (e.g. `docker.io` is a hardcoded alias to
  `registry-1.docker.io`).
- Versions must be valid semantic version tags. Digest-pinned, textual tags or
  untagged images are treated as non-fatal errors for now.
- Channel suffixes (e.g. `-rc1`, `-alpine`, `-trixie`) are compared opaquely,
  they're not ordered against each other. This is a gotcha that isn't obvious in
  certain scenarios:
    - `1.0.0-rc1` and `1.0.0-rc2` are entirely different channels, `-rc2`
      will never be considered an updated to `-rc1`.
    - A bare release (`1.0.0`) is just another channel compared to a suffixed
      tag (`1.0.0-rc1`), no different than what  `-alpine` vs `-trixie`
      would be.
    - If you really need to override this behavior, perhaps you can try and
      define a custom pattern that ignores extra, or includes the release
      candidate as a patch version.

## License

Copyright (C) 2026 ieris19

This project is licensed under the GPL-3.0-only.

See [`LICENSE`](LICENSE) for the full text.

[spec]: https://github.com/opencontainers/distribution-spec
