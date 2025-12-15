#!/bin/bash

set -e

# Tempus installer script

VERSION="${1:-latest}"
INSTALL_DIR="${2:-/usr/local/bin}"

echo "Installing Tempus ${VERSION}..."

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Download URL
if [[ "$VERSION" = "latest" ]]; then
    DOWNLOAD_URL="https://github.com/malpanez/tempus/releases/latest/download/tempus-${OS}-${ARCH}"
else
    DOWNLOAD_URL="https://github.com/malpanez/tempus/releases/download/${VERSION}/tempus-${OS}-${ARCH}"
fi

# Create temporary file
TMP_FILE=$(mktemp)

echo "Downloading from ${DOWNLOAD_URL}..."
curl -L -o "$TMP_FILE" "$DOWNLOAD_URL"

# Make executable and move to install directory
chmod +x "$TMP_FILE"
sudo mv "$TMP_FILE" "${INSTALL_DIR}/tempus"

echo "Tempus installed successfully to ${INSTALL_DIR}/tempus"
echo "Run 'tempus --help' to get started!"