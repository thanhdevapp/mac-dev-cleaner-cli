# Branching Strategy Recommendation

## Current Situation Analysis

### Repository State
- **dev-mvp**: Active development branch (origin/HEAD)
  - Latest: b7571b0 (v1.0.1 + CI/CD fixes)
  - All recent development happens here
  - Tags created from this branch

- **main**: Stale production branch
  - Latest: d84a7fb (stopped at v1.0.0 era)
  - ~20 commits behind dev-mvp
  - Not used for releases

**Problem**: Two "main" branches causing confusion

---

## Industry Best Practices (2025)

### 1. GitHub Flow (Recommended for This Project) ⭐

**What**: Simple, production-first workflow

```
main (production-ready)
  ↓
feature/xxx → PR → main → tag → release
  ↓
hotfix/xxx → PR → main → tag → patch release
```

**Characteristics**:
- Single source of truth: `main` = production
- Feature branches from main
- PR required for merge
- Tag from main for releases
- Deploy continuously from main

**Best For**:
- Small to medium teams ✅ (fits this project)
- Web applications ✅
- Single production version ✅
- Frequent releases ✅
- CI/CD workflows ✅

**Used By**: GitHub, GitLab, many modern SaaS

### 2. GitFlow (Legacy, Not Recommended)

**What**: Complex multi-branch workflow

```
main (production)
  ↑
release/v1.x
  ↑
develop (integration)
  ↑
feature/xxx
```

**Characteristics**:
- Multiple long-lived branches
- Separate develop and main
- Release branches for staging
- Complex merge workflow

**Best For**:
- Large enterprise teams
- Multiple production versions
- Scheduled releases (not continuous)

**Status**: Declining popularity, considered legacy

**Why Not**: Overkill for this project's needs

### 3. Trunk-Based Development

**What**: Single branch, very frequent merges

```
main
  ↑
short-lived feature branches (<1 day)
```

**Best For**: Very mature CI/CD, large teams with strong automation

**Why Not**: Requires extensive test automation

---

## Recommendation for This Project

### ✅ Adopt GitHub Flow with `main` as Production

**Rationale**:
1. ✅ Aligns with modern CI/CD practices
2. ✅ Simple, easy to understand
3. ✅ Standard across industry (GitHub default)
4. ✅ Matches project characteristics (small team, SaaS tool, frequent updates)
5. ✅ Reduces confusion (one clear production branch)

### Migration Plan

#### Option A: Merge dev-mvp to main, Use main Going Forward

**Steps**:
```bash
# 1. Update main with all dev-mvp changes
git checkout main
git merge dev-mvp --no-ff -m "merge: Sync dev-mvp into main for v1.0.1"
git push origin main

# 2. Update default branch on GitHub
# Settings → Branches → Default branch → main

# 3. Create releases from main going forward
git checkout main
git tag -a v1.0.2 -m "Release v1.0.2"
git push origin v1.0.2

# 4. Keep dev-mvp for ongoing development (optional)
# Or delete if switching fully to main
```

**Pros**:
- ✅ Standard GitHub workflow
- ✅ Clear semantics (main = production)
- ✅ Easier for new contributors

**Cons**:
- ⚠️ Requires updating workflows/docs
- ⚠️ Team needs to switch mental model

#### Option B: Rename dev-mvp to main, Archive old main

**Steps**:
```bash
# 1. Rename dev-mvp to main locally
git branch -m dev-mvp main

# 2. Delete old main on remote
git push origin --delete main

# 3. Push renamed branch
git push origin main

# 4. Set as default on GitHub
# Settings → Branches → Default branch → main

# 5. Archive old dev-mvp
git push origin --delete dev-mvp
```

**Pros**:
- ✅ No merge conflicts
- ✅ Clean history
- ✅ Immediate alignment with standard

**Cons**:
- ⚠️ Existing clones need to update
- ⚠️ May break existing PRs

#### Option C: Keep Current Setup (Not Recommended)

**Keep dev-mvp as de facto main**

**Pros**:
- ✅ No immediate work required
- ✅ No disruption

**Cons**:
- ❌ Confusing for contributors
- ❌ Non-standard workflow
- ❌ Two "main" branches
- ❌ `main` becomes dead weight

---

## Recommended Workflow (After Migration)

### Daily Development

```bash
# 1. Create feature branch from main
git checkout main
git pull origin main
git checkout -b feature/add-rust-scanner

# 2. Develop and commit
git add .
git commit -m "feat: Add Rust scanner support"

# 3. Push and create PR
git push origin feature/add-rust-scanner
gh pr create --base main

# 4. After approval, merge to main
# (Via GitHub UI with "Squash and merge")

# 5. Delete feature branch
git branch -d feature/add-rust-scanner
git push origin --delete feature/add-rust-scanner
```

### Release Process

```bash
# 1. Ensure main is ready
git checkout main
git pull origin main

# 2. Run final checks
go test ./...
go build -o dev-cleaner ./cmd/dev-cleaner

# 3. Create release tag
git tag -a v1.0.2 -m "Release v1.0.2: Description"
git push origin v1.0.2

# 4. Automation takes over
# - GitHub Actions builds binaries
# - Creates GitHub Release
# - Updates Homebrew formula
# - Updates documentation
```

### Hotfix Process

```bash
# 1. Create hotfix branch from main
git checkout main
git checkout -b hotfix/critical-bug

# 2. Fix and test
git commit -m "fix: Critical bug in scanner"

# 3. PR to main (expedited review)
gh pr create --base main

# 4. After merge, immediate release
git checkout main
git pull origin main
git tag -a v1.0.3 -m "Release v1.0.3: Hotfix for critical bug"
git push origin v1.0.3
```

---

## Branch Protection Rules (After Migration)

### For `main` branch

**Settings → Branches → Add rule → main**

**Require pull request before merging**:
- ✅ Require approvals: 1 (for team) or 0 (for solo)
- ✅ Dismiss stale reviews when new commits pushed

**Require status checks before merging**:
- ✅ Require branches to be up to date
- ✅ Status checks: CI (tests must pass)

**Do not allow bypassing the above settings** (optional for solo dev)

---

## Comparison: Before vs After

### Before (Current)

```
dev-mvp (origin/HEAD)     main (stale)
    ↓                         ↓
  Active                   v1.0.0 era
  v1.0.1                   Outdated
  Tag here                 Unused
```

**Issues**:
- Two main branches
- Confusing which is source of truth
- Non-standard setup

### After (GitHub Flow)

```
main (production)
  ↓
feature branches → PR → merge → tag → release
```

**Benefits**:
- ✅ One source of truth
- ✅ Standard workflow
- ✅ Clear semantics
- ✅ Industry standard

---

## Migration Checklist

### Pre-Migration
- [ ] Review all open PRs
- [ ] Notify team of upcoming change
- [ ] Backup current state (git bundle)

### Migration (Option A - Recommended)
- [ ] Merge dev-mvp into main
- [ ] Update default branch on GitHub → main
- [ ] Update workflow triggers (.github/workflows/*.yml)
- [ ] Update documentation (README.md, RELEASE_PROCESS.md)
- [ ] Test release process from main
- [ ] Communicate new workflow to team

### Post-Migration
- [ ] Archive dev-mvp branch (optional)
- [ ] Update branch protection rules
- [ ] Monitor first few releases
- [ ] Update CONTRIBUTING.md

---

## Updated Release Process (After Migration)

### Old Process (Current)
```bash
git checkout dev-mvp
git tag v1.0.2
git push origin v1.0.2
```

### New Process (After Migration)
```bash
git checkout main
git pull origin main
git tag -a v1.0.2 -m "Release v1.0.2"
git push origin v1.0.2
```

**Minimal change in practice!**

---

## Decision Matrix

| Criteria | GitHub Flow | GitFlow | Current Setup |
|----------|-------------|---------|---------------|
| Simplicity | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ |
| Industry Standard | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐ |
| CI/CD Support | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| Team Size Fit | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐ |
| Contributor Friendly | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐ |
| **Recommendation** | **✅ YES** | ❌ NO | ⚠️ MIGRATE |

---

## FAQ

### Q: Do I need to merge to main before tagging?
**A**: After migration, YES. Tag from `main` only.

### Q: Can I still use dev-mvp after migration?
**A**: Can keep for experimental features, but releases from `main` only.

### Q: What if I have uncommitted changes on dev-mvp?
**A**: Commit to dev-mvp, then merge to main via PR.

### Q: Will this break existing automation?
**A**: Need to update workflow triggers:
```yaml
# Old
on:
  push:
    branches: [dev-mvp]

# New
on:
  push:
    branches: [main]
```

### Q: What about the current v1.0.1 tag on dev-mvp?
**A**: Tags are repository-wide, not branch-specific. They'll work on main too.

### Q: How does this affect Homebrew users?
**A**: No impact. Automation works the same regardless of branch name.

---

## References

- [GitHub Flow Guide](https://www.alexhyett.com/git-flow-github-flow/)
- [GitFlow vs GitHub Flow](https://www.harness.io/blog/github-flow-vs-git-flow-whats-the-difference)
- [Git Branching Strategies](https://www.abtasty.com/blog/git-branching-strategies/)
- [Trunk-Based Development](https://www.flagship.io/git-branching-strategies/)

---

## Final Recommendation

**✅ Migrate to GitHub Flow with `main` as production branch**

**Timeline**:
1. **This week**: Merge dev-mvp → main (Option A)
2. **Next release**: Tag from main (test new workflow)
3. **Long term**: Archive dev-mvp, pure GitHub Flow

**Reason**: Aligns with 2025 best practices, simplifies workflow, standard across industry.

---

**Last Updated**: 2025-12-17
**Decision**: Pending team approval
