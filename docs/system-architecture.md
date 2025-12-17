# System Architecture - Mac Dev Cleaner GUI

**Last Updated**: December 16, 2025
**Phase**: Wails GUI Phase 1
**Architecture Version**: 1.0.0

## Table of Contents
1. [System Overview](#system-overview)
2. [Architecture Patterns](#architecture-patterns)
3. [Component Architecture](#component-architecture)
4. [Communication Patterns](#communication-patterns)
5. [Service Architecture](#service-architecture)
6. [Data Layer](#data-layer)
7. [Event System](#event-system)
8. [Concurrency Model](#concurrency-model)
9. [Error Handling Architecture](#error-handling-architecture)
10. [Deployment Architecture](#deployment-architecture)

---

## System Overview

### Multi-Tier Architecture

```
┌────────────────────────────────────────────────────┐
│                 Presentation Layer                  │
│  React Components + Zustand State Management        │
│  (User Interface, Event Listeners)                  │
└────────────────┬─────────────────────────────────┘
                 │
         ┌───────┴──────────┐
         │ Wails IPC Bridge │ (RPC + Events)
         └───────┬──────────┘
                 │
┌────────────────▼─────────────────────────────────┐
│             Wails Application Layer                │
│  App.go: Service Registration & Lifecycle          │
│  Exposed Public Methods (RPC endpoints)            │
└────────────────┬─────────────────────────────────┘
                 │
┌────────────────▼─────────────────────────────────┐
│              Service Layer (Domain)                │
│  • ScanService      (Orchestration)                │
│  • TreeService      (Navigation + Caching)        │
│  • CleanService     (Deletion + Tracking)         │
│  • SettingsService  (Persistence)                 │
└────────────────┬─────────────────────────────────┘
                 │
┌────────────────▼─────────────────────────────────┐
│             Business Logic Layer                   │
│  • Scanner Package  (Scanning implementation)      │
│  • Cleaner Package  (Deletion + Safety checks)    │
│  • Types Package    (Domain models)                │
└────────────────────────────────────────────────────┘
```

### Key Characteristics
- **Separation of Concerns**: Clear boundaries between layers
- **Testability**: Each service independently testable
- **Scalability**: Event-driven allows async operations
- **Type Safety**: Full TypeScript on frontend, Go types on backend
- **Thread Safety**: Mutex-protected concurrent operations

---

## Architecture Patterns

### 1. Service Locator Pattern
**Location**: cmd/gui/app.go

The `App` struct acts as a service locator, managing all service instances:

```go
type App struct {
    scanService     *services.ScanService
    treeService     *services.TreeService
    cleanService    *services.CleanService
    settingsService *services.SettingsService
}

func (a *App) OnStartup(app *application.App) error {
    // Initialize all services
    a.scanService, _ = services.NewScanService(app)
    a.treeService, _ = services.NewTreeService(app)
    a.cleanService, _ = services.NewCleanService(app, false)
    a.settingsService = services.NewSettingsService()
}
```

**Benefits:**
- Centralized dependency management
- Wails lifecycle integration
- Single initialization point

### 2. Facade Pattern
**Location**: cmd/gui/app.go exposed methods

Public methods on `App` provide simplified interface to complex subsystems:

```go
// Simplified interface for frontend
func (a *App) Scan(opts types.ScanOptions) error
func (a *App) GetScanResults() []types.ScanResult
func (a *App) Clean(items []types.ScanResult) ([]cleaner.CleanResult, error)
```

**Benefits:**
- Frontend doesn't know about service internals
- Easy to add/modify services without changing RPC interface
- Clean API surface

### 3. Observer Pattern (Events)
**Location**: internal/services/*.go

Services emit events that frontend listens to:

```go
// In ScanService
s.app.Event.Emit("scan:started")
s.app.Event.Emit("scan:complete", results)

// In frontend
Events.On('scan:complete', (ev) => {
    setResults(ev.data as ScanResult[])
})
```

**Benefits:**
- Loose coupling between frontend and services
- Real-time progress updates
- Multiple listeners can subscribe

### 4. Repository/Data Access Pattern
**Location**: internal/services/settings_service.go

Services manage data persistence:

```go
type SettingsService struct {
    settings Settings
    path     string
    mu       sync.RWMutex
}

func (s *SettingsService) Save() error {
    // Atomically write to disk
}

func (s *SettingsService) Load() error {
    // Read from disk with defaults
}
```

**Benefits:**
- Isolated data access logic
- File I/O errors centralized
- Easy to add caching or network storage

### 5. Strategy Pattern
**Location**: internal/scanner/ (xcode.go, android.go, node.go, react_native.go)

Different scanning strategies for each artifact type:

```
Scanner interface {
    Scanner(opts ScanOptions) -> []ScanResult
}

Implementations:
- XcodeScanner (~/Library/Developer/Xcode/...)
- AndroidScanner (~/.gradle/..., ~/.android/...)
- NodeScanner (node_modules, npm cache, ...)
- ReactNativeScanner ($TMPDIR/metro-*, ...)
```

**Benefits:**
- Each scanner independently developed/tested
- New artifact types easily added
- Reusable across CLI and GUI

---

## Component Architecture

### Frontend Component Hierarchy

```
App
├── ThemeProvider
│   └── Layout (flex column)
│       ├── Toolbar
│       │   ├── Scan Button
│       │   ├── View Mode Buttons
│       │   ├── Search Input
│       │   └── Settings Button
│       ├── MainContent
│       │   └── ScanResults
│       │       ├── Loading State
│       │       ├── Empty State
│       │       └── Results List
│       │           ├── ResultItem
│       │           ├── ResultItem
│       │           └── ...
│       └── Toaster
│           └── Toast notifications
```

### Component Responsibilities

#### **App.tsx**
- Root component
- Theme provider setup
- Layout structure
- Event error boundary (future)

#### **Toolbar.tsx**
- User input: scan, view mode selection, search
- Local state: scanning flag
- Exports scan method calls to backend
- Uses UI store for view mode preference

#### **ScanResults.tsx**
- Displays scan results in current view mode
- Listens to all scan-related events
- Manages results and loading state
- Handles empty and loading states
- Calculates and displays statistics

#### **Theme Provider**
- System dark mode detection
- Theme switching capability
- Tailwind dark mode class management

#### **UI Components (Shadcn)**
- Fully accessible component library
- Consistent styling across application
- WCAG AA compliance built-in
- Customizable via Tailwind tokens

---

## Communication Patterns

### 1. RPC Calls (Frontend → Backend)

**Synchronous Method Calls:**
```typescript
// Frontend
const results = await GetScanResults()
await Scan(opts)
await Clean(items)
const settings = await GetSettings()
```

**Backend Wails Bindings (auto-generated):**
```go
// Exposed methods on App struct
func (a *App) GetScanResults() []types.ScanResult
func (a *App) Scan(opts types.ScanOptions) error
func (a *App) Clean(items []types.ScanResult) ([]cleaner.CleanResult, error)
func (a *App) GetSettings() services.Settings
```

**Type Safety:**
- Frontend has TypeScript types (auto-generated)
- Go types defined in pkg/types/
- Wails generates bindings for serialization

### 2. Event Broadcasting (Backend → Frontend)

**Event Emission Pattern:**
```go
// In services
s.app.Event.Emit("scan:started")
s.app.Event.Emit("scan:complete", results)
s.app.Event.Emit("scan:error", errMsg)
```

**Event Listening Pattern:**
```typescript
// In React components
useEffect(() => {
    const unsub = Events.On('scan:complete', (ev) => {
        setResults(ev.data)
    })

    return () => unsub()
}, [])
```

**Event Channels:**
```
scan:started      → No data
scan:complete     → []ScanResult
scan:error        → string

tree:updated      → *TreeNode
tree:cleared      → No data

clean:started     → int (item count)
clean:complete    → { results, freedSpace, successCount }
clean:error       → string
```

### 3. State Synchronization

**Frontend State:** Zustand store (local only)
```typescript
const UIStore = {
    viewMode: 'split',
    searchQuery: '',
    settingsOpen: false
}
```

**Backend State:** In-memory service structs (protected by mutex)
```go
type ScanService struct {
    results  []types.ScanResult  // Protected by mu
    scanning bool               // Protected by mu
    mu       sync.RWMutex
}
```

**Sync Strategy:**
- Frontend events trigger UI state updates
- Backend maintains authoritative state
- No sync protocol needed (stateless REST-like)
- Frontend loads state on mount via `GetScanResults()`, etc.

---

## Service Architecture

### ScanService

```go
type ScanService struct {
    app      *application.App      // Wails app reference
    scanner  *scanner.Scanner      // Core scanner
    results  []types.ScanResult    // Cached results
    scanning bool                  // In-progress flag
    mu       sync.RWMutex          // Thread safety
}
```

**Lifecycle:**
1. `NewScanService()` - Create, initialize scanner
2. `Scan(opts)` - Begin scan, emit start event, run scanner, sort, emit complete
3. `GetResults()` - Read-lock, return cached results
4. `IsScanning()` - Read-lock, return scanning flag

**Thread Model:**
- RWMutex allows concurrent reads
- Lock-free reads for `GetResults()`
- Exclusive lock during scan state change

**Error Handling:**
- Returns error if already scanning
- Emits `scan:error` event on scanner failure
- Continues despite partial errors

### TreeService

```go
type TreeService struct {
    app     *application.App
    scanner *scanner.Scanner
    cache   map[string]*types.TreeNode  // Path → Node
    mu      sync.RWMutex
}
```

**Caching Strategy:**
1. Check cache for existing node
2. If not found, perform directory scan
3. Store in cache for future access
4. Emit `tree:updated` event

**Cache Invalidation:**
- Manual via `ClearCache()`
- No TTL (cache persists session lifetime)
- Memory-efficient (nodes created on-demand)

### CleanService

```go
type CleanService struct {
    app     *application.App
    cleaner *cleaner.Cleaner      // Core deletion logic
    cleaning bool
    mu       sync.RWMutex
}
```

**Deletion Model:**
- Validates input (non-empty items)
- Prevents concurrent deletions
- Aggregates results (success/fail per item)
- Emits detailed completion event

**Progress Tracking:**
- Emits `clean:started` with count
- Calculates freed space
- Emits `clean:complete` with stats

### SettingsService

```go
type SettingsService struct {
    settings Settings
    path     string          // ~/.dev-cleaner-gui.json
    mu       sync.RWMutex
}
```

**Storage:**
- Single JSON file
- Atomic writes with lock
- Load on initialization with defaults
- Save on every update

**Settings Structure:**
- Theme preference (light/dark/auto)
- Default view mode (list/treemap/split)
- Auto-scan on launch
- Confirmation prompts
- Scan categories filter
- Tree depth limit

---

## Data Layer

### Domain Models

#### **ScanResult**
```go
type ScanResult struct {
    Type      string  // "xcode" | "android" | "node" | "react-native"
    Name      string  // Display name
    Path      string  // Absolute file path
    Size      int64   // Size in bytes
    FileCount int     // Number of files/directories
}
```

**Created by:** Scanner.ScanAll()
**Consumed by:** ScanResults component, Clean operation
**Storage:** In-memory (ScanService.results)

#### **ScanOptions**
```go
type ScanOptions struct {
    IncludeXcode       bool   // Scan ~/Library/Developer/Xcode/
    IncludeAndroid     bool   // Scan ~/.gradle/, ~/.android/
    IncludeNode        bool   // Scan node_modules, ~/.npm/
    IncludeReactNative bool   // Scan $TMPDIR react-native caches
    IncludeCache       bool   // Generic cache directories
    MaxDepth           int    // Directory traversal depth
    ProjectRoot        string // Starting directory
}
```

**Created by:** Frontend Toolbar
**Consumed by:** ScanService.Scan()

#### **TreeNode**
```go
type TreeNode struct {
    Path      string
    Name      string
    Size      int64
    FileCount int
    IsDir     bool
    Children  []*TreeNode
    Scanned   bool
}
```

**Created by:** TreeService.GetTreeNode()
**Used for:** Directory tree navigation
**Cached:** In TreeService.cache map

#### **Settings**
```go
type Settings struct {
    Theme          string   // "light" | "dark" | "auto"
    DefaultView    string   // "list" | "treemap" | "split"
    AutoScan       bool
    ConfirmDelete  bool
    ScanCategories []string // ["xcode", "android", "node"]
    MaxDepth       int
}
```

**Storage:** JSON file at ~/.dev-cleaner-gui.json
**Persistence:** Auto-load on startup, auto-save on update
**Defaults:** Hardcoded in SettingsService.Load()

### Type Serialization

**Go → TypeScript (Wails Auto-generates):**
- All public struct fields become TypeScript interfaces
- JSON tags determine field names
- Nested types fully typed
- Output: `frontend/bindings/`

**Serialization Process:**
```
Go Type Definition (pkg/types/types.go)
    ↓
Wails Compiler (wails build)
    ↓
TypeScript Type Definition (frontend/bindings/...)
    ↓
Frontend components use typed API
```

---

## Event System

### Event Lifecycle

```
1. Backend emits: s.app.Event.Emit("event:name", data)
                              ↓
2. Wails IPC: Serializes data to JSON, sends to frontend
                              ↓
3. Frontend receives: Events.On("event:name", callback)
                              ↓
4. Component updates: setResults(ev.data)
                              ↓
5. React re-renders: <ScanResults results={results} />
```

### Event Categories

#### **Scan Events**
- `scan:started` - Emitted before scanner runs (0 data)
- `scan:complete` - Emitted after scan finishes with sorted results
- `scan:error` - Emitted if scan fails with error message

#### **Tree Events**
- `tree:updated` - Emitted after directory scan with TreeNode
- `tree:cleared` - Emitted after cache clear

#### **Clean Events**
- `clean:started` - Emitted with item count
- `clean:complete` - Emitted with aggregated results
- `clean:error` - Emitted if clean operation fails

### Frontend Event Handling

```typescript
// Pattern 1: Use effect listener
useEffect(() => {
    const unsub = Events.On('scan:complete', (ev) => {
        setResults(ev.data as ScanResult[])
    })
    return () => unsub()
}, [])

// Pattern 2: In component methods
const handleScan = async () => {
    try {
        await Scan(opts)
        // Wait for scan:complete event via listener
    } catch (error) {
        // Handle RPC error
    }
}
```

### Error Propagation

```
Errors at multiple layers:
1. RPC Error: Catch in Wails method call
   → throw error() → catch in Frontend

2. Event Error: Listen to event:error channel
   → s.app.Event.Emit("scan:error", msg)
   → Events.On("scan:error", handler)

3. Silent Failure: Check IsScanning() flag
   → Prevents UI from appearing stuck
```

---

## Concurrency Model

### Backend Concurrency

**Multi-threaded Go:**
- Wails framework handles incoming RPC calls in goroutines
- Services use sync.RWMutex for state protection
- No explicit goroutine spawning (simple sequential model)

**Mutex Usage Pattern:**
```go
func (s *ScanService) Scan(opts types.ScanOptions) error {
    // Acquire exclusive lock
    s.mu.Lock()
    if s.scanning {
        s.mu.Unlock()
        return fmt.Errorf("scan already in progress")
    }
    s.scanning = true
    s.mu.Unlock()  // Release lock before expensive operation

    // Perform scan (not protected)
    results, err := s.scanner.ScanAll(opts)

    // Update state under lock
    s.mu.Lock()
    s.results = results
    s.scanning = false
    s.mu.Unlock()
}
```

**Read-Only Access:**
```go
func (s *ScanService) GetResults() []types.ScanResult {
    // Only read lock needed
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.results
}
```

**Benefits:**
- Multiple readers can access results simultaneously
- Writers (scan operations) are serialized
- No deadlocks (simple lock/unlock pattern)
- Prevents race conditions

### Frontend Concurrency

**React/JavaScript:**
- Single-threaded event loop
- No explicit concurrency management needed
- Async/await for Wails method calls

```typescript
const handleScan = async () => {
    setScanning(true)
    try {
        await Scan(opts)  // Await IPC bridge
    } finally {
        setScanning(false)
    }
}
```

**State Updates:**
- Queue-based (React batches updates)
- No race conditions due to single thread
- Event listeners update state safely

---

## Error Handling Architecture

### Error Handling Flow

```
Layer 1: Backend (Go)
    ├─ Validation errors (empty items, invalid options)
    ├─ I/O errors (permission denied, file not found)
    ├─ Scanner errors (unable to read directory)
    └─ Cleaner errors (unable to delete file)

Layer 2: Service (Go)
    ├─ Catch errors from scanner/cleaner
    ├─ Emit event:error if appropriate
    └─ Return error to caller

Layer 3: RPC/Wails
    ├─ Serialize error as RPC response
    ├─ Send to frontend
    └─ Throw as JavaScript Error

Layer 4: Frontend (React)
    ├─ Catch in try/catch block
    ├─ Show toast notification
    └─ Log to console for debugging
```

### Error Messages

**User-Facing Errors:**
- Toast notifications (auto-dismiss 4s)
- Simple, actionable messages
- Examples: "Scan Failed - Permission Denied", "No items to clean"

**Internal Errors:**
- Logged to console
- Includes full stack trace
- Helps developers debug issues

**File-System Errors:**
- Permission denied → Suggest system prompt
- Path validation → Prevent dangerous paths
- Locked files → Show partial results

### Graceful Degradation

**Partial Failures:**
```go
// CleanService continues despite failures
results := []CleanResult{}
for _, item := range items {
    if err := deleteItem(item); err != nil {
        results = append(results, CleanResult{
            Item: item,
            Success: false,
            Error: err.Error(),
        })
    }
}
// Returns partial results, not error
```

**Frontend Handling:**
```typescript
// Show which items failed
const failedItems = results.filter(r => !r.success)
const successCount = results.filter(r => r.success).length

// Display summary
toast({
    title: `Cleaned ${successCount} items`,
    description: `${failedItems.length} items failed`
})
```

---

## Deployment Architecture

### Development Build

```bash
# 1. Install dependencies
go mod download
cd frontend && npm install

# 2. Run in development
wails dev  # Hot reload enabled

# 3. Frontend bundled by Vite
# 4. Backend compiled by Go
# 5. Wails manages IPC bridge
```

### Production Build

```bash
# 1. Build frontend assets
cd frontend && npm run build  # Outputs to dist/

# 2. Build Wails application
wails build -o dev-cleaner

# 3. Platform-specific binary
# macOS: arm64 (Apple Silicon) or amd64 (Intel)
# Includes frontend bundle embedded
```

### Binary Packaging

**Wails Embedding:**
- Frontend assets embedded in Go binary
- No separate files needed
- Single executable distribution

**Cross-Platform Binaries:**
```bash
wails build -platform darwin/arm64  # macOS Apple Silicon
wails build -platform darwin/amd64  # macOS Intel
wails build -platform linux/arm64   # Linux ARM
wails build -platform linux/amd64   # Linux x86_64
```

### Configuration Files

**wails.json** - Framework configuration
- App name, description
- Frontend build configuration
- Platform-specific options
- Asset embedding settings

**vite.config.ts** - Build tool settings
- Hot Module Replacement (HMR)
- Asset optimization
- TypeScript configuration
- CSS preprocessing

**tailwind.config.js** - Style system
- Color theme customization
- Spacing scale
- Dark mode configuration
- Component plugins

---

## Performance Architecture

### Scanning Performance

**Optimization Strategies:**
1. Parallel directory traversal (Go built-in concurrency)
2. Results sorted once after completion (O(n log n))
3. No intermediate allocations (append capacity pre-allocated)

**Complexity:**
- Time: O(n log n) for sorting (n = number of found items)
- Space: O(n) for results array
- I/O: Sequential filesystem scans (can't parallelize effectively on macOS)

### Tree Navigation Performance

**Caching Strategy:**
- Cache directory nodes by path
- Avoid re-scanning same directory
- Memory footprint: One TreeNode per unique path

**Lookup Complexity:**
- Cache hit: O(1) map lookup
- Cache miss: O(k) where k = children per level

### UI Rendering Performance

**Current Approach:**
- Render all results in list view (no virtualization)
- Suitable for < 1000 items
- Room for improvement with virtual scrolling

**Future Optimization:**
- Virtual scroll window (only render visible items)
- Memoization of result items
- Debounced search filtering

---

## Monitoring & Observability

### Logging

**Backend Logging:**
- Error messages on stderr
- Operation start/end logged
- No structured logging (simple print statements)

**Frontend Logging:**
- Console.log for debugging
- Error stack traces in console
- Event system progress

### Events as Telemetry

**Progress Tracking:**
- scan:started → Operation began
- scan:complete → Success, include item count
- scan:error → Failure, include error message

**UI State:**
- Loading spinner during scan
- Result count display
- Freed space statistics

---

## Security Architecture

### Path Validation

**Cleaner Safety Checks:**
- Validates paths are within allowed directories
- Prevents deletion of system files
- Rejects symlinks pointing outside safe zone

**Blocked Paths:**
- System directories (/System, /Library system paths)
- User home directory (prevents `rm -rf ~/`)
- Application directories

### Permission Handling

**User Permissions:**
- Request OS permission prompts as needed
- Handle permission denied gracefully
- Show partial results if some items inaccessible

**File Locks:**
- Some files locked during use (open processes)
- Skip locked files (safe behavior)
- Retry not attempted

---

## Scalability Considerations

### Current Limits
- Effective for 10-100K+ artifacts per type
- UI can handle 1000+ items without virtualization
- Memory footprint: ~100KB per 1000 items (metadata only)

### Growth Path
1. **Short-term** (< 10K items): Current approach works fine
2. **Medium-term** (10K-100K): Add virtual scrolling, batch processing
3. **Long-term** (100K+): Implement incremental scanning, background worker threads

### Optimization Roadmap
- [ ] Virtual scrolling for large result sets
- [ ] Incremental scanning (don't rescan unchanged directories)
- [ ] Background worker threads for parallel scanning
- [ ] Streaming results (emit progress events)
- [ ] Database backend for scan history

---

**Document Version**: 1.0.0
**Last Updated**: December 16, 2025
**Applicable to**: Wails GUI Phase 1 and later
