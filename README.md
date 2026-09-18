<!-- mirror-notice:start -->
> [!NOTE]
> **This repository is a mirror.** The canonical source lives at:
> [git.ierislabs.dev/ieris19/kluzo](https://git.ierislabs.dev/ieris19/kluzo)
<!-- mirror-notice:end -->

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
directories to check for. The file is mandatory. Passing `--config` points the
tool at an exact file that must exist. Without `--config`, the following paths
are checked in order, and the first one that exists is used:

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

- `TagPattern`: defines the regex to match the tag as a semantic version.
  Segments are named either by convention (`major`, `minor`, `patch`) or by
  depth (`level1`, `level2`, … `levelN`). The two forms are interchangeable, but
  one segment must not answer to both names at once. At least two segments are
  required, and they must be contiguous, you can't define `level4` without
  `level3`. The optional `extra` group is the "channel" and is used for channel
  pinning. This only fixes tags that are oddly formatted or don't parse with the
  standard regex, and still requires the images to be tagged using semantic
  versioning.
    - Depth is arbitrary. `level1` through `level5` is as valid as
      `major`/`minor`/`patch`, and versions of differing depth are compared by
      padding the shorter one with zeroes. You can also mix and match, as long
      as you don't define the same level twice (e.g. `major` and
      `level1` in the same pattern is invalid)
    - A trailing segments can be optional, intermediate segments may not. If a
      pattern lets `minor` go unmatched while `patch` matches, the tag is
      ambiguous and is rejected rather than quietly read as a shorter version.
    - Every segment group must capture digits only. Else it will fail to parse
      any versions at runtime. The parser only rejects invalid patterns, but
      cannot introspect the pattern to verify it matches only digits.
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
- `VersionPin`: how many leading version segments a candidate tag must share.
  The policies are cumulative: a deeper pin carries every restriction above it.
  When a pin is the only thing holding a container back, that is, no update
  exists within the pin but a newer version exists outside it, the report says
  so rather than reporting the container as up to date. `freeze` does not check
  against upstream at all, so it never performs this check. The names below are
  the common depths, but any non-negative number is accepted.
    - `channel` (`0`) **(default)**: only the "channel" (`extra`) is pinned; any
      newer version within that channel is considered an update. E.g.
      `1.0.0-alpine` will only match other tags ending in `-alpine`.
    - `major` (`1`): also restricts tags to the same first segment, useful for
      projects where crossing a major version requires manual intervention.
    - `minor` (`2`): also restricts tags to the same first two segments, only
      patch releases are suggested as possible updates.
    - Any deeper pin is written as a plain number. Pinning deeper than a tag
      actually goes matches everything, since absent segments are padded with
      zeroes.
    - `freeze`: the upstream registry is not checked at all; the container is
      always reported as frozen at its current version. This sits outside the
      scale above; it is an exception, not a depth, and cannot be written as a
      number.
- `SemVer`: **(default `true`)** set to `false` to declare that this container's
  tag carries no comparable version (e.g. `latest`). Tags with no digits at all
  are already detected automatically; this is for the cases the automatic check
  can't tell apart from a genuinely malformed version. Reported as skipped, not
  an error.

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
- Versions must be valid semantic version tags. Digest-pinned and untagged
  images are treated as errors. Purely textual tags (e.g. `latest`) are skipped
  automatically rather than erroring; tags that mix digits and text but still
  aren't valid semver (e.g. `rc1`) are treated as errors. `SemVer=false`
  overrides all of the above and skips the container outright, whatever the
  reference looks like (see "Customized behavior" above).
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
