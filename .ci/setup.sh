#!/bin/sh
# Set up the environment for workflows running in this repository
set -eu

readonly CI_DIRECTORY="./.ci"
readonly MISE_BINARY="${CI_DIRECTORY}/bin/mise"
readonly CHECKSUMS="${CI_DIRECTORY}/SHASUMS256.asc"
# Fingerprint of mise's release-signing key
# Cross-check https://mise.jdx.dev/gpg-key.asc if it ever needs to be rotated.
readonly MISE_GPG_KEY="24853EC9F655CE80B48E6C3A8B81C9D17413A06D"

# Make sure all the directories exist:
mkdir -p "${CI_DIRECTORY}"
mkdir -p "$(dirname "${MISE_BINARY}")"
mkdir -p "$(dirname "${CHECKSUMS}")"

# Use a throwaway GPG homedir instead of the user's real keyring. Cleaned up on exit.
GNUPGHOME=$(mktemp -d)
export GNUPGHOME
trap 'rm -rf "${GNUPGHOME}"' EXIT

# Select version and architecture to request
arch=$(uname -m)
case "${arch}" in
  # Map the runner's architecture to mise's release asset naming
  x86_64) mise_arch="x64" ;;
  aarch64) mise_arch="arm64" ;;
  *)
    echo "setup.sh: unsupported architecture '${arch}' (expected x86_64 or aarch64)" >&2
    exit 1
    ;;
esac
latest_release=$(curl -fsSL https://api.github.com/repos/jdx/mise/releases/latest)
version=$(printf '%s\n' "${latest_release}" | grep -m1 '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
mise_asset="mise-${version}-linux-${mise_arch}"

# Download mise
curl -fsSL -o "${MISE_BINARY}" "https://github.com/jdx/mise/releases/download/${version}/${mise_asset}"

# Pull mise keys from keys.openpgp.org
gpg --keyserver hkps://keys.openpgp.org --recv-keys "${MISE_GPG_KEY}"
curl -fsSL -o "${CHECKSUMS}" "https://github.com/jdx/mise/releases/download/${version}/SHASUMS256.asc"
# Download checksums and ensure the downloaded binary matches
gpg --verify "${CHECKSUMS}"
expected_sha=$(grep " ./${mise_asset}\$" "${CHECKSUMS}" | awk '{print $1}')
echo "${expected_sha}  ${MISE_BINARY}" | sha256sum -c -
# Prepare the binary for use
chmod +x "${MISE_BINARY}"
rm -f "${CHECKSUMS}"

# Unset GITHUB_TOKEN: it's actually a Forgejo Token for compat but it confuses tools:
unset GITHUB_TOKEN
# Pull all dependencies required by mise
"${MISE_BINARY}" install

# Make mise tools available to future steps
# Run ONLY if FORGEJO_PATH is set so it can still run locally
if [ -n "${FORGEJO_PATH:-}" ]; then
  echo "$HOME/.local/share/mise/shims" >> "${FORGEJO_PATH}"
fi
