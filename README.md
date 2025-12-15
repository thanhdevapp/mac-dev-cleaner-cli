# Mac Dev Cleaner CLI

> 🧹 Clean development artifacts on macOS - free up disk space fast!

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

## Overview

Mac Dev Cleaner is a CLI tool that helps developers reclaim disk space by removing:

- **Xcode** - DerivedData, Archives, Caches
- **Android** - Gradle caches, SDK caches
- **Node.js** - node_modules, npm/yarn/pnpm/bun caches
- **Flutter/Dart** - .pub-cache, .dart_tool, build artifacts

## Installation

### Homebrew (Coming Soon)

```bash
brew tap thanhdevapp/tools
brew install dev-cleaner
```

### From Source

```bash
git clone https://github.com/thanhdevapp/dev-cleaner.git
cd dev-cleaner
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
dev-cleaner scan --flutter
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

### Flutter/Dart
- `~/.pub-cache/`
- `~/.dart_tool/`
- `~/Library/Caches/Flutter/`
- `~/Library/Caches/dart/`
- `*/build/` (in Flutter projects)
- `*/.dart_tool/` (in Flutter projects)
- `*/ios/build/`, `*/android/build/` (in Flutter projects)

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

- [x] MVP: Scan and clean commands
- [ ] TUI with interactive selection (BubbleTea)
- [ ] Config file support
- [ ] Homebrew distribution
- [ ] Progress bars

## License

MIT License - see [LICENSE](LICENSE) for details.
