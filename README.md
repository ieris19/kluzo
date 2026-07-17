# UpdateLink

A Podman-focused CLI tool that checks container versions in Quadlet `.container`
files against their upstream registries and reports which containers have
updates available.

## Usage

```
update-link [--config /path/to/config.toml]
```

The config file defaults to `/etc/update-link/config.toml` if `--config` is not
specified.

## Configuration

You can find an example of a configuration file in the repository called
`sample.toml`. For further information, refer to the config package that manages
the configuration.

## Customized behavior

For extensible, per image configuration of this tool's behavior, you can add
special keys to your container definitions under the `[X-UpdateLink]`
section. The following options are supported at the current time:

- `TagPattern`: for tags that aren't standard semantic versions. Must define
  named groups `major` and `minor`; `patch` and `extra` are optional but
  accounted for, `extra` will be used for channel pinning, the rest is just
  semantic versioning.
- `ImagePattern`: if your registry cannot be parsed by the default regex, you
  can override this. For example, registries that demand more than 2 segments
  for the image repository (assuming the segments to be:
  `host/repository:tag@digest`). Must define a named group `name` while `host`,
  `user`, `tag` and `digest`are all optional.
- `VersionPin`: determines how the current tag is matched against upstream 
  tags. They're cumulative in the following order:
  - `channel` (default): only the channel (`extra`) is pinned; any newer
    version within that channel is considered an update.
  - `major`: also restricts tags to the same major version, e.g. for
    projects where crossing a major version requires manual intervention.
  - `minor`: also restricts tags to the same minor version,
    only patch releases are considered updates.
  - `freeze`: the upstream registry is not checked at all; the container is
    always reported as frozen at its current version.

For `ImagePattern`, by convention, the sections are called `user` and `name`,
however, OCI image names don't make such distinction. It's all a repository to
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
  API endpoint (e.g. `docker.io`).
- Versions must be valid semantic version tags. Digest-pinned, textual tags or
  untagged images are treated as non-fatal errors.

[spec]: https://github.com/opencontainers/distribution-spec
