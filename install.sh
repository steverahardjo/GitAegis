#!/usr/bin/env bash
set -e

CLI_NAME="gitaegis"
BIN_DIR="/usr/local/bin"
URL="https://github.com/steverahardjo/GitAegis/releases/download/v2.0.0/gitaegis-linux-amd64"

echo "Installing $CLI_NAME..."

# Ensure /usr/local/bin is writable
if [ ! -w "$BIN_DIR" ]; then
    SUDO="sudo"
else
    SUDO=""
fi
# Download binary
$SUDO curl -L "$URL" -o "$BIN_DIR/$CLI_NAME"

# Make it executable
$SUDO chmod +x "$BIN_DIR/$CLI_NAME"

echo "Running initial setup..."
$SUDO "$BIN_DIR/$CLI_NAME" init --bash

echo "✅ $CLI_NAME installation complete! You can now run: $CLI_NAME"
