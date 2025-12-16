# Homebrew Publication Plan - Mac Dev Cleaner

> **Date:** 2025-12-15
> **Status:** Ready to Publish
> **Current Version:** MVP Complete (preparing v1.0.0)

---

## ✅ Pre-Publication Checklist

### Code Status
- ✅ MVP implementation complete
- ✅ TUI with NCDU-style navigation implemented
- ✅ All tests passing (`go test ./...`)
- ✅ Binary builds successfully (5.1MB)
- ✅ GoReleaser config ready (`.goreleaser.yaml`)
- ✅ GitHub Actions workflow ready (`.github/workflows/release.yml`)
- ✅ README.md with installation instructions
- ✅ LICENSE file (MIT)
- ✅ Makefile for build automation

### Pending Changes
- ⚠️ Uncommitted changes in `dev-mvp` branch:
  - `cmd/root/scan.go`
  - `internal/tui/tui.go`
  - `dev-cleaner` (binary)

---

## 📋 Publication Steps

### Phase 1: Prepare Code (10 mins)

**1.1 Commit Current Changes**
```bash
# Stage changes
git add cmd/root/scan.go internal/tui/tui.go

# Commit (don't commit binary)
git commit -m "feat: Finalize NCDU-style TUI with tree navigation

- Add interactive tree view with expandable nodes
- Implement keyboard shortcuts (arrows, space, enter, d, q)
- Add multi-select with visual checkboxes
- Integrate with existing scanner/cleaner modules"

# Merge to main
git checkout main
git merge dev-mvp
```

**1.2 Update Version in Code (Optional)**
If you have version in `cmd/root/root.go`:
```go
var Version = "1.0.0"
```

---

### Phase 2: Create GitHub Repositories (15 mins)

**2.1 Create Main Repository**
- Repository name: `dev-cleaner`
- Owner: `thanhdevapp`
- Description: "Clean development artifacts on macOS - Xcode, Gradle, node_modules"
- Visibility: Public
- License: MIT
- URL: `https://github.com/thanhdevapp/dev-cleaner`

**Steps:**
1. Go to https://github.com/new
2. Fill in repository details
3. Do NOT initialize with README (we have one)
4. Create repository

**2.2 Create Homebrew Tap Repository**
- Repository name: `homebrew-tools`
- Owner: `thanhdevapp`
- Description: "Homebrew tap for thanhdevapp's tools"
- Visibility: Public
- URL: `https://github.com/thanhdevapp/homebrew-tools`

**Steps:**
1. Go to https://github.com/new
2. Name MUST be `homebrew-tools` (or `homebrew-{tap-name}`)
3. Create empty repository
4. GoReleaser will auto-populate it

---

### Phase 3: Setup GitHub Secrets (5 mins)

**3.1 Generate Personal Access Token**

Go to: https://github.com/settings/tokens/new

**Settings:**
- Note: `GoReleaser Homebrew Tap`
- Expiration: No expiration (or 1 year)
- Scopes:
  - ✅ `repo` (full control)
  - ✅ `write:packages`
  - ✅ `workflow`

Click "Generate token" and **SAVE IT** (you won't see it again)

**3.2 Add Secret to Repository**

1. Go to: `https://github.com/thanhdevapp/dev-cleaner/settings/secrets/actions`
2. Click "New repository secret"
3. Name: `HOMEBREW_TAP_GITHUB_TOKEN`
4. Value: [paste token from 3.1]
5. Click "Add secret"

---

### Phase 4: Push to GitHub (5 mins)

**4.1 Add Remote**
```bash
git remote add origin https://github.com/thanhdevapp/dev-cleaner.git
```

**4.2 Push Code**
```bash
# Push main branch
git push -u origin main

# Push dev-mvp branch (optional)
git push origin dev-mvp
```

**4.3 Verify Push**
Visit: https://github.com/thanhdevapp/dev-cleaner

---

### Phase 5: Create Release Tag (2 mins)

**5.1 Create and Push Tag**
```bash
# Create annotated tag
git tag -a v1.0.0 -m "Release v1.0.0 - Mac Dev Cleaner MVP

Features:
- Scan Xcode, Android, Node.js artifacts
- NCDU-style TUI with tree navigation
- Interactive selection with keyboard shortcuts
- Safe deletion with confirmation
- Dry-run mode by default
- Comprehensive safety validation
"

# Push tag to GitHub
git push origin v1.0.0
```

**5.2 Monitor GitHub Actions**

1. Go to: https://github.com/thanhdevapp/dev-cleaner/actions
2. Watch "Release" workflow run
3. Should complete in ~3-5 minutes

**Expected Steps:**
- ✅ Checkout code
- ✅ Setup Go
- ✅ Run tests
- ✅ Run GoReleaser
  - Build binaries (darwin/linux, amd64/arm64)
  - Create archives
  - Create GitHub Release
  - Update Homebrew tap

---

### Phase 6: Verify Homebrew Formula (5 mins)

**6.1 Check Homebrew Tap Repository**

Visit: https://github.com/thanhdevapp/homebrew-tools

You should see:
- `Formula/dev-cleaner.rb` file created automatically by GoReleaser

**6.2 Verify Formula Content**

The formula should look like:
```ruby
class DevCleaner < Formula
  desc "Clean development artifacts on macOS - Xcode, Gradle, node_modules"
  homepage "https://github.com/thanhdevapp/dev-cleaner"
  url "https://github.com/thanhdevapp/dev-cleaner/releases/download/v1.0.0/dev-cleaner_1.0.0_darwin_arm64.tar.gz"
  sha256 "[checksum]"
  license "MIT"

  def install
    bin.install "dev-cleaner"
  end

  test do
    system "#{bin}/dev-cleaner", "--version"
  end
end
```

---

### Phase 7: Test Installation (5 mins)

**7.1 Test Homebrew Installation**
```bash
# Add tap
brew tap thanhdevapp/tools

# Install
brew install dev-cleaner

# Verify
dev-cleaner --version
dev-cleaner --help
```

**7.2 Test Functionality**
```bash
# Run scan
dev-cleaner scan --no-tui

# Test TUI
dev-cleaner scan
```

---

## 🎉 Publication Complete!

### User Installation Instructions

**Update README.md** to change "Coming Soon" to:

```markdown
## Installation

### Homebrew

\`\`\`bash
brew tap thanhdevapp/tools
brew install dev-cleaner
\`\`\`

### Verify Installation

\`\`\`bash
dev-cleaner --version
# Output: dev-cleaner version 1.0.0
\`\`\`
```

---

## 📊 Post-Publication Tasks

### Immediate (Same Day)

- [ ] Update README.md installation section
- [ ] Create GitHub Release notes with screenshots
- [ ] Test installation on fresh macOS system
- [ ] Share on Twitter/X, Reddit r/golang, Hacker News

### Within 1 Week

- [ ] Add GitHub repo to:
  - https://github.com/avelino/awesome-go
  - https://github.com/sindresorhus/awesome
  - https://github.com/toolleeo/cli-apps
- [ ] Create demo GIF/video for README
- [ ] Write blog post about development process
- [ ] Submit to Product Hunt

### Within 1 Month

- [ ] Gather user feedback
- [ ] Monitor GitHub issues
- [ ] Plan Phase 2 features (config file, progress bars)
- [ ] Implement analytics (optional, privacy-respecting)

---

## 🔄 Future Releases

### Versioning Strategy

Follow Semantic Versioning (semver):
- **Major (x.0.0)**: Breaking changes
- **Minor (1.x.0)**: New features, backwards-compatible
- **Patch (1.0.x)**: Bug fixes

### Release Process

For future releases:

```bash
# 1. Make changes, commit
git add .
git commit -m "feat: add new feature"

# 2. Create new tag
git tag -a v1.1.0 -m "Release v1.1.0 - Description"

# 3. Push
git push origin main
git push origin v1.1.0

# 4. GitHub Actions handles the rest!
```

---

## 🐛 Troubleshooting

### GoReleaser Fails

**Check logs:**
https://github.com/thanhdevapp/dev-cleaner/actions

**Common issues:**
- Missing `HOMEBREW_TAP_GITHUB_TOKEN` secret
- Token doesn't have correct permissions
- Tests failing
- Build errors

**Fix:**
```bash
# Test GoReleaser locally
goreleaser release --snapshot --clean

# Check build
go build -o dev-cleaner .
go test ./...
```

### Homebrew Install Fails

**Check formula:**
```bash
# View formula
brew cat thanhdevapp/tools/dev-cleaner

# Debug install
brew install --debug thanhdevapp/tools/dev-cleaner
```

**Common issues:**
- Wrong SHA256 checksum
- Binary not executable
- Missing dependencies

### Binary Size Too Large

Current: 5.1MB (acceptable)

If >10MB:
```bash
# Build with flags
go build -ldflags="-s -w" -o dev-cleaner .

# Further compress (optional)
upx dev-cleaner  # Reduces by ~60%
```

---

## 📞 Support

### Documentation
- Repository: https://github.com/thanhdevapp/dev-cleaner
- Issues: https://github.com/thanhdevapp/dev-cleaner/issues
- Homebrew Tap: https://github.com/thanhdevapp/homebrew-tools

### Community
- Create GitHub Discussions for Q&A
- Monitor issues for bug reports
- Accept pull requests for improvements

---

## ✅ Success Metrics

### Week 1 Goals
- [ ] 10+ GitHub stars
- [ ] 5+ successful installations
- [ ] 0 critical bugs reported

### Month 1 Goals
- [ ] 50+ GitHub stars
- [ ] 100+ installations
- [ ] Featured in one tech publication

### Long-term Goals
- [ ] 500+ stars
- [ ] 1000+ active users
- [ ] Community contributions

---

## 🎯 Next Steps

**Immediate Action Required:**

1. ✅ Review this plan
2. ⏳ Commit pending changes
3. ⏳ Create GitHub repositories
4. ⏳ Setup secrets
5. ⏳ Push code + tag
6. ⏳ Monitor release workflow
7. ⏳ Test installation
8. ⏳ Announce release

**Estimated Total Time:** 45-60 minutes

**Ready to publish? Let's go! 🚀**
