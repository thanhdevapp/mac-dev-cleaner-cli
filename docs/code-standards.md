# Code Standards - Mac Dev Cleaner

**Last Updated**: December 16, 2025
**Version**: 1.0.0
**Scope**: Go backend, TypeScript/React frontend, build configuration

## Table of Contents
1. [General Principles](#general-principles)
2. [Go Code Standards](#go-code-standards)
3. [TypeScript/React Standards](#typescriptreact-standards)
4. [Frontend Component Patterns](#frontend-component-patterns)
5. [Testing Standards](#testing-standards)
6. [File Organization](#file-organization)
7. [Error Handling](#error-handling)
8. [Performance Standards](#performance-standards)
9. [Documentation Standards](#documentation-standards)
10. [Security Standards](#security-standards)

---

## General Principles

### YAGNI (You Aren't Gonna Need It)
- Don't implement features you can't demonstrate
- Don't add complexity for hypothetical use cases
- Keep codebase lean and maintainable

### KISS (Keep It Simple, Stupid)
- Prefer simple solutions over complex ones
- Easy to understand > Easy to write
- Optimize for readability first, performance second

### DRY (Don't Repeat Yourself)
- Extract common patterns into reusable functions
- Share logic across packages/components
- Use composition to reduce duplication

### Code Ownership
- Every file has clear ownership (package/module)
- Clear dependency direction (no circular imports)
- Avoid god classes/components

---

## Go Code Standards

### Package Organization

**Package Structure:**
```go
package services  // Clear, descriptive package name

// Types first
type ScanService struct {
    app     *application.App
    scanner *scanner.Scanner
    results []types.ScanResult
    mu      sync.RWMutex
}

// Constructors next
func NewScanService(app *application.App) (*ScanService, error) {
    // ...
}

// Public methods (alphabetical)
func (s *ScanService) GetResults() []types.ScanResult { }
func (s *ScanService) IsScanning() bool { }
func (s *ScanService) Scan(opts types.ScanOptions) error { }

// Private methods last
func (s *ScanService) validate(opts types.ScanOptions) error { }
```

**Naming Conventions:**
- Package names: lowercase, single word (services, scanner, cleaner)
- Type names: PascalCase (ScanService, ScanResult)
- Methods: PascalCase (GetResults, IsScanning)
- Variables: camelCase (scanService, results)
- Constants: UPPER_SNAKE_CASE (MAX_DEPTH, DEFAULT_TIMEOUT)
- Unexported: lowercase (scanService private is fine within type)

### Interface Design

**Minimal Interfaces:**
```go
// Good: Specific interface
type Scanner interface {
    ScanAll(opts ScanOptions) ([]ScanResult, error)
    ScanDirectory(path string, depth int) (*TreeNode, error)
}

// Bad: Too large
type FileOperations interface {
    Create() error
    Delete() error
    Read() error
    Write() error
    Move() error
    Copy() error
    // ... 10 more methods
}
```

**Accept Interfaces, Return Concrete Types:**
```go
// Good
func (a *App) Scan(opts types.ScanOptions) error {
    return a.scanService.Scan(opts)
}

// Not ideal
func NewScanService(scanner types.Scanner) *ScanService {
    // Don't require interfaces if concrete type works
}
```

### Concurrency Guidelines

**Mutex Usage:**
```go
type ScanService struct {
    results []types.ScanResult
    scanning bool
    mu sync.RWMutex  // Protect above fields
}

// Write operation
func (s *ScanService) updateResults(results []types.ScanResult) {
    s.mu.Lock()
    s.results = results
    s.mu.Unlock()
}

// Read operation - prefer read lock
func (s *ScanService) GetResults() []types.ScanResult {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.results
}

// Avoid this pattern
func (s *ScanService) BadPattern() {
    s.mu.Lock()
    defer s.mu.Unlock()
    // Holding lock during expensive operation
    s.ExpensiveOperation()
}
```

**Goroutine Guidelines:**
- Don't spawn goroutines in services unless necessary
- Wails framework handles concurrent RPC calls
- Use channels for inter-goroutine communication
- Always provide context for cancellation

### Error Handling

**Error Pattern:**
```go
// Return errors, don't ignore them
func (s *ScanService) Scan(opts types.ScanOptions) error {
    results, err := s.scanner.ScanAll(opts)
    if err != nil {
        s.app.Event.Emit("scan:error", err.Error())
        return err
    }
    return nil
}

// Wrap errors with context
if err != nil {
    return fmt.Errorf("scanning directory %s: %w", path, err)
}

// Custom error type when needed
type ScanError struct {
    Phase string  // "init", "scan", "sort"
    Cause error
}

func (e *ScanError) Error() string {
    return fmt.Sprintf("scan failed during %s: %v", e.Phase, e.Cause)
}
```

**Error Messages:**
- Start with lowercase (except package name)
- Include context (what was being done)
- Omit redundant prefixes (error already indicates failure)
- Don't include newlines

```go
// Good
return fmt.Errorf("scanning %s: %w", path, err)

// Bad
return fmt.Errorf("ERROR: Failed to scan the directory path: %s", path)
```

### Code Formatting

**Line Length:** Max 100 characters
**Indentation:** Tab (1 tab = standard indentation)
**Spacing:** Single blank line between logical sections

**Linting:**
```bash
golangci-lint run ./...
```

**Format Before Commit:**
```bash
go fmt ./...
```

### Comments

**Comment Style:**
```go
// Package scanner provides directory scanning functionality
// for detecting development artifacts.
package scanner

// ScanService coordinates scanning operations with progress events.
// Use NewScanService to create instances.
type ScanService struct {
    // ...
}

// Scan performs a full directory scan with the given options.
// Returns error if already scanning.
//
// Events:
//   - scan:started: No data
//   - scan:complete: []types.ScanResult
//   - scan:error: string
func (s *ScanService) Scan(opts types.ScanOptions) error {
    // Exported methods require comments
}

// internal detail: no comment needed if private and obvious
func (s *ScanService) validate(opts types.ScanOptions) error {
}
```

**Documentation Comments:**
- All exported types, functions, methods require comments
- Start with entity name (ScanService, Scan, Results)
- Explain what, why, any gotchas
- Include event names and data for observable operations

---

## TypeScript/React Standards

### File Organization

**Component File Structure:**
```typescript
// Imports (external, then internal)
import { useState, useEffect } from 'react'
import { Events } from '@wailsio/runtime'
import { GetScanResults } from '../../bindings/...'
import { formatBytes } from '@/lib/utils'

// Type definitions
interface Props {
  items: ScanResult[]
  onSelect?: (item: ScanResult) => void
}

// Component
export function ScanResults({ items, onSelect }: Props) {
  // Hooks
  const [loading, setLoading] = useState(false)

  // Effects
  useEffect(() => {
    // ...
  }, [])

  // Handlers
  const handleClick = (item: ScanResult) => {
    // ...
  }

  // Render
  return (
    <div>
      {items.map(item => (
        <div key={item.path} onClick={() => handleClick(item)}>
          {item.name}
        </div>
      ))}
    </div>
  )
}
```

### Naming Conventions

**Components:** PascalCase
```typescript
function ScanResults() { }
function Toolbar() { }
function Button() { }  // Even UI components
```

**Variables & Functions:** camelCase
```typescript
const scanResults = []
const handleScan = () => { }
const formatBytes = (bytes) => { }
```

**Constants:** UPPER_SNAKE_CASE
```typescript
const MAX_DEPTH = 5
const DEFAULT_THEME = 'auto'
const CACHE_TTL_MS = 3600000
```

**Types/Interfaces:** PascalCase
```typescript
interface ScanResult {
  type: string
  name: string
  path: string
  size: number
}

type ViewMode = 'list' | 'treemap' | 'split'
```

### Component Patterns

**Functional Components Only:**
```typescript
// Good
export function MyComponent() {
  const [state, setState] = useState()
  return <div>{state}</div>
}

// Avoid
class MyComponent extends React.Component {
  state = {}
  render() {
    return <div>{this.state}</div>
  }
}
```

**Event Listener Cleanup:**
```typescript
useEffect(() => {
  const unsub = Events.On('scan:complete', handler)

  // Always cleanup
  return () => {
    if (typeof unsub === 'function') unsub()
  }
}, [])
```

**State Management with Zustand:**
```typescript
// store/ui-store.ts
import { create } from 'zustand'

interface UIStore {
  viewMode: 'list' | 'treemap' | 'split'
  setViewMode: (mode: UIStore['viewMode']) => void

  searchQuery: string
  setSearchQuery: (query: string) => void
}

export const useUIStore = create<UIStore>((set) => ({
  viewMode: 'split',
  setViewMode: (mode) => set({ viewMode: mode }),

  searchQuery: '',
  setSearchQuery: (query) => set({ searchQuery: query }),
}))

// Usage
const { viewMode, setViewMode } = useUIStore()
```

**Conditional Rendering:**
```typescript
// Good: Explicit guard
if (loading) return <LoadingState />
if (results.length === 0) return <EmptyState />

// Good: Early return
if (!isVisible) return null

// Avoid: Ternary for complex logic
{isLoading ? <Spinner /> : results.length > 0 ? <List /> : <Empty />}
```

### Props & TypeScript

**Typed Props:**
```typescript
interface ToolbarProps {
  onScan: () => Promise<void>
  scanning: boolean
  viewMode: 'list' | 'treemap' | 'split'
  onViewModeChange: (mode: typeof viewMode) => void
}

export function Toolbar({
  onScan,
  scanning,
  viewMode,
  onViewModeChange,
}: ToolbarProps) {
  // ...
}
```

**Optional Props:**
```typescript
interface ButtonProps {
  label: string
  disabled?: boolean
  variant?: 'primary' | 'secondary' | 'ghost'
  onClick?: () => void
}

// Usage with defaults
<Button label="Click me" />
```

**Never Use `any`:**
```typescript
// Bad
function process(data: any) {
  return data.map(...)
}

// Good
function process(data: ScanResult[]) {
  return data.map(item => item.name)
}

// If truly unknown
function process(data: unknown) {
  if (Array.isArray(data)) {
    return data.map(...)
  }
}
```

### Code Style

**Formatting:**
- Use Prettier (auto-format on save)
- Max line length: 80 characters
- Use single quotes for strings (unless JSX requires double)
- Trailing commas in objects/arrays

**ESLint Rules:**
```javascript
// .eslintrc.json
{
  "extends": ["eslint:recommended"],
  "parser": "@typescript-eslint/parser",
  "rules": {
    "no-unused-vars": "error",
    "no-console": ["warn", { "allow": ["error", "warn"] }],
    "quotes": ["error", "single"],
    "semi": ["error", "never"]
  }
}
```

**Imports Organization:**
```typescript
// 1. React and framework imports
import { useState, useEffect } from 'react'
import { Events } from '@wailsio/runtime'

// 2. External dependencies
import { Button } from '@/components/ui/button'

// 3. Internal modules
import { useUIStore } from '@/store/ui-store'
import { formatBytes } from '@/lib/utils'

// 4. Types/Interfaces
import type { ScanResult } from '../../bindings/...'
```

---

## Frontend Component Patterns

### Container vs Presentation Components

**Container (Smart) Components:**
- Handle state management
- Fetch data from backend
- Listen to events
- Pass data to presentation components

```typescript
export function ScanResultsContainer() {
  const [results, setResults] = useState<ScanResult[]>([])

  useEffect(() => {
    Events.On('scan:complete', (ev) => {
      setResults(ev.data as ScanResult[])
    })
  }, [])

  return <ScanResultsPresentation results={results} />
}
```

**Presentation (Dumb) Components:**
- Receive all data via props
- Pure functions (same input → same output)
- No side effects
- Reusable across contexts

```typescript
interface ScanResultsPresentationProps {
  results: ScanResult[]
}

export function ScanResultsPresentation({ results }: ScanResultsPresentationProps) {
  return (
    <div>
      {results.map(item => (
        <div key={item.path}>{item.name}</div>
      ))}
    </div>
  )
}
```

### Custom Hooks

**Pattern:**
```typescript
// hooks/useScanResults.ts
import { useState, useEffect } from 'react'
import { Events } from '@wailsio/runtime'
import { GetScanResults } from '../../bindings/...'
import type { ScanResult } from '../../bindings/...'

export function useScanResults() {
  const [results, setResults] = useState<ScanResult[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    const unsub = Events.On('scan:complete', (ev) => {
      setResults(ev.data as ScanResult[])
      setLoading(false)
    })

    GetScanResults().then(setResults)

    return () => unsub()
  }, [])

  return { results, loading }
}

// Usage in component
function MyComponent() {
  const { results, loading } = useScanResults()
  // ...
}
```

### Error Boundaries (Future)

**Pattern for error handling:**
```typescript
class ErrorBoundary extends React.Component {
  state = { hasError: false, error: null }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error('Error caught:', error, info)
  }

  render() {
    if (this.state.hasError) {
      return <ErrorFallback error={this.state.error} />
    }
    return this.props.children
  }
}
```

---

## Testing Standards

### Go Testing

**Test File Naming:**
```
file.go  →  file_test.go
```

**Test Pattern:**
```go
package scanner

import "testing"

func TestScannerFindXcode(t *testing.T) {
    // Arrange
    scanner := NewScanner()

    // Act
    results, err := scanner.ScanAll(ScanOptions{
        IncludeXcode: true,
    })

    // Assert
    if err != nil {
        t.Fatalf("ScanAll() error = %v", err)
    }
    if len(results) == 0 {
        t.Error("ScanAll() expected results, got empty")
    }
}

func TestValidatePath(t *testing.T) {
    tests := []struct {
        name    string
        path    string
        wantErr bool
    }{
        {"valid path", "/tmp/test", false},
        {"system path", "/System/Library", true},
        {"empty path", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validatePath(tt.path)
            if (err != nil) != tt.wantErr {
                t.Errorf("validatePath(%q) error = %v, want %v", tt.path, err, tt.wantErr)
            }
        })
    }
}
```

**Coverage Goal:** 80%+ for business logic

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### React Testing (Future)

**Test File Location:**
```
components/Toolbar.tsx  →  components/Toolbar.test.tsx
```

**Testing Patterns:**
```typescript
import { render, screen, fireEvent } from '@testing-library/react'
import { Toolbar } from './Toolbar'

describe('Toolbar', () => {
  it('renders scan button', () => {
    render(<Toolbar />)
    expect(screen.getByRole('button', { name: /scan/i })).toBeInTheDocument()
  })

  it('disables scan button while scanning', () => {
    render(<Toolbar scanning={true} />)
    expect(screen.getByRole('button', { name: /scan/i })).toBeDisabled()
  })

  it('calls onScan when button clicked', () => {
    const onScan = jest.fn()
    render(<Toolbar onScan={onScan} />)
    fireEvent.click(screen.getByRole('button', { name: /scan/i }))
    expect(onScan).toHaveBeenCalled()
  })
})
```

---

## File Organization

### Maximum File Size

**Go Files:** 200 lines (excluding tests)
**React Components:** 150 lines (excluding types)

**When to Split:**
- Component > 150 lines: Extract sub-components
- Package > 5 files: Split by responsibility (services, models, utils)
- Test file > 200 lines: Split by feature/concern

### Directory Conventions

**Backend (Go):**
```
internal/
├── services/       # Application services (orchestration)
├── scanner/        # Domain logic (scanning)
├── cleaner/        # Domain logic (deletion)
└── types.go        # Shared types

cmd/
├── gui/            # Wails application
└── root/           # CLI commands

pkg/
└── types/          # Public types (exposed to frontend)
```

**Frontend:**
```
frontend/src/
├── components/     # React components
│   ├── ui/        # Shadcn UI components
│   ├── Toolbar.tsx
│   └── ...
├── hooks/         # Custom React hooks (future)
├── store/         # Zustand state
├── lib/           # Utilities (formatBytes, etc.)
└── types.ts       # Type definitions (Wails bindings)
```

---

## Error Handling

### Error Propagation Strategy

**Backend:**
```go
// Layer 1: Domain (Low level)
func deleteFile(path string) error {
    // Return concrete error
    return os.Remove(path)
}

// Layer 2: Service (Mid level)
func (c *CleanService) Clean(items []types.ScanResult) error {
    // Handle error, continue with others
    for _, item := range items {
        if err := deleteFile(item.Path); err != nil {
            // Log, emit event, but continue
            c.app.Event.Emit("clean:error", err.Error())
        }
    }
}

// Layer 3: App/RPC (High level)
func (a *App) Clean(items []types.ScanResult) error {
    // Return error to frontend
    return a.cleanService.Clean(items)
}
```

**Frontend:**
```typescript
const handleClean = async () => {
    try {
        const results = await Clean(items)
        // Process results
    } catch (error) {
        // Wails error (RPC failed)
        toast({
            variant: 'destructive',
            title: 'Clean Failed',
            description: error instanceof Error ? error.message : 'Unknown error'
        })
    }
}
```

### User-Facing Error Messages

**Do:**
- Keep simple and actionable
- Include what can be fixed by user
- Show relevant details (file path, size)

**Don't:**
- Stack traces to users
- Technical jargon (use "disk space" not "inode")
- Multiple technical details

**Example:**
```
Good:   "Unable to delete node_modules: Permission denied. Try running with sudo."
Bad:    "EACCES: permission denied, unlink '/Users/..../node_modules/...'"
```

---

## Performance Standards

### Go Performance Guidelines

**Benchmark Tests:**
```go
func BenchmarkScan(b *testing.B) {
    scanner := NewScanner()
    opts := ScanOptions{IncludeXcode: true}

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        scanner.ScanAll(opts)
    }
}

// Run: go test -bench=. -benchmem
```

**Memory Allocation:**
- Pre-allocate slices when size is known
- Avoid unnecessary copies
- Reuse buffers where possible

```go
// Good: Pre-allocate
results := make([]ScanResult, 0, len(items))

// Avoid: Repeated allocations
var results []ScanResult
for _, item := range items {
    results = append(results, result)  // Grows in steps
}
```

### React Performance Guidelines

**Memoization:**
```typescript
// Use React.memo for expensive components
export const ResultItem = React.memo(function ResultItem({
  item,
  onClick,
}: Props) {
  return <div onClick={onClick}>{item.name}</div>
})
```

**Key Selection:**
```typescript
// Good: Stable unique identifier
{results.map(item => (
  <ResultItem key={item.path} item={item} />
))}

// Bad: Array index changes on filter/sort
{results.map((item, index) => (
  <ResultItem key={index} item={item} />
))}
```

---

## Documentation Standards

### Code Comments

**Comment Quality:**
- Explain "why", not "what" (code shows what)
- Add context for non-obvious decisions
- Link to related code/issues

```go
// Good: Explains why
// Use RWMutex instead of Mutex for many reads with few writes
type ScanService struct {
    results []types.ScanResult
    mu      sync.RWMutex
}

// Bad: Just restates code
// Create a new scan service
func NewScanService() *ScanService {
}
```

### Function/Method Documentation

**Go:**
```go
// Scan performs a full directory scan with the given options.
// It acquires an exclusive lock preventing concurrent scans.
//
// Events:
//   - scan:started: No data
//   - scan:complete: []types.ScanResult
//   - scan:error: string
//
// Returns error if already scanning.
func (s *ScanService) Scan(opts types.ScanOptions) error {
}
```

**TypeScript:**
```typescript
/**
 * Formats bytes as human-readable size (B, KB, MB, GB)
 *
 * @param bytes - Size in bytes
 * @param decimals - Number of decimal places (default: 2)
 * @returns Formatted string (e.g., "1.23 MB")
 */
export function formatBytes(bytes: number, decimals = 2): string {
}
```

### README Standards

**Each Module Should Have:**
- Purpose statement (1 sentence)
- Key types/interfaces
- Usage examples
- Important implementation details

---

## Security Standards

### Input Validation

**Path Validation:**
```go
func (c *Cleaner) Clean(items []types.ScanResult) error {
    for _, item := range items {
        // Validate before deletion
        if !isAllowedPath(item.Path) {
            return fmt.Errorf("path not allowed: %s", item.Path)
        }

        if !isWithinSafeZone(item.Path) {
            return fmt.Errorf("path outside safe zone: %s", item.Path)
        }
    }
}
```

**Safe Paths:**
- User's home directory artifacts
- Project directories
- Known cache locations

**Blocked Paths:**
- /System, /Library (system paths)
- /Applications (installed apps)
- Symlinks outside safe zone

### Credential Handling

**Rules:**
- Never log secrets
- Don't commit credentials
- Use environment variables for sensitive config
- Mask sensitive data in error messages

```go
// Bad
log.Printf("Connecting to %s with password %s", server, password)

// Good
log.Printf("Connecting to %s", server)
// Store password in env var or secure store
```

---

## Development Workflow

### Before Commit

1. **Format Code**
   ```bash
   go fmt ./...
   cd frontend && npm run format && cd ..
   ```

2. **Run Tests**
   ```bash
   go test ./...
   cd frontend && npm test && cd ..
   ```

3. **Lint**
   ```bash
   golangci-lint run ./...
   cd frontend && npm run lint && cd ..
   ```

4. **Manual Testing**
   ```bash
   wails dev  # Run application
   # Test the feature
   ```

### Commit Messages

**Format:**
```
feat: Add scan progress events
fix: Prevent concurrent scan operations
docs: Update architecture documentation
refactor: Extract path validation logic
test: Add TreeService caching tests
```

**Body (if needed):**
- Explain why the change is needed
- Reference issue numbers
- Note any breaking changes

---

**Document Version**: 1.0.0
**Last Updated**: December 16, 2025
**Applicable to**: All new code in Wails GUI Phase 1+
