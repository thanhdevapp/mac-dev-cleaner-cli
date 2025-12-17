<!-- AUTO-GENERATED - DO NOT EDIT MANUALLY -->
<!-- This file is automatically updated by .github/workflows/update-docs.yml -->
<!-- Last updated: 2025-12-17 -->

# Installation Guide

## Quick Install

### macOS (Homebrew) - Recommended

```bash
# Add tap
brew tap thanhdevapp/tools

# Install
brew install dev-cleaner

# Verify
dev-cleaner --version
```

**Update to latest version:**
```bash
brew update
brew upgrade dev-cleaner
```

### One-Line Installer (macOS & Linux)

**Automatic installation:**
```bash
curl -fsSL https://raw.githubusercontent.com/thanhdevapp/mac-dev-cleaner-cli/dev-mvp/install.sh | bash
```

**What it does:**
- Detects your OS and architecture automatically
- Downloads the latest release binary
- Installs to `/usr/local/bin/`
- Verifies installation

**Manual review before running:**
```bash
# View the script first
curl -fsSL https://raw.githubusercontent.com/thanhdevapp/mac-dev-cleaner-cli/dev-mvp/install.sh

# Then run it
curl -fsSL https://raw.githubusercontent.com/thanhdevapp/mac-dev-cleaner-cli/dev-mvp/install.sh | bash
```

> **Note:** After migrating to `main` branch (see BRANCHING_STRATEGY.md), URLs will use `/main/` instead of `/dev-mvp/`

---

## Download Binaries

### Latest Release: v1.0.1

#### macOS

**Apple Silicon (M1/M2/M3)**
```bash
# Download
curl -L https://github.com/thanhdevapp/mac-dev-cleaner-cli/releases/download/v1.0.1/dev-cleaner_1.0.1_darwin_arm64.tar.gz -o dev-cleaner.tar.gz

# Extract
tar -xzf dev-cleaner.tar.gz

# Install to PATH
sudo mv dev-cleaner /usr/local/bin/

# Verify
dev-cleaner --version
```

**Intel**
```bash
# Download
curl -L https://github.com/thanhdevapp/mac-dev-cleaner-cli/releases/download/v1.0.1/dev-cleaner_1.0.1_darwin_amd64.tar.gz -o dev-cleaner.tar.gz

# Extract
tar -xzf dev-cleaner.tar.gz

# Install to PATH
sudo mv dev-cleaner /usr/local/bin/

# Verify
dev-cleaner --version
```

#### Linux

**ARM64**
```bash
curl -L https://github.com/thanhdevapp/mac-dev-cleaner-cli/releases/download/v1.0.1/dev-cleaner_1.0.1_linux_arm64.tar.gz -o dev-cleaner.tar.gz
tar -xzf dev-cleaner.tar.gz
sudo mv dev-cleaner /usr/local/bin/
dev-cleaner --version
```

**AMD64**
```bash
curl -L https://github.com/thanhdevapp/mac-dev-cleaner-cli/releases/download/v1.0.1/dev-cleaner_1.0.1_linux_amd64.tar.gz -o dev-cleaner.tar.gz
tar -xzf dev-cleaner.tar.gz
sudo mv dev-cleaner /usr/local/bin/
dev-cleaner --version
```

---

## Build from Source

### Prerequisites
- Go 1.21 or higher
- Git

### Steps

```bash
# Clone repository
git clone https://github.com/thanhdevapp/mac-dev-cleaner-cli.git
cd mac-dev-cleaner-cli

# Build CLI
go build -o dev-cleaner ./cmd/dev-cleaner

# Install to PATH
sudo mv dev-cleaner /usr/local/bin/

# Verify
dev-cleaner --version
```

### Build GUI (Wails)

```bash
# Install dependencies
cd frontend
npm install
npm run build
cd ..

# Build GUI app
wails build

# Run
./build/bin/Mac\ Dev\ Cleaner.app/Contents/MacOS/Mac\ Dev\ Cleaner
```

---

## Verify Installation

After installation, verify the tool works:

```bash
# Check version
dev-cleaner --version
# Expected output: dev-cleaner version 1.0.1

# Run help
dev-cleaner --help

# Test scan (dry-run)
dev-cleaner scan --help
```

---

## All Releases

View all releases: [GitHub Releases](https://github.com/thanhdevapp/mac-dev-cleaner-cli/releases)

---

## Uninstall

### Homebrew
```bash
brew uninstall dev-cleaner
brew untap thanhdevapp/tools
```

### Manual Install
```bash
sudo rm /usr/local/bin/dev-cleaner
rm ~/.dev-cleaner.log  # Optional: remove log file
```

---

## Troubleshooting

### "command not found: dev-cleaner"
- Check if `/usr/local/bin` is in your PATH
- Try: `echo $PATH | grep /usr/local/bin`
- Add to PATH if needed: `export PATH="/usr/local/bin:$PATH"`

### "permission denied"
- Run with sudo: `sudo mv dev-cleaner /usr/local/bin/`
- Or install to user directory: `mv dev-cleaner ~/bin/`

### Homebrew: "Xcode version outdated"
- This is a warning, not an error
- Installation still works
- Alternatively, use direct download method

---

## Support

- **Issues**: [GitHub Issues](https://github.com/thanhdevapp/mac-dev-cleaner-cli/issues)
- **Documentation**: [README.md](README.md)
- **Releases**: [GitHub Releases](https://github.com/thanhdevapp/mac-dev-cleaner-cli/releases)
