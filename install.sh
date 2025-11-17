#!/usr/bin/env bash
set -e

VERSION="0.0.2"
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then ARCH="amd64"; fi
if [ "$ARCH" = "aarch64" ]; then ARCH="arm64"; fi

URL="https://github.com/steverahardjo/GitAegis/releases/download/v${VERSION}/gitaegis-${OS}-${ARCH}"

echo "Downloading GitAegis..."
curl -L -o "/tmp/gitaegis" "$URL"

echo "Installing..."
sudo mv /tmp/gitaegis /usr/local/bin/gitaegis
sudo chmod +x /usr/local/bin/gitaegis

echo "GitAegis installed!"
