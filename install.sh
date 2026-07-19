#!/usr/bin/env bash
# git-swap install script for macOS / Linux
# Usage: curl -sL https://raw.githubusercontent.com/jokot/git-swap/main/install.sh | bash

set -e

REPO="jokot/git-swap"
BIN_NAME="git-swap"

# Detect OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     OS="Linux";;
    Darwin*)    OS="Darwin";;
    *)          echo "Unsupported OS: ${OS}" && exit 1;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "${ARCH}" in
    x86_64)     ARCH="x86_64";;
    arm64)      ARCH="arm64";;
    aarch64)    ARCH="arm64";;
    *)          echo "Unsupported architecture: ${ARCH}" && exit 1;;
esac

# Fetch latest release URL
echo "Fetching latest release for ${OS}_${ARCH}..."
URL=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep "browser_download_url" | grep "${OS}_${ARCH}\.tar\.gz" | cut -d '"' -f 4)

if [ -z "$URL" ]; then
    echo "Could not find a release for ${OS}_${ARCH}"
    exit 1
fi

# Download and extract
TEMP_DIR=$(mktemp -d)
echo "Downloading from ${URL}..."
curl -sL "$URL" | tar -xz -C "$TEMP_DIR"

# Install binary
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    echo "Requires sudo to install to ${INSTALL_DIR}"
    sudo mv "${TEMP_DIR}/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
    sudo chmod +x "${INSTALL_DIR}/${BIN_NAME}"
else
    mv "${TEMP_DIR}/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
    chmod +x "${INSTALL_DIR}/${BIN_NAME}"
fi

rm -rf "$TEMP_DIR"
echo "✅ Successfully installed ${BIN_NAME} to ${INSTALL_DIR}"
echo "Run 'git-swap import' or 'git-swap add' to get started."