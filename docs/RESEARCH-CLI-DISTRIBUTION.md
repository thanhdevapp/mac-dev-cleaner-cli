# CLI Distribution & Packaging Research

> **Research Date:** 2025-12-15  
> **Purpose:** Phân tích các phương pháp đóng gói và phân phối CLI tool cho dev cleaner app

---

## 📋 Tổng Quan

Để phân phối một CLI tool, có 3 yếu tố chính cần xem xét:

1. **Packaging** - Cách đóng gói code thành executable
2. **Distribution** - Kênh phân phối (homebrew, npm, cargo, binary download)
3. **User Experience** - Yêu cầu từ phía người dùng để cài đặt

---

## 🔄 So Sánh Theo Ngôn Ngữ

### 1️⃣ Go

| Aspect                | Details                                       |
| --------------------- | --------------------------------------------- |
| **Build**             | `go build -o app` → Single binary             |
| **Cross-compile**     | Built-in: `GOOS=darwin GOARCH=arm64 go build` |
| **Release Tool**      | **GoReleaser** - tự động hóa toàn bộ process  |
| **User Requirements** | ❌ **Không cần cài Go runtime**                |

#### GoReleaser Workflow

```yaml
# .goreleaser.yaml
builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
      - windows
    goarch:
      - amd64
      - arm64

brews:
  - repository:
      owner: your-username
      name: homebrew-tap
    homepage: https://github.com/your-username/your-cli
```

**Installation cho User:**
```bash
# Option 1: Homebrew (macOS/Linux)
brew tap your-username/tap
brew install your-cli

# Option 2: Direct binary (all platforms)
curl -sL https://github.com/.../releases/download/v1.0.0/app_darwin_arm64.tar.gz | tar xz
sudo mv app /usr/local/bin/

# Option 3: Go install (yêu cầu Go)
go install github.com/your-username/your-cli@latest
```

---

### 2️⃣ Rust

| Aspect                | Details                                                |
| --------------------- | ------------------------------------------------------ |
| **Build**             | `cargo build --release` → Single binary                |
| **Cross-compile**     | Via `cross` or `cargo-zigbuild`                        |
| **Release Tool**      | **cargo-dist** hoặc **GoReleaser** (now supports Rust) |
| **User Requirements** | ❌ **Không cần cài Rust runtime**                       |

#### cargo-dist Workflow

```bash
# Init cargo-dist
cargo dist init

# Build for release
cargo dist build

# Generate CI for auto-release
cargo dist generate
```

**Installation cho User:**
```bash
# Option 1: Homebrew
brew install your-cli

# Option 2: Cargo binstall (không cần compile)
cargo binstall your-cli

# Option 3: Cargo install (compile từ source - cần Rust)
cargo install your-cli

# Option 4: Direct binary
curl -LsSf https://github.com/.../releases/download/v1.0.0/app-aarch64-apple-darwin.tar.gz | tar xz
```

---

### 3️⃣ Node.js / TypeScript

| Aspect                | Details                                          |
| --------------------- | ------------------------------------------------ |
| **Runtime Approach**  | `npm install -g your-cli` → Yêu cầu Node.js      |
| **Binary Approach**   | `pkg` hoặc `bun build --compile` → Single binary |
| **User Requirements** | Tùy thuộc vào phương pháp đóng gói               |

#### Phương pháp A: npm (yêu cầu Node.js)

```json
// package.json
{
  "name": "your-cli",
  "bin": {
    "your-cli": "./dist/cli.js"
  }
}
```

**Installation:**
```bash
npm install -g your-cli
# or
npx your-cli
```

#### Phương pháp B: Bun compile (binary - ✅ RECOMMENDED)

```bash
# Compile thành binary
bun build ./src/cli.ts --compile --outfile your-cli

# Cross-compile
bun build --compile --target=bun-darwin-arm64 ./src/cli.ts --outfile your-cli-macos
bun build --compile --target=bun-linux-x64 ./src/cli.ts --outfile your-cli-linux
bun build --compile --target=bun-windows-x64 ./src/cli.ts --outfile your-cli.exe
```

**User Installation:**
```bash
# Download binary - không cần Node.js/Bun
curl -sL https://github.com/.../releases/download/v1.0.0/your-cli-darwin-arm64 -o your-cli
chmod +x your-cli
sudo mv your-cli /usr/local/bin/
```

#### Phương pháp C: pkg (deprecated nhưng vẫn hoạt động)

```bash
npm install -g pkg
pkg . --targets node18-macos-arm64,node18-linux-x64,node18-win-x64
```

> ⚠️ **Lưu ý:** `pkg` đã deprecated từ Node.js 21. Khuyên dùng Bun thay thế.

---

## 📦 Phương Thức Phân Phối

### Homebrew (macOS/Linux)

**Yêu cầu setup:**
1. Tạo repo `homebrew-<tap-name>` trên GitHub
2. Tạo formula file `Formula/your-cli.rb`
3. Host binary trên GitHub Releases

**Formula Example:**
```ruby
# Formula/dev-cleaner.rb
class DevCleaner < Formula
  desc "Clean development project artifacts"
  homepage "https://github.com/username/dev-cleaner"
  url "https://github.com/username/dev-cleaner/releases/download/v1.0.0/dev-cleaner-darwin-arm64.tar.gz"
  sha256 "abc123..."
  version "1.0.0"
  
  def install
    bin.install "dev-cleaner"
  end
end
```

**User Experience:**
```bash
brew tap username/tap
brew install dev-cleaner
```

---

### GitHub Releases (Universal)

**Workflow:**
1. Tag version: `git tag v1.0.0`
2. Build binaries cho tất cả platforms
3. Upload lên GitHub Releases
4. User download và thêm vào PATH

**Tự động hóa với GitHub Actions:**
```yaml
name: Release
on:
  push:
    tags: ['v*']

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: goreleaser/goreleaser-action@v5
        with:
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

### npm Registry (cho Node.js tools)

```bash
npm publish
# User installs with:
npm install -g your-cli
```

---

### Cargo/crates.io (cho Rust tools)

```bash
cargo publish
# User installs with:
cargo install your-cli
```

---

## 👤 User Requirements Summary

| Distribution Method | User Needs to Install   | Difficulty |
| ------------------- | ----------------------- | ---------- |
| **Homebrew**        | Homebrew only           | ⭐ Easy     |
| **Direct Binary**   | Nothing (just download) | ⭐ Easy     |
| **npm global**      | Node.js                 | ⭐⭐ Medium  |
| **npx**             | Node.js                 | ⭐⭐ Medium  |
| **cargo install**   | Rust toolchain          | ⭐⭐⭐ Hard   |
| **go install**      | Go toolchain            | ⭐⭐⭐ Hard   |

---

## 🎯 Recommendation cho Mac Dev Cleaner

### Best Options (ranked):

#### 🥇 **Option 1: Go + GoReleaser + Homebrew**

**Pros:**
- Single binary, no runtime needed
- GoReleaser automates everything
- Easy Homebrew tap setup
- Fast builds, small binary size

**User Experience:**
```bash
brew tap thanhdevapp/tools
brew install dev-cleaner
dev-cleaner scan ~/Projects
```

---

#### 🥈 **Option 2: Rust + cargo-dist + Homebrew**

**Pros:**
- Best performance
- Memory safety
- Growing ecosystem

**User Experience:** Same as Go

---

#### 🥉 **Option 3: Bun/TypeScript + Binary**

**Pros:**
- Fastest development time
- TypeScript familiarity
- Bun's compile feature works well

**Cons:**
- Larger binary size (~70-100MB)
- Bun still maturing

---

## 📊 Quick Decision Matrix

```
┌─────────────────────────────────────────────────────────────┐
│                    CLI Tool Decision Tree                    │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Need fastest development?                                   │
│    └─ YES → TypeScript/Bun                                  │
│    └─ NO  ↓                                                 │
│                                                              │
│  Need best performance?                                      │
│    └─ YES → Rust                                            │
│    └─ NO  ↓                                                 │
│                                                              │
│  Want balance of speed + simplicity?                         │
│    └─ YES → Go ✅ (Recommended for this project)            │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔗 References

- [GoReleaser Documentation](https://goreleaser.com/)
- [cargo-dist Guide](https://opensource.axo.dev/cargo-dist/)
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Bun Build Documentation](https://bun.sh/docs/bundler/executables)
- [GitHub Actions - goreleaser-action](https://github.com/goreleaser/goreleaser-action)

---

## 📝 Next Steps

1. [ ] Chọn ngôn ngữ (Go recommended)
2. [ ] Setup project structure
3. [ ] Implement core scanning logic
4. [ ] Configure GoReleaser/cargo-dist
5. [ ] Create Homebrew tap
6. [ ] Setup GitHub Actions for automated releases
