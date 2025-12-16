# UI Unit Testing Strategy - Mac Dev Cleaner GUI

**Date:** 2025-12-16
**Topic:** Viết unit test UI và chạy (Writing and running UI unit tests)
**Status:** Brainstorming Complete

---

## Problem Statement

Need to implement unit testing for React + TypeScript frontend in Wails v2 app. Currently no testing infrastructure exists - package.json has no test scripts, no testing libraries installed, no test files.

### Current State
- **Framework:** Wails v2 (Go backend) + React 18 + TypeScript + Vite
- **UI Library:** Radix UI + Custom shadcn/ui components
- **State Management:** Zustand
- **Styling:** Tailwind CSS
- **No testing setup:** Zero test files, no testing dependencies

### Requirements
1. Write unit tests for UI components
2. Run tests successfully
3. Maintain compatibility with Wails v2 architecture
4. Test components that call Go backend functions via Wails bindings

---

## Evaluated Approaches

### Approach 1: Vitest + React Testing Library (RECOMMENDED)

**Stack:**
- `vitest` - Fast, Vite-native test runner
- `@testing-library/react` - React component testing utilities
- `@testing-library/jest-dom` - Custom matchers
- `@testing-library/user-event` - User interaction simulation
- `jsdom` - DOM environment for Node.js

**Pros:**
- ✅ **Native Vite integration** - Zero config, uses existing vite.config.ts
- ✅ **Fast** - 10-100x faster than Jest due to Vite's speed
- ✅ **TypeScript support** - First-class TS support out of box
- ✅ **Compatible API** - Jest-like API, easy migration path
- ✅ **Watch mode** - Built-in, excellent DX
- ✅ **Coverage** - Built-in with v8 or istanbul
- ✅ **Industry standard** - React Testing Library is the de facto standard for React testing

**Cons:**
- ⚠️ Need to mock Wails runtime bindings (`../wailsjs/go/main/App`)
- ⚠️ Some Go-specific interactions can't be fully tested in isolation

**Recommended for:**
- Component logic testing
- UI rendering tests
- User interaction tests
- State management (Zustand) tests

---

### Approach 2: Jest + React Testing Library

**Stack:**
- `jest` - Traditional test runner
- `@testing-library/react` - Same as above
- `ts-jest` - TypeScript transformer for Jest

**Pros:**
- ✅ **Mature ecosystem** - Most documentation/examples available
- ✅ **Comprehensive mocking** - Powerful mock capabilities

**Cons:**
- ❌ **Slower** - Significantly slower than Vitest
- ❌ **Configuration overhead** - Requires babel/ts-jest configuration
- ❌ **Not Vite-native** - Requires separate config from build tooling
- ❌ **Bloated** - Larger dependency footprint

**Not recommended:** Vitest provides all benefits with better performance.

---

### Approach 3: Playwright Component Testing

**Stack:**
- `@playwright/experimental-ct-react` - Component testing

**Pros:**
- ✅ **Real browser** - Tests in actual browser environment
- ✅ **E2E-like** - Can test Wails bindings more realistically

**Cons:**
- ❌ **Overkill for unit tests** - Too heavy for component testing
- ❌ **Slower** - Browser startup overhead
- ❌ **Experimental** - Component testing still experimental
- ❌ **Complex setup** - Requires separate config

**Not recommended for unit tests:** Better suited for E2E testing.

---

## Final Recommendation: Vitest + React Testing Library

### Why This Wins
1. **Vite-native** - Reuses existing build config, zero friction
2. **Speed** - Fast test execution = better DX
3. **TypeScript** - No transform overhead, native TS support
4. **React Testing Library** - Industry standard, encourages good testing practices
5. **Simple setup** - Minimal config required

### Implementation Plan

#### 1. Install Dependencies

```bash
cd frontend
npm install -D vitest @testing-library/react @testing-library/jest-dom @testing-library/user-event jsdom
```

**Estimated time:** 2 minutes

---

#### 2. Configure Vitest

Create `frontend/vitest.config.ts`:

```typescript
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    css: true,
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
})
```

**Key points:**
- `globals: true` - No need to import `describe`, `it`, `expect`
- `environment: 'jsdom'` - DOM environment for React components
- `setupFiles` - Load test utilities and mocks
- Reuses same alias config as vite.config.ts

---

#### 3. Create Test Setup File

Create `frontend/src/test/setup.ts`:

```typescript
import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock Wails runtime
vi.mock('../../wailsjs/go/main/App', () => ({
  Scan: vi.fn(),
  GetScanResults: vi.fn(),
  GetSettings: vi.fn(),
  CleanItems: vi.fn(),
  SaveSettings: vi.fn(),
}))

vi.mock('../../wailsjs/go/models', () => ({
  types: {
    ScanOptions: vi.fn((opts) => opts),
  },
  services: {},
}))
```

**Purpose:**
- Import jest-dom matchers (toBeInTheDocument, etc.)
- Mock Wails Go function bindings
- Prevent runtime errors when components import Wails functions

---

#### 4. Add Test Scripts to package.json

```json
{
  "scripts": {
    "test": "vitest",
    "test:ui": "vitest --ui",
    "test:run": "vitest run",
    "coverage": "vitest run --coverage"
  }
}
```

**Commands:**
- `npm test` - Watch mode (recommended for development)
- `npm run test:run` - Single run (CI/CD)
- `npm run test:ui` - Visual UI for tests
- `npm run coverage` - Generate coverage report

---

#### 5. Write Sample Tests

**Example 1: Button Component Test**

Create `frontend/src/components/ui/button.test.tsx`:

```typescript
import { render, screen } from '@testing-library/react'
import { Button } from './button'
import userEvent from '@testing-library/user-event'

describe('Button', () => {
  it('renders children correctly', () => {
    render(<Button>Click me</Button>)
    expect(screen.getByText('Click me')).toBeInTheDocument()
  })

  it('calls onClick when clicked', async () => {
    const handleClick = vi.fn()
    render(<Button onClick={handleClick}>Click</Button>)

    await userEvent.click(screen.getByText('Click'))

    expect(handleClick).toHaveBeenCalledTimes(1)
  })

  it('is disabled when disabled prop is true', () => {
    render(<Button disabled>Disabled</Button>)
    expect(screen.getByRole('button')).toBeDisabled()
  })
})
```

---

**Example 2: Toolbar Component Test (with Wails mocks)**

Create `frontend/src/components/toolbar.test.tsx`:

```typescript
import { render, screen, waitFor } from '@testing-library/react'
import { Toolbar } from './toolbar'
import { Scan, GetSettings } from '../../wailsjs/go/main/App'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'

// Mock Zustand store
vi.mock('@/store/ui-store', () => ({
  useUIStore: () => ({
    viewMode: 'list',
    setViewMode: vi.fn(),
    toggleSettings: vi.fn(),
    searchQuery: '',
    setSearchQuery: vi.fn(),
    isScanning: false,
    setScanning: vi.fn(),
    scanResults: [],
    setScanResults: vi.fn(),
    selectedPaths: new Set(),
    selectAll: vi.fn(),
    clearSelection: vi.fn(),
  }),
}))

describe('Toolbar', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders scan button', () => {
    render(<Toolbar />)
    expect(screen.getByText('Scan')).toBeInTheDocument()
  })

  it('calls Scan when scan button is clicked', async () => {
    const mockGetSettings = vi.mocked(GetSettings)
    mockGetSettings.mockResolvedValue({ maxDepth: 5, autoScan: false, defaultView: 'list' })

    const mockScan = vi.mocked(Scan)
    mockScan.mockResolvedValue()

    render(<Toolbar />)

    await userEvent.click(screen.getByText('Scan'))

    await waitFor(() => {
      expect(mockScan).toHaveBeenCalled()
    })
  })

  it('renders view mode buttons', () => {
    render(<Toolbar />)
    expect(screen.getByTitle('List view')).toBeInTheDocument()
    expect(screen.getByTitle('Treemap view')).toBeInTheDocument()
    expect(screen.getByTitle('Split view')).toBeInTheDocument()
  })

  it('renders settings button', () => {
    render(<Toolbar />)
    expect(screen.getByTitle('Settings')).toBeInTheDocument()
  })
})
```

---

**Example 3: Zustand Store Test**

Create `frontend/src/store/ui-store.test.ts`:

```typescript
import { renderHook, act } from '@testing-library/react'
import { useUIStore } from './ui-store'

describe('useUIStore', () => {
  beforeEach(() => {
    // Reset store state
    useUIStore.setState({
      viewMode: 'list',
      selectedPaths: new Set(),
      scanResults: [],
    })
  })

  it('changes view mode', () => {
    const { result } = renderHook(() => useUIStore())

    act(() => {
      result.current.setViewMode('treemap')
    })

    expect(result.current.viewMode).toBe('treemap')
  })

  it('selects and deselects paths', () => {
    const { result } = renderHook(() => useUIStore())

    act(() => {
      result.current.toggleSelection('/path/to/file')
    })

    expect(result.current.selectedPaths.has('/path/to/file')).toBe(true)

    act(() => {
      result.current.toggleSelection('/path/to/file')
    })

    expect(result.current.selectedPaths.has('/path/to/file')).toBe(false)
  })

  it('selects all paths', () => {
    const { result } = renderHook(() => useUIStore())

    act(() => {
      result.current.selectAll(['/path1', '/path2', '/path3'])
    })

    expect(result.current.selectedPaths.size).toBe(3)
  })
})
```

---

#### 6. Run Tests

```bash
cd frontend

# Watch mode (recommended during development)
npm test

# Single run (CI/CD)
npm run test:run

# With coverage
npm run coverage
```

---

### Testing Strategy by Component Type

#### UI Components (shadcn/ui)
**Priority:** Medium
**Test focus:**
- Renders correctly
- Handles props
- User interactions (click, hover, etc.)
- Accessibility (ARIA attributes)

**Example:** button.test.tsx, input.test.tsx

---

#### Feature Components (Toolbar, Sidebar, ScanResults)
**Priority:** HIGH
**Test focus:**
- Integration with Zustand store
- Mocked Wails function calls
- Conditional rendering based on state
- User workflows (scan → select → clean)

**Example:** toolbar.test.tsx, scan-results.test.tsx

---

#### Store/State Management
**Priority:** HIGH
**Test focus:**
- State changes
- Action dispatchers
- Derived state/selectors

**Example:** ui-store.test.ts

---

#### Utils/Helpers
**Priority:** Medium
**Test focus:**
- Pure functions
- Edge cases
- Input validation

**Example:** utils.test.ts (formatBytes, etc.)

---

### Wails-Specific Testing Considerations

#### Challenge: Go Backend Bindings
Wails generates TypeScript bindings at `wailsjs/go/main/App.ts` that call Go functions. In unit tests, we can't actually call Go code.

**Solution: Mock Wails bindings**

```typescript
// setup.ts
vi.mock('../../wailsjs/go/main/App', () => ({
  Scan: vi.fn().mockResolvedValue(undefined),
  GetScanResults: vi.fn().mockResolvedValue([]),
  GetSettings: vi.fn().mockResolvedValue({ maxDepth: 5 }),
  CleanItems: vi.fn().mockResolvedValue(undefined),
}))
```

**In individual tests, override mocks:**

```typescript
it('handles scan failure', async () => {
  vi.mocked(Scan).mockRejectedValue(new Error('Permission denied'))

  render(<Toolbar />)
  await userEvent.click(screen.getByText('Scan'))

  await waitFor(() => {
    expect(screen.getByText('Scan Failed')).toBeInTheDocument()
  })
})
```

---

#### What You CAN'T Test (Unit Level)
- Actual Go function execution
- File system operations
- Real scan results from disk

**For these, you need:**
- **Integration tests** - Test Go backend separately
- **E2E tests** - Test full Wails app (Playwright/Cypress)

---

### Coverage Goals

**Initial target:** 60-70% coverage
**Components to prioritize:**
1. Toolbar (scan, clean, view switching)
2. ScanResults (rendering, selection)
3. UI Store (state management)
4. Utility functions (formatBytes, etc.)

**Components lower priority:**
- Theme provider (mostly pass-through)
- UI primitives (already tested by Radix UI)

---

### CI/CD Integration

Add to GitHub Actions:

```yaml
# .github/workflows/test-frontend.yml
name: Test Frontend

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - name: Install dependencies
        run: cd frontend && npm ci
      - name: Run tests
        run: cd frontend && npm run test:run
      - name: Generate coverage
        run: cd frontend && npm run coverage
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          directory: ./frontend/coverage
```

---

## Implementation Risks & Mitigations

### Risk 1: Wails Bindings Complexity
**Impact:** Medium
**Mitigation:**
- Centralized mock setup in `setup.ts`
- Document mock patterns
- Keep components loosely coupled from Wails API

### Risk 2: Test Maintenance Burden
**Impact:** Medium
**Mitigation:**
- Follow React Testing Library best practices (test behavior, not implementation)
- Use data-testid sparingly, prefer semantic queries
- Keep tests simple and focused

### Risk 3: False Sense of Security
**Impact:** High
**Mitigation:**
- Unit tests don't replace E2E tests
- Plan separate E2E testing for Go integration
- Document testing boundaries clearly

---

## Success Metrics

1. ✅ All tests pass on `npm run test:run`
2. ✅ Coverage reports generated successfully
3. ✅ No false positives (tests pass but bugs exist)
4. ✅ Fast feedback loop (< 5 seconds for watch mode)
5. ✅ CI/CD integration working

---

## Next Steps

### Phase 1: Setup (30 minutes)
1. Install dependencies
2. Configure vitest.config.ts
3. Create setup.ts with Wails mocks
4. Add test scripts to package.json
5. Verify test runner works

### Phase 2: Write Core Tests (2-3 hours)
1. Write tests for Toolbar component
2. Write tests for UI store
3. Write tests for utility functions
4. Verify coverage >= 60%

### Phase 3: Expand Coverage (Ongoing)
1. Add tests for remaining components
2. Refine mocks based on real usage
3. Document testing patterns
4. Set up CI/CD integration

---

## Alternative: Quick Start Script

If you want to automate setup, create `frontend/scripts/setup-tests.sh`:

```bash
#!/bin/bash
set -e

echo "📦 Installing test dependencies..."
npm install -D vitest @testing-library/react @testing-library/jest-dom @testing-library/user-event jsdom

echo "📝 Creating test config..."
cat > vitest.config.ts << 'EOF'
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    css: true,
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
})
EOF

echo "🔧 Creating test setup..."
mkdir -p src/test
cat > src/test/setup.ts << 'EOF'
import '@testing-library/jest-dom'
import { vi } from 'vitest'

vi.mock('../../wailsjs/go/main/App', () => ({
  Scan: vi.fn(),
  GetScanResults: vi.fn(),
  GetSettings: vi.fn(),
  CleanItems: vi.fn(),
  SaveSettings: vi.fn(),
}))

vi.mock('../../wailsjs/go/models', () => ({
  types: {
    ScanOptions: vi.fn((opts) => opts),
  },
  services: {},
}))
EOF

echo "✅ Test setup complete!"
echo "Run 'npm test' to start testing"
```

**Usage:** `cd frontend && bash scripts/setup-tests.sh`

---

## Questions & Clarifications Needed

### Before Implementation
1. **Coverage requirement?** - What minimum coverage % do you want?
2. **CI/CD platform?** - GitHub Actions, GitLab CI, other?
3. **Test scope?** - Unit tests only, or also integration/E2E?
4. **Priority components?** - Which components should be tested first?

### Unresolved Technical Questions
1. Should we test theme provider (dark/light mode switching)?
2. Do we need snapshot testing for UI components?
3. Should we mock all Wails bindings globally or per-test?
4. Coverage reporting - which format (HTML, LCOV, terminal)?

---

## Conclusion

**Recommended approach:** Vitest + React Testing Library provides the best balance of speed, simplicity, and compatibility with existing Vite setup.

**Key advantages:**
- Zero-friction setup (reuses Vite config)
- Fast test execution (critical for TDD)
- Industry-standard testing practices
- Easy mocking of Wails bindings

**Critical success factors:**
1. Proper Wails mock setup
2. Focus on behavior testing (not implementation)
3. Keep tests simple and maintainable
4. Don't over-mock (mock only boundaries)

**Timeline estimate:**
- Setup: 30 minutes
- First test suite: 2-3 hours
- Full coverage: 1-2 days (depending on component count)

---

**Status:** Ready for implementation decision
**Blocking issues:** None
**Dependencies:** None - can start immediately
