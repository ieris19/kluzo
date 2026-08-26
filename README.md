# Kluzo

A CLI tool that checks container tags in file definitions against the upstream
registries and reports which containers have updates available.

The tool focuses mostly on Podman Quadlets for now (`.container`). Support for
docker compose and other formats is planned in the future.

## Name

Kluzo (/ˈkluzo/, see [Wiktionary][kluzo-dict]) is the Esperanto word for a canal
lock and sluice. Like a boat at a lock, containers are slowly raised to match
upstream. However, the English word "lock" means something entirely different
and opposite at a glance. On the other hand, "sluice" is already a name
saturated with other tools. Looking beyond English, many European languages
happen to share similar words for these two concepts, so the Esperanto word for
it made sense as a neutral, short, and easily pronounceable option.

## Usage

```
kluzo [--config /path/to/config.toml]
```

## Configuration

The configuration file is the source of many important settings, such as what
directories to check for. The file is mandatory. Passing `--config` points
the tool at an exact file that must exist. Without `--config`, the
following paths are checked in order, and the first one that exists is used:

1. `$XDG_CONFIG_HOME/kluzo/config.toml`, if `XDG_CONFIG_HOME` is set.
2. `~/.config/kluzo/config.toml` otherwise.
3. `/etc/kluzo/config.toml` as the system-wide fallback.

A sample configuration can be found at `config/sample.toml`. You can use it as a
starting point.

`scanner.exclude` skips files and directories that match the given patterns.
Patterns are relative to each `scanner.directories` entry, and use
[`filepath.Match`][filepath-match] under the hood, with its corresponding
syntax.

For further reference, please refer to the `config` package.

## Customized behavior

A tool trying to manage container configuration is bound to hit an image that
does not conform to standards or needs some sort of special treatment. This is
achieved by adding custom attributes to the container definitions, allowing
certain behaviors to be altered on a per-image basis.

As of right now, the following keys are respected under the `[X-Kluzo]` section
in Quadlet files:

- `TagPattern`: Must define 2 named groups `major` and `minor`; `patch` and
  `extra` are also recognized but optional. The `extra` group will be used for
  channel pinning. This only fixes tags that are oddly formatted or don't parse
  with the standard regex, and still requires the images to be tagged using
  semantic versioning.
    - For example, an image tagged with dates can use
      `^(?P<major>\d+)-(?P<minor>\d+)(?:-(?P<patch>\d+))?$` and match
      `2026-08-18` as a semantic version. `18-08-2026` and `08-18-2026`
      could also match with small tweaks.
- `ImagePattern`: the base assumption is that image identifiers will have a
  stable form. That is `host/user/name:tag@digest`, when that assumption does
  not apply, you can supply your own regex, recognizing the following named
  groups:
  `host`, `user`, `name`, `tag`, `digest`. Only `name` is mandatory, but without
  `tag`, an error will be reported. Some repositories nest images in deeper or
  shallower URLs, so feel free to adjust to match your registries.
- `VersionPin`: defines a cumulative pinning policy. The base strategy applies
  at all times. The stricter policies accumulate on top of the previous levels.
  In order, from looser to stricter, the following are the allowed policies:
    - `channel` **(default)**: only the "channel" (`extra`) is pinned; any newer
      version within that channel is considered an update. E.g. `1.0.0-alpine`
      will only match other tags ending in `-alpine`.
    - `major`: also restricts tags to the same major version, useful for
      projects where crossing a major version requires manual intervention.
    - `minor`: also restricts tags to the same minor version, only patch
      releases are suggested as possible updates.
    - `freeze`: the upstream registry is not checked at all; the container is
      always reported as frozen at its current version.

For `ImagePattern`, by convention, the sections are called `user` and `name`.
However, OCI image names don't make such distinction. It's all a "repository" to
the container runtime. Thus, when `user` is absent, `name` must be the full
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
- Versions must be valid semantic version tags. Digest-pinned, textual tags and
  untagged images are treated as non-fatal errors for now.
- Channel suffixes (e.g. `-rc1`, `-alpine`, `-trixie`) are compared opaquely,
  they're not ordered against each other. This is a gotcha that isn't obvious in
  certain scenarios:
    - `1.0.0-rc1` and `1.0.0-rc2` are entirely different channels, `-rc2`
      will never be considered an update to `-rc1`.
    - A bare release (`1.0.0`) is just another channel compared to a suffixed
      tag (`1.0.0-rc1`), no different from what `-alpine` vs `-trixie`
      would be.
    - If you really need to override this behavior, perhaps you can try and
      define a custom pattern that ignores extra, or includes the release
      candidate as a patch version.

## License

This project is licensed under the MIT License.

See [`LICENSE`](LICENSE) for the full text.

[spec]: https://github.com/opencontainers/distribution-spec
[filepath-match]: https://pkg.go.dev/path/filepath#Match
[kluzo-dict]: https://en.wiktionary.org/wiki/kluzo
