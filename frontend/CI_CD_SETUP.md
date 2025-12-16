# CI/CD Setup Documentation

This document explains the CI/CD pipeline configuration for the Mac Dev Cleaner project.

## Overview

The project uses **GitHub Actions** for continuous integration and deployment. The pipeline includes:

- ✅ Backend tests (Go)
- ✅ Frontend tests (React + TypeScript)
- ✅ Code coverage reporting
- ✅ Linting
- ✅ Automated releases

## CI Workflow (`.github/workflows/ci.yml`)

### Triggers

The CI workflow runs on:
- **Push** to branches: `main`, `dev-mvp`, `feat/wails-v2-migration`
- **Pull requests** to: `main`, `dev-mvp`

### Jobs

#### 1. Backend Tests
- **Runner:** Ubuntu Latest
- **Steps:**
  1. Checkout code
  2. Set up Go 1.21
  3. Build Go application
  4. Run Go tests

#### 2. Frontend Tests
- **Runner:** Ubuntu Latest
- **Working Directory:** `./frontend`
- **Steps:**
  1. Checkout code
  2. Set up Node.js 18 with npm cache
  3. Install dependencies (`npm ci`)
  4. TypeScript type checking (`npx tsc --noEmit`)
  5. Run tests (`npm run test:run`)
  6. Generate coverage report (`npm run test:coverage`)
  7. Upload coverage to Codecov
  8. Upload coverage artifacts (7-day retention)

#### 3. Go Lint
- **Runner:** Ubuntu Latest
- **Steps:**
  1. Checkout code
  2. Set up Go 1.21
  3. Run golangci-lint

## Coverage Configuration

### Vitest Coverage (`frontend/vitest.config.ts`)

```typescript
coverage: {
  provider: 'v8',
  reporter: ['text', 'json', 'html', 'lcov'],
  exclude: [
    'node_modules/',
    'src/test/',
    '**/*.test.{ts,tsx}',
    '**/*.spec.{ts,tsx}',
    'dist/',
    'vite.config.ts',
    'vitest.config.ts',
    'wailsjs/',
  ],
}
```

### Coverage Reports

Coverage is reported in multiple formats:
- **text** - Terminal output
- **json** - Machine-readable
- **html** - Visual HTML report (in `coverage/` directory)
- **lcov** - For Codecov integration

### Current Coverage Status

As of last run:
- **Overall:** 57.86% statements, 46.8% branches
- **Tested Components:** 100% coverage
  - `utils.ts` - 100%
  - `ui-store.ts` - 100%
  - `button.tsx` - 100%

**Coverage Gaps:**
- `clean-dialog.tsx` - 23% (needs tests)
- `toolbar.tsx` - 69% (partial coverage)
- Other components - 0% (not yet tested)

## Setup Requirements

### 1. Codecov Integration (Optional)

To enable coverage badges:

1. Sign up at [codecov.io](https://codecov.io)
2. Add your repository
3. Get your `CODECOV_TOKEN`
4. Add as GitHub Secret:
   - Go to: `Settings` → `Secrets and variables` → `Actions`
   - Create: `CODECOV_TOKEN` = `your-token`

5. Update README badge with your token:
   ```markdown
   [![codecov](https://img.shields.io/codecov/c/github/thanhdevapp/mac-dev-cleaner-cli?style=flat-square&token=YOUR_TOKEN)](https://codecov.io/gh/thanhdevapp/mac-dev-cleaner-cli)
   ```

### 2. Local Testing

Before pushing, test locally:

```bash
# Backend tests
go test ./...

# Frontend tests
cd frontend
npm run test:run

# Coverage
npm run test:coverage

# TypeScript check
npx tsc --noEmit
```

## CI/CD Best Practices

### ✅ Do's

1. **Run tests locally** before pushing
2. **Write tests** for new features
3. **Keep builds fast** (< 5 minutes)
4. **Fix failing tests** immediately
5. **Monitor coverage trends** (aim for 70%+)

### ❌ Don'ts

1. **Don't skip tests** with `[skip ci]`
2. **Don't commit failing tests**
3. **Don't decrease coverage** without justification
4. **Don't ignore CI failures**
5. **Don't push untested code**

## Troubleshooting

### CI Failure: "npm ci" fails

**Solution:** Update `package-lock.json`:
```bash
cd frontend
rm package-lock.json
npm install
git add package-lock.json
git commit -m "chore: update package-lock.json"
```

### CI Failure: Tests fail on CI but pass locally

**Common causes:**
1. **Environment differences** - Check Node.js version
2. **Missing dependencies** - Run `npm ci` locally
3. **Timezone issues** - Use UTC in tests
4. **Race conditions** - Fix non-deterministic tests

**Debug:**
```bash
# Run tests in CI-like environment
cd frontend
rm -rf node_modules
npm ci
npm run test:run
```

### Coverage Upload Fails

**Solution:**
1. Check `CODECOV_TOKEN` is set correctly
2. Verify coverage files exist: `ls frontend/coverage`
3. Check Codecov status: https://status.codecov.io

## Future Enhancements

### Planned Improvements

1. **Coverage Thresholds**
   - Uncomment in `vitest.config.ts`
   - Fail CI if coverage drops below 70%

2. **E2E Tests**
   - Add Playwright tests
   - Run in separate job

3. **Visual Regression Testing**
   - Add Percy or Chromatic
   - Test UI changes automatically

4. **Performance Monitoring**
   - Add Lighthouse CI
   - Track performance metrics

5. **Automated Security Scanning**
   - Add Snyk or Dependabot
   - Scan for vulnerabilities

## Monitoring

### GitHub Actions Dashboard

Monitor CI runs at:
```
https://github.com/thanhdevapp/mac-dev-cleaner-cli/actions
```

### Coverage Reports

View coverage at:
```
https://codecov.io/gh/thanhdevapp/mac-dev-cleaner-cli
```

### Status Badges

Badges in README show:
- ✅ CI passing/failing
- 📊 Code coverage percentage
- 🏷️ Latest release version

## Contact

For CI/CD issues, contact:
- **Email:** thanhdevapp@gmail.com
- **GitHub:** @thanhdevapp

---

**Last Updated:** 2025-12-16
**Maintained by:** Dev Team
