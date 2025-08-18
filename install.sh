#!/bin/bash

set -e

# gURL installer script

REPO="academic/gURL"
BINARY_NAME="gurl"

# Set install directory based on OS
if [[ "$OS" == "windows" ]]; then
    INSTALL_DIR="$HOME/bin"
    # Create directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"
else
    INSTALL_DIR="/usr/local/bin"
fi

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    armv7l) ARCH="arm" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux) OS="linux" ;;
    darwin) OS="darwin" ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

echo "Detecting platform: ${OS}/${ARCH}"

# Get latest release version
echo "Fetching latest release..."
LATEST_VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)

if [ -z "$LATEST_VERSION" ]; then
    echo "Error: Could not fetch latest version"
    exit 1
fi

echo "Latest version: $LATEST_VERSION"

# Construct download URL
BINARY_EXT=""
if [ "$OS" = "windows" ]; then
    BINARY_EXT=".exe"
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/${BINARY_NAME}_${OS}_${ARCH}${BINARY_EXT}"

# Check if we have a tarball instead
if ! curl -s --head "$DOWNLOAD_URL" | head -n 1 | grep -q "200 OK"; then
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/${BINARY_NAME}_${LATEST_VERSION#v}_${OS}_${ARCH}.tar.gz"
fi

echo "Downloading from: $DOWNLOAD_URL"

# Create temporary directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# Download the binary
FINAL_BINARY_NAME="$BINARY_NAME"
if [ "$OS" = "windows" ]; then
    FINAL_BINARY_NAME="${BINARY_NAME}.exe"
fi

if echo "$DOWNLOAD_URL" | grep -q "\.tar\.gz$"; then
    curl -sL "$DOWNLOAD_URL" | tar xz
    BINARY_PATH="$TMP_DIR/$FINAL_BINARY_NAME"
else
    curl -sL "$DOWNLOAD_URL" -o "$FINAL_BINARY_NAME"
    BINARY_PATH="$TMP_DIR/$FINAL_BINARY_NAME"
fi

# Check if binary exists
if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: Binary not found after download"
    exit 1
fi

# Make binary executable
chmod +x "$BINARY_PATH"

# Install binary
echo "Installing $FINAL_BINARY_NAME to $INSTALL_DIR..."

if [ "$OS" = "windows" ]; then
    # For Windows, just copy to user's bin directory
    cp "$BINARY_PATH" "$INSTALL_DIR/"
    echo "✅ $FINAL_BINARY_NAME installed to $INSTALL_DIR"
    echo "📝 Make sure $INSTALL_DIR is in your PATH environment variable"
elif [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_PATH" "$INSTALL_DIR/"
else
    echo "Need sudo privileges to install to $INSTALL_DIR"
    sudo mv "$BINARY_PATH" "$INSTALL_DIR/"
fi

# Cleanup
rm -rf "$TMP_DIR"

echo "✅ $BINARY_NAME installed successfully!"
echo "Run '$BINARY_NAME --help' to get started."

# Verify installation
if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    echo "✅ $BINARY_NAME is now available in your PATH"
    "$BINARY_NAME" --version 2>/dev/null || echo "Installation complete!"
else
    echo "⚠️  $BINARY_NAME installed but not found in PATH. You may need to restart your terminal or add $INSTALL_DIR to your PATH."
fi
