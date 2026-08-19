#!/bin/sh
# Set up the environment for workflows running in this repository
set -eu

readonly CI_DIRECTORY="./.ci"
readonly MISE_BINARY="${CI_DIRECTORY}/bin/mise"
readonly CHECKSUMS="${CI_DIRECTORY}/SHASUMS256.asc"

# Make sure all the directories exist:
mkdir -p "${CI_DIRECTORY}"
mkdir -p "$(dirname "${MISE_BINARY}")"
mkdir -p "$(dirname "${CHECKSUMS}")"

# Pull mise keys from keys.openpgp.org
gpg --keyserver hkps://keys.openpgp.org --recv-keys 24853EC9F655CE80B48E6C3A8B81C9D17413A06D
# Select and download the mise version
version=$(curl -fsSL https://api.github.com/repos/jdx/mise/releases/latest | grep -m1 '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
curl -fsSL -o "${MISE_BINARY}" "https://github.com/jdx/mise/releases/download/${version}/mise-${version}-linux-x64"
# Download checksums and ensure the downloaded binary matches
curl -fsSL -o "${CHECKSUMS}" "https://github.com/jdx/mise/releases/download/${version}/SHASUMS256.asc"
gpg --verify "${CHECKSUMS}"
grep " mise-${version}-linux-x64\$" "${CHECKSUMS}" | sha256sum -c -
# Prepare the binary for use
chmod +x "${MISE_BINARY}"
rm -f "${CHECKSUMS}"

# Pull all dependencies required by mise
"${MISE_BINARY}" install

# Make mise tools available to future steps
echo "$HOME/.local/share/mise/shims" >> "${FORGEJO_PATH}"
