# Mac Dev Cleaner - Codebase Summary

**Last Updated**: December 16, 2025
**Version**: 1.0.0
**Phase**: Wails GUI Phase 1 Implementation

## Table of Contents
1. [Project Overview](#project-overview)
2. [Architecture](#architecture)
3. [Directory Structure](#directory-structure)
4. [Core Components](#core-components)
5. [Service Layer](#service-layer)
6. [Frontend Architecture](#frontend-architecture)
7. [Data Flow & Communication](#data-flow--communication)
8. [Key Technologies](#key-technologies)

---

## Project Overview

Mac Dev Cleaner is a cross-platform desktop application that helps developers reclaim disk space by identifying and safely removing development artifacts:

- **Xcode**: DerivedData, Archives, Caches
- **Android**: Gradle caches, SDK caches
- **Node.js**: node_modules, npm/yarn/pnpm/bun caches
- **React Native**: Metro bundler, Haste maps, packager caches

**Phase 1 (Wails GUI)**: Introduces a native desktop GUI using Wails v3 framework with React frontend and Go backend services.

---

## Architecture

### High-Level Design

```
┌─────────────────────────────────────────┐
│        Frontend (React + TypeScript)     │
│   (App.tsx, Components, UI Library)      │
└────────────┬────────────────────────────┘
             │ Events & RPC calls
             │ (Wails IPC Bridge)
┌────────────▼────────────────────────────┐
│       Wails v3 Application Core          │
│   (cmd/gui/main.go, cmd/gui/app.go)     │
└────────────┬────────────────────────────┘
             │ Service Method Calls
             │
┌────────────▼────────────────────────────┐
│         Service Layer (Go)               │
│  • ScanService                           │
│  • TreeService                           │
│  • CleanService                          │
│  • SettingsService                       │
└────────────┬────────────────────────────┘
             │
┌────────────▼────────────────────────────┐
│      Core Domain Logic (Go)              │
│  • Scanner (xcode, android, node, rn)   │
│  • Cleaner (deletion & safety)           │
│  • Types & Models                        │
└─────────────────────────────────────────┘
```

### Key Architectural Decisions

1. **Service Layer Pattern**: Isolates business logic from Wails application core
2. **Event-Driven Communication**: Frontend listens to progress events (scan:started, scan:complete, clean:complete)
3. **Separation of Concerns**: Clear boundaries between frontend UI, services, and domain logic
4. **Thread-Safe Services**: Mutex-protected concurrent operations (scanning, cleaning)
5. **Lazy-Loaded UI State**: Settings and tree nodes cached for performance

---

## Directory Structure

```
mac-dev-cleaner-cli/
├── cmd/
│   ├── gui/
│   │   ├── main.go              # Wails application entry point
│   │   └── app.go               # Wails App struct & exposed methods
│   └── root/
│       ├── root.go              # CLI root command (TUI mode)
│       ├── scan.go              # CLI scan subcommand
│       └── clean.go             # CLI clean subcommand
│
├── frontend/
│   ├── src/
│   │   ├── App.tsx              # Root React component
│   │   ├── main.tsx             # Frontend entry point
│   │   ├── index.css            # Global styles
│   │   ├── vite-env.d.ts        # Vite type definitions
│   │   ├── components/
│   │   │   ├── App.tsx          # Main layout
│   │   │   ├── toolbar.tsx      # Control toolbar (scan, view modes, search)
│   │   │   ├── scan-results.tsx # Results display & event listener
│   │   │   ├── theme-provider.tsx # Dark mode support
│   │   │   └── ui/
│   │   │       ├── button.tsx           # shadcn button
│   │   │       ├── checkbox.tsx         # shadcn checkbox
│   │   │       ├── dialog.tsx          # shadcn dialog
│   │   │       ├── input.tsx           # shadcn input
│   │   │       ├── label.tsx           # shadcn label
│   │   │       ├── select.tsx          # shadcn select
│   │   │       ├── switch.tsx          # shadcn switch
│   │   │       ├── toast.tsx           # shadcn toast
│   │   │       ├── toaster.tsx         # Toast container
│   │   │       └── use-toast.ts        # Toast hook
│   │   ├── lib/
│   │   │   └── utils.ts         # Utility functions (formatBytes, etc.)
│   │   └── store/
│   │       └── ui-store.ts      # Zustand UI state (viewMode, searchQuery)
│   ├── bindings/
│   │   └── github.com/
│   │       ├── wailsapp/wails/v3/
│   │       │   ├── internal/
│   │       │   │   ├── eventcreate.ts
│   │       │   │   └── eventdata.d.ts
│   │       │   └── pkg/application/
│   │       │       ├── index.ts
│   │       │       └── models.ts
│   │       └── thanhdevapp/dev-cleaner/
│   │           ├── cmd/gui/app.ts (generated)
│   │           └── pkg/types/models.ts (generated)
│   ├── public/
│   │   ├── index.html           # HTML entry point
│   │   ├── react.svg
│   │   ├── wails.png
│   │   ├── style.css
│   │   └── Inter-Medium.ttf
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts           # Vite build config
│   ├── tailwind.config.js       # Tailwind CSS config
│   └── postcss.config.js
│
├── internal/
│   ├── scanner/
│   │   ├── scanner.go           # Main scanner orchestrator
│   │   ├── scanner_test.go
│   │   ├── xcode.go             # Xcode artifacts scanning
│   │   ├── android.go           # Android artifacts scanning
│   │   ├── node.go              # Node.js artifacts scanning
│   │   ├── react_native.go      # React Native artifacts scanning
│   │   ├── react_native_test.go
│   │
│   ├── services/
│   │   ├── scan_service.go      # Scan operation coordinator
│   │   ├── tree_service.go      # Tree navigation with caching
│   │   ├── clean_service.go     # Clean operation coordinator
│   │   └── settings_service.go  # Settings persistence
│   │
│   ├── cleaner/
│   │   ├── cleaner.go           # Core cleanup logic
│   │   ├── safety.go            # Safety checks & validation
│   │   └── safety_test.go
│   │
│   ├── tui/
│   │   └── tui.go               # Legacy TUI implementation (Bubble Tea)
│   │
│   └── ui/
│       ├── formatter.go         # Display formatting (sizes, colors)
│       └── formatter_test.go
│
├── pkg/
│   └── types/
│       ├── types.go             # Domain models (ScanResult, ScanOptions)
│       └── tree.go              # Tree node structures
│
├── go.mod                        # Go module definition
├── go.sum                        # Go dependencies lock file
├── main.go                       # CLI entry point
├── Makefile
├── wails.json                    # Wails configuration
├── .repomixignore
├── .golangci.yml
├── .goreleaser.yaml
└── README.md
```

---

## Core Components

### 1. Frontend Structure

#### **App.tsx** - Root Component
- Wraps application with `ThemeProvider` (light/dark mode)
- Composes `Toolbar` and `ScanResults` components
- Renders `Toaster` for notifications

#### **Toolbar Component** (toolbar.tsx)
**Responsibilities:**
- Scan trigger button with loading state
- View mode toggle buttons (list/treemap/split)
- Search input field
- Settings toggle button

**State Management:**
- Uses Zustand store (`useUIStore`)
- Manages: viewMode, searchQuery, settingsOpen

**Event Handling:**
- Calls `Scan(opts)` with ScanOptions (includes all categories)
- Shows toast notifications on success/error

#### **ScanResults Component** (scan-results.tsx)
**Responsibilities:**
- Displays scan results in list format
- Shows loading/empty states
- Calculates and displays total disk space

**Event Listeners:**
- `scan:started` - Show loading spinner
- `scan:complete` - Update results list
- `scan:error` - Show error toast

**Integration Points:**
- Calls `GetScanResults()` on mount
- Listens to Wails events using `Events.On()`

#### **UI Components** (components/ui/)
Shadcn-based component library using Tailwind CSS:
- `button.tsx` - Primary, secondary, ghost, destructive variants
- `checkbox.tsx` - Multi-select support
- `dialog.tsx` - Modal dialogs
- `input.tsx` - Text input with search support
- `label.tsx` - Form labels
- `select.tsx` - Dropdown selection
- `switch.tsx` - Toggle switches
- `toast.tsx` + `toaster.tsx` - Notification system

### 2. Backend Structure

#### **Wails Application (cmd/gui/)**

**main.go**
- Creates Wails application instance
- Configures app metadata (name, description)
- Registers App service
- Sets macOS-specific options (close on last window)
- Starts application event loop

**app.go** - Wails Service
- Implements Wails service interface (`OnStartup`, `OnShutdown`)
- Exposes public methods callable from frontend
- Initializes all services on startup

**Exposed Methods Pattern:**
```go
// Each method is callable from frontend via Wails RPC
func (a *App) Scan(opts types.ScanOptions) error
func (a *App) GetScanResults() []types.ScanResult
func (a *App) IsScanning() bool
func (a *App) GetTreeNode(path string, depth int) (*types.TreeNode, error)
func (a *App) Clean(items []types.ScanResult) ([]cleaner.CleanResult, error)
func (a *App) GetSettings() services.Settings
func (a *App) UpdateSettings(settings services.Settings) error
```

---

## Service Layer

### **ScanService** (internal/services/scan_service.go)

**Purpose**: Orchestrates directory scanning with event emission

**Key Features:**
- Concurrent scan execution with state management
- Atomic result updates using RWMutex
- Event emission at lifecycle points (started, complete, error)
- Results sorted by size (largest first)

**Public API:**
```go
func (s *ScanService) Scan(opts types.ScanOptions) error
func (s *ScanService) GetResults() []types.ScanResult
func (s *ScanService) IsScanning() bool
```

**Events Emitted:**
- `scan:started` - No data
- `scan:complete` - Results []types.ScanResult
- `scan:error` - Error message string

**Thread Safety:**
- RWMutex protects: `scanning` flag, `results` array
- Lock/Unlock on state changes, RLock for reads

### **TreeService** (internal/services/tree_service.go)

**Purpose**: Provides lazy-loaded directory tree navigation with caching

**Key Features:**
- Caches scanned directory nodes by path
- Lazy evaluation (only scans requested paths)
- Limited depth traversal (default 5 levels)
- Thread-safe cache operations

**Public API:**
```go
func (t *TreeService) GetTreeNode(path string, depth int) (*types.TreeNode, error)
func (t *TreeService) ClearCache()
```

**Events Emitted:**
- `tree:updated` - Updated *types.TreeNode
- `tree:cleared` - No data

**Cache Strategy:**
- Check-before-scan to avoid redundant operations
- Atomic cache write after scan completion

### **CleanService** (internal/services/clean_service.go)

**Purpose**: Orchestrates file deletion with progress and error tracking

**Key Features:**
- Concurrent cleanup with state management
- Aggregates clean results with stats
- Emits detailed completion event with freed space
- Input validation (prevents empty clean)

**Public API:**
```go
func (c *CleanService) Clean(items []types.ScanResult) ([]cleaner.CleanResult, error)
func (c *CleanService) IsCleaning() bool
```

**Events Emitted:**
- `clean:started` - Item count int
- `clean:complete` - Map with results, freedSpace, successCount
- `clean:error` - Error message string

**Data Aggregation:**
- Tracks success/failure per item
- Calculates total freed space
- Counts successful deletions

### **SettingsService** (internal/services/settings_service.go)

**Purpose**: Manages persistent application preferences

**Settings Structure:**
```go
type Settings struct {
    Theme          string   // "light" | "dark" | "auto"
    DefaultView    string   // "list" | "treemap" | "split"
    AutoScan       bool
    ConfirmDelete  bool
    ScanCategories []string // ["xcode", "android", "node"]
    MaxDepth       int      // Tree traversal depth limit
}
```

**Storage:**
- File-based JSON: `~/.dev-cleaner-gui.json`
- Auto-loads on service creation
- Auto-saves on every update

**Public API:**
```go
func (s *SettingsService) Load() error
func (s *SettingsService) Save() error
func (s *SettingsService) Get() Settings
func (s *SettingsService) Update(settings Settings) error
```

---

## Frontend Architecture

### State Management (Zustand)

**UI Store** (store/ui-store.ts)
```typescript
interface UIStore {
  viewMode: 'list' | 'treemap' | 'split'
  setViewMode: (mode) => void

  searchQuery: string
  setSearchQuery: (query) => void

  settingsOpen: boolean
  toggleSettings: () => void
}
```

- Lightweight, persistence-ready
- Used by Toolbar and ScanResults components
- No server sync required (local UI state only)

### Event System

**Event Flow:**
1. Frontend calls Wails method (e.g., `Scan(opts)`)
2. Backend service emits events (e.g., `scan:started`)
3. Frontend listens via `Events.On()` (Wails runtime API)
4. Component state updates trigger re-render

**Event Channels:**
- `scan:*` - Scan lifecycle events
- `tree:*` - Tree navigation events
- `clean:*` - Cleanup lifecycle events

### Styling

**Technology Stack:**
- **Tailwind CSS** - Utility-first styling
- **Shadcn UI** - Pre-built accessible components
- **Dark Mode** - System preference auto-detection via ThemeProvider

**Design Tokens:**
- Color system follows design-guidelines.md
- Spacing: 8px base unit
- Border radius: 4px (small), 8px (medium), 12px (large)

---

## Data Flow & Communication

### Scan Operation Flow

```
User clicks "Scan" button
    ↓
Toolbar.handleScan() executes
    ↓
Frontend calls: Scan(ScanOptions)
    ↓
Wails IPC Bridge
    ↓
App.Scan() invokes ScanService.Scan()
    ↓
ScanService emits "scan:started" event
    ↓
Scanner.ScanAll() scans directories (blocking)
    ↓
Results sorted by size (descending)
    ↓
ScanService emits "scan:complete" with results
    ↓
Wails sends event to frontend
    ↓
ScanResults listener updates state
    ↓
React re-renders with results list
```

### Type Definitions

**ScanResult** (pkg/types/types.go)
```go
type ScanResult struct {
    Type      string  // "xcode", "android", "node", "react-native"
    Name      string  // Display name
    Path      string  // Absolute file path
    Size      int64   // Bytes
    FileCount int     // Number of files
    [additional fields]
}
```

**ScanOptions** (pkg/types/types.go)
```go
type ScanOptions struct {
    IncludeXcode        bool
    IncludeAndroid      bool
    IncludeNode         bool
    IncludeReactNative  bool
    IncludeCache        bool
    MaxDepth            int
    ProjectRoot         string
}
```

**TreeNode** (pkg/types/tree.go)
```go
type TreeNode struct {
    Path     string
    Name     string
    Size     int64
    FileCount int
    IsDir    bool
    Children []*TreeNode
    Scanned  bool
}
```

---

## Key Technologies

### Backend Stack
- **Go 1.21+** - Core application language
- **Wails v3** - Desktop framework (Windows/macOS/Linux)
- **Sync Package** - Thread-safe operations (RWMutex)

### Frontend Stack
- **React 18+** - UI framework
- **TypeScript** - Type-safe JavaScript
- **Vite** - Build tool (HMR, optimal bundling)
- **Tailwind CSS** - Utility-first CSS
- **Shadcn UI** - Component library
- **Zustand** - State management
- **Lucide React** - Icon library

### Generated Code
- **Wails Bindings** - Auto-generated TypeScript from Go types
- Located in: `frontend/bindings/`
- Provides type-safe RPC method calls

### Design System Reference
- See `docs/design-guidelines.md` for detailed specifications
- Component styling, spacing, colors, accessibility

---

## Build & Development

### Building the GUI Application

```bash
# Install dependencies
go mod download
cd frontend && npm install && cd ..

# Build Wails app (development)
wails dev

# Build release binary
wails build -o dev-cleaner -platform darwin/arm64
```

### Frontend Development

```bash
cd frontend
npm run dev      # Start Vite dev server with HMR
npm run build    # Production build
npm run lint     # Code linting
```

### Configuration Files

- **wails.json** - Wails framework configuration
- **tailwind.config.js** - Tailwind CSS customization
- **vite.config.ts** - Build tool configuration
- **tsconfig.json** - TypeScript compiler options

---

## Testing

### Unit Tests
- **internal/scanner/scanner_test.go** - Scanner tests
- **internal/scanner/react_native_test.go** - React Native scanner
- **internal/cleaner/safety_test.go** - Safety checks
- **internal/ui/formatter_test.go** - Formatter tests

### Test Execution

```bash
go test ./...           # Run all tests
go test -v ./...        # Verbose output
go test -cover ./...    # Coverage report
```

---

## Error Handling

### Service-Level Error Handling
1. **ScanService**: Emits `scan:error` event with error message
2. **CleanService**: Emits `clean:error` event, returns partial results
3. **TreeService**: Returns error tuple, frontend handles gracefully
4. **SettingsService**: Returns error on file I/O failures

### Frontend Error Handling
1. Try/catch around async Wails method calls
2. Toast notifications for user feedback
3. Graceful degradation (show empty states)
4. Detailed logging to browser console

---

## Performance Considerations

### Scanning Optimization
- Parallel directory traversal
- Results sorted in O(n log n) time
- Atomic result storage (no intermediate copies)

### Caching Strategy
- TreeService caches nodes by path
- Cache invalidation via `ClearTreeCache()`
- Lazy-loading prevents unnecessary scans

### UI Performance
- React component memoization
- Virtual scrolling ready (for future implementation)
- Event batching to avoid excessive re-renders

---

## Security & Safety

### Safety Measures
1. **Path Validation**: Cleaner validates paths before deletion
2. **Dry-Run Mode**: Default behavior (no-op deletion)
3. **Confirmation Dialog**: User confirms before destructive operations
4. **Logging**: All operations logged to file

### Permission Handling
- Graceful error messages for permission denied
- User prompted to grant access when needed
- Safe defaults for restricted paths

---

## Future Enhancements

### Phase 2 (Planned)
- Multi-select checkboxes with bulk operations
- Treemap visualization
- Progress bars during long operations
- Settings panel (theme, auto-scan, scan categories)
- Deep clean for React Native projects

### Phase 3 (Planned)
- Cross-platform support (Windows, Linux UI)
- Export reports (JSON/CSV)
- Scheduled background cleaning
- Advanced filtering and search
- Project-specific scanning

---

## Documentation References

- **System Architecture**: `docs/system-architecture.md`
- **Code Standards**: `docs/code-standards.md`
- **Design Guidelines**: `docs/design-guidelines.md`
- **Project Overview & PDR**: `docs/project-overview-pdr.md`
- **API Documentation**: Generated from Go code comments

---

**Last Updated**: December 16, 2025
**Generated by**: Wails GUI Phase 1 Documentation Update
**Repomix Version**: Used for codebase analysis
