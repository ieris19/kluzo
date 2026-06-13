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

## Limitations

- Only Quadlet `.container` files are supported.
- Upstream registries must expose
  the [OCI Distribution](https://github.com/opencontainers/distribution-spec)
  tag-listing API. Aliases may be needed for registries whose public hostname
  differs from their API endpoint (e.g. `docker.io`).
- Versions must be valid semantic version tags. Digest-pinned, textual tags or
  untagged images are skipped.
