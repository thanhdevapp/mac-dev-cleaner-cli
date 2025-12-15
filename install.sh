#!/bin/bash
#
# Install script for dev-cleaner
# Usage: curl -fsSL https://raw.githubusercontent.com/thanhdevapp/mac-dev-cleaner-cli/dev-mvp/install.sh | bash
#

set -e

REPO="thanhdevapp/mac-dev-cleaner-cli"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="dev-cleaner"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}📦 Installing dev-cleaner...${NC}"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case $OS in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    *)
        echo -e "${RED}❌ Unsupported OS: $OS${NC}"
        echo "dev-cleaner currently supports macOS and Linux only."
        exit 1
        ;;
esac

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)
        echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

echo "   OS: $OS"
echo "   Arch: $ARCH"

# Get latest release version
echo -e "${YELLOW}🔍 Fetching latest version...${NC}"
LATEST=$(curl -sL https://api.github.com/repos/$REPO/releases/latest 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
    echo -e "${YELLOW}⚠️  Could not fetch latest release, using dev build...${NC}"
    # Fallback to dev build from artifacts
    echo -e "${RED}❌ No releases available yet. Please build from source:${NC}"
    echo "   git clone https://github.com/$REPO.git"
    echo "   cd mac-dev-cleaner-cli"
    echo "   go build -o dev-cleaner ."
    echo "   sudo mv dev-cleaner /usr/local/bin/"
    exit 1
fi

echo "   Version: $LATEST"

# Construct download URL
# Format: dev-cleaner_v1.0.0_darwin_arm64.tar.gz
VERSION_NUM=${LATEST#v}
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST/dev-cleaner_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"

echo -e "${YELLOW}📥 Downloading from: $DOWNLOAD_URL${NC}"

# Create temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download and extract
if ! curl -sL "$DOWNLOAD_URL" | tar xz -C "$TMP_DIR" 2>/dev/null; then
    echo -e "${RED}❌ Failed to download release${NC}"
    echo "URL: $DOWNLOAD_URL"
    exit 1
fi

# Check if binary exists
if [ ! -f "$TMP_DIR/$BINARY_NAME" ]; then
    echo -e "${RED}❌ Binary not found in archive${NC}"
    ls -la "$TMP_DIR"
    exit 1
fi

# Install binary
echo -e "${YELLOW}🔧 Installing to $INSTALL_DIR...${NC}"
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    echo "   (requires sudo)"
    sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

# Verify installation
if command -v $BINARY_NAME &> /dev/null; then
    echo ""
    echo -e "${GREEN}✅ Successfully installed dev-cleaner!${NC}"
    echo ""
    $BINARY_NAME --version
    echo ""
    echo "Run 'dev-cleaner scan' to get started"
else
    echo -e "${YELLOW}⚠️  Installed but not in PATH${NC}"
    echo "Add $INSTALL_DIR to your PATH or run: $INSTALL_DIR/$BINARY_NAME"
fi
