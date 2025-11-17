#!/usr/bin/env bash
set -e

VERSION="0.2"
URL="https://yourdomain.com/gitaegis-${VERSION}.deb"

echo "Downloading GitAegis..."
curl -L -o "/tmp/gitaegis.deb" "$URL"

echo "Installing..."
sudo apt install -y "/tmp/gitaegis.deb"

echo ":GitAegis installed!"