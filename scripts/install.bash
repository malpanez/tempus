#!/bin/bash

set -euo pipefail

# Tempus installer script

REPO="malpanez/tempus"
VERSION="${1:-latest}"
INSTALL_DIR="${2:-/usr/local/bin}"
CURL=(curl --proto '=https' --tlsv1.2 -fsSL)
TMP_DIR=""

die() {
    echo "Error: $*" >&2
    exit 1
}

detect_os() {
    local OS_NAME
    OS_NAME=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$OS_NAME" in
        linux|darwin) echo "$OS_NAME" ;;
        *) die "unsupported operating system: $OS_NAME" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) die "unsupported architecture: $(uname -m)" ;;
    esac
}

resolve_version() {
    local REQUESTED="$1"
    local TAG
    if [[ "$REQUESTED" == "latest" ]]; then
        TAG=$("${CURL[@]}" -I -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest")
        TAG="${TAG##*/}"
    else
        TAG="v${REQUESTED#v}"
    fi
    [[ "$TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z.]+)?$ ]] || die "invalid release tag: $TAG"
    echo "$TAG"
}

verify_checksum() {
    local DIR="$1"
    local SUM_FILE="$2"
    if command -v sha256sum >/dev/null 2>&1; then
        (cd "$DIR" && sha256sum -c --quiet "$SUM_FILE")
    elif command -v shasum >/dev/null 2>&1; then
        (cd "$DIR" && shasum -a 256 -c --quiet "$SUM_FILE")
    else
        die "sha256sum or shasum is required to verify the download"
    fi
}

main() {
    local OS ARCH TAG ARCHIVE BINARY BASE_URL
    OS=$(detect_os)
    ARCH=$(detect_arch)
    TAG=$(resolve_version "$VERSION")
    ARCHIVE="tempus-${TAG}-${OS}-${ARCH}.tar.gz"
    BINARY="tempus-${OS}-${ARCH}"
    BASE_URL="https://github.com/${REPO}/releases/download/${TAG}"

    TMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TMP_DIR"' EXIT

    echo "Installing Tempus ${TAG} (${OS}/${ARCH})..."
    echo "Downloading ${BASE_URL}/${ARCHIVE}..."
    "${CURL[@]}" -o "${TMP_DIR}/${ARCHIVE}" "${BASE_URL}/${ARCHIVE}"
    "${CURL[@]}" -o "${TMP_DIR}/${ARCHIVE}.sha256" "${BASE_URL}/${ARCHIVE}.sha256"

    verify_checksum "$TMP_DIR" "${ARCHIVE}.sha256" || die "checksum verification failed for ${ARCHIVE}"
    tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "$TMP_DIR" "$BINARY"

    if [[ -w "$INSTALL_DIR" ]]; then
        install -m 0755 "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/tempus"
    else
        sudo install -m 0755 "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/tempus"
    fi

    echo "Tempus ${TAG} installed successfully to ${INSTALL_DIR}/tempus"
    echo "Run 'tempus --help' to get started!"
}

main
