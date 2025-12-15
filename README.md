# Mac Dev Cleaner CLI

> 🧹 Clean development artifacts on macOS - free up disk space fast!

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Release](https://img.shields.io/github/v/release/thanhdevapp/mac-dev-cleaner-cli?style=flat-square)](https://github.com/thanhdevapp/mac-dev-cleaner-cli/releases)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

## Overview

Mac Dev Cleaner is a CLI tool that helps developers reclaim disk space by removing:

- **Xcode** - DerivedData, Archives, Caches
- **Android** - Gradle caches, SDK caches
- **Node.js** - node_modules, npm/yarn/pnpm/bun caches
- **React Native** - Metro bundler, Haste maps, packager caches

## ✨ Features

- 🎯 **Smart Scanning** - Automatically detects development artifacts
- 🎨 **Interactive TUI** - NCDU-style tree navigation with keyboard shortcuts
- 🔒 **Safe by Default** - Dry-run mode prevents accidental deletions
- ✅ **Multi-select** - Choose exactly what to delete with checkboxes
- 🚀 **Fast & Efficient** - Scans thousands of directories in seconds
- 📦 **Single Binary** - No dependencies, just download and run
- 🌍 **Cross-platform** - Works on macOS and Linux (Intel & ARM64)

## 📸 Screenshot

![Mac Dev Cleaner TUI](screen1.png)

*Interactive TUI showing cleanable items with sizes and multi-select checkboxes*

## Installation

### Homebrew (Recommended)

```bash
brew tap thanhdevapp/tools
brew install dev-cleaner
```

### Direct Download

Download the latest release for your platform:

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/thanhdevapp/mac-dev-cleaner-cli/releases/download/v1.0.0/dev-cleaner_1.0.0_darwin_arm64.tar.gz | tar xz
sudo mv dev-cleaner /usr/local/bin/
```

**macOS (Intel):**
```bash
curl -L https://github.com/thanhdevapp/mac-dev-cleaner-cli/releases/download/v1.0.0/dev-cleaner_1.0.0_darwin_amd64.tar.gz | tar xz
sudo mv dev-cleaner /usr/local/bin/
```

**Linux (ARM64):**
```bash
curl -L https://github.com/thanhdevapp/mac-dev-cleaner-cli/releases/download/v1.0.0/dev-cleaner_1.0.0_linux_arm64.tar.gz | tar xz
sudo mv dev-cleaner /usr/local/bin/
```

**Linux (x86_64):**
```bash
curl -L https://github.com/thanhdevapp/mac-dev-cleaner-cli/releases/download/v1.0.0/dev-cleaner_1.0.0_linux_amd64.tar.gz | tar xz
sudo mv dev-cleaner /usr/local/bin/
```

### Verify Installation

```bash
dev-cleaner --version
# Output: dev-cleaner version 1.0.0
```

### From Source

```bash
git clone https://github.com/thanhdevapp/mac-dev-cleaner-cli.git
cd mac-dev-cleaner-cli
go build -o dev-cleaner .
sudo mv dev-cleaner /usr/local/bin/
```

## Usage

### Scan for Cleanable Items

```bash
# Scan all categories
dev-cleaner scan

# Scan specific category
dev-cleaner scan --ios
dev-cleaner scan --android
dev-cleaner scan --node
dev-cleaner scan --react-native  # or --rn

# Combine flags for React Native projects
dev-cleaner scan --rn --ios --android --node
```

**Example Output:**
```
🔍 Scanning for development artifacts...
──────────────────────────────────────────────────

  [1] android      9.0 GB  Android System Images
  [2] xcode        7.4 GB  Xcode DerivedData
  [3] android      5.4 GB  Gradle Caches
  [4] xcode        3.9 GB  DerivedData/Runner-xxx
  [5] node         1.8 GB  npm Cache
  ...
──────────────────────────────────────────────────
Total: 14 items, 31.4 GB
```

### Clean Items

```bash
# Interactive clean (dry-run by default)
dev-cleaner clean

# Actually delete files
dev-cleaner clean --confirm

# Clean specific category
dev-cleaner clean --ios --confirm
dev-cleaner clean --rn --confirm

# Clean React Native project (all RN-related caches)
dev-cleaner clean --rn --ios --android --node --confirm
```

### Safety Features

- ✅ **Dry-run by default** - preview before deleting
- ✅ **Confirmation required** - must type `yes` to delete
- ✅ **Path validation** - never touches system files
- ✅ **Logging** - all actions logged to `~/.dev-cleaner.log`

## Scanned Directories

### iOS/Xcode
- `~/Library/Developer/Xcode/DerivedData/`
- `~/Library/Developer/Xcode/Archives/`
- `~/Library/Caches/com.apple.dt.Xcode/`
- `~/Library/Developer/CoreSimulator/Caches/`
- `~/Library/Caches/CocoaPods/`

### Android
- `~/.gradle/caches/`
- `~/.gradle/wrapper/`
- `~/.android/cache/`
- `~/Library/Android/sdk/system-images/`

### Node.js
- `*/node_modules/` (in common project directories)
- `~/.npm/`
- `~/.pnpm-store/`
- `~/.yarn/cache/`
- `~/.bun/install/cache/`

### React Native
- `$TMPDIR/metro-*` - Metro bundler cache
- `$TMPDIR/haste-map-*` - Haste map cache
- `$TMPDIR/react-native-packager-cache-*` - RN packager cache
- `$TMPDIR/react-*` - React Native temp files

## Development

```bash
# Build
go build -o dev-cleaner .

# Run tests
go test ./...

# Run locally
./dev-cleaner scan
```

## Roadmap

### Completed ✅
- [x] MVP: Scan and clean commands
- [x] TUI with interactive selection (Bubble Tea)
- [x] NCDU-style tree navigation
- [x] Homebrew distribution
- [x] Cross-platform support (macOS, Linux)
- [x] Multi-platform binaries (Intel, ARM64)
- [x] React Native support (Metro, Haste, packager caches)

### Planned 🚀
- [ ] React Native project-specific builds (`--deep` flag)
- [ ] Config file support (~/.dev-cleaner.yaml)
- [ ] Progress bars for large operations
- [ ] Wails GUI (v2.0.0)
- [ ] Scheduled cleaning (cron integration)
- [ ] Export reports (JSON/CSV)

## License

MIT License - see [LICENSE](LICENSE) for details.
