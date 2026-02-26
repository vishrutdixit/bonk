#!/usr/bin/env bash
set -euo pipefail

# Bonk installer
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/vishrutdixit/bonk/main/install.sh | bash
#   BONK_VERSION=v0.2.0 curl -fsSL https://raw.githubusercontent.com/vishrutdixit/bonk/main/install.sh | bash

REPO="vishrutdixit/bonk"
VERSION="${BONK_VERSION:-latest}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    darwin|linux) ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

INSTALL_DIR="${HOME}/.local/bin"
BINARY="bonk-${OS}-${ARCH}"
ARCHIVE="${BINARY}.tar.gz"

resolve_version() {
    if [[ "$VERSION" == "latest" ]]; then
        local latest_url
        latest_url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest")
        VERSION="${latest_url##*/}"
    elif [[ "$VERSION" != v* ]]; then
        VERSION="v${VERSION}"
    fi
}

resolve_version
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"

echo "Installing bonk ${VERSION} (${OS}/${ARCH})..."

# Create install directory
mkdir -p "$INSTALL_DIR"

# Download and extract release archive
tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

curl -fsSL "${BASE_URL}/${ARCHIVE}" -o "${tmpdir}/${ARCHIVE}"
tar -xzf "${tmpdir}/${ARCHIVE}" -C "$tmpdir"
if [[ ! -f "${tmpdir}/${BINARY}" ]]; then
    echo "Failed to locate ${BINARY} in ${ARCHIVE}"
    exit 1
fi

cp "${tmpdir}/${BINARY}" "${INSTALL_DIR}/bonk"
chmod +x "${INSTALL_DIR}/bonk"

echo "Installed to ${INSTALL_DIR}/bonk"

# Check if in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "Add this to your shell profile:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
fi

echo ""
echo "Run 'bonk' to start drilling!"
