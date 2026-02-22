#!/bin/bash
set -e

# ============================================
# Bonk installer
# Usage: curl -fsSL https://raw.githubusercontent.com/vishrutdixit/bonk/main/install.sh | bash
# ============================================

VERSION="v0.1.0"
REPO="vishrutdixit/bonk"
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"

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

BINARY="bonk-${OS}-${ARCH}"
INSTALL_DIR="${HOME}/.local/bin"

echo "Installing bonk (${OS}/${ARCH})..."

# Create install directory
mkdir -p "$INSTALL_DIR"

# Download binary
curl -fsSL "${BASE_URL}/${BINARY}" -o "${INSTALL_DIR}/bonk"
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
