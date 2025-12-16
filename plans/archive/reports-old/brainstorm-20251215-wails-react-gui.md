# Wails + React GUI Architecture - Brainstorm Report

**Date:** 2025-12-15
**Topic:** Add GUI desktop app using Wails v3 + React
**Status:** Architecture designed, ready for implementation

---

## Problem Statement

Current Mac Dev Cleaner is CLI-only:
```
CLI: dev-cleaner scan → TUI selection → Clean
```

**Limitations:**
- Terminal-only - not accessible for non-technical users
- No visual size representation (just numbers)
- Hard to distribute (requires Go install or binary download)
- Can't leverage native OS features (menubar, notifications, dock)

**Goal:** Add professional desktop GUI app while keeping CLI for power users.

---

## Requirements Summary

From user answers:

1. **Mode:** Dual (CLI + GUI coexist)
2. **Experience level:** Advanced Wails user
3. **Goals:** All - Better UX + Visual exploration + Easy distribution
4. **Timeline:** 1 month (polished v1, production-ready)
5. **Wails version:** v3 (alpha) - Multi-window, better bindings
6. **State management:** Hybrid - Go backend state + React UI state
7. **Visualization:** Tree list + Treemap combo
8. **Code structure:** Monorepo - Shared Go core

---

## Architecture Decisions

### Decision 1: Wails v3 (Alpha) ✅

**Rationale:**
- **Multi-window:** Separate windows for scan results, treemap, settings
- **Better bindings:** Static analyzer preserves types, comments, param names
- **Procedural API:** More flexible than v2's declarative approach
- **Production-ready:** Alpha but stable enough per docs, user is advanced

**Trade-offs:**
- ⚠️ Alpha instability risk
- ⚠️ Less documentation than v2
- ⚠️ Breaking changes possible before final release

**Mitigation:**
- Pin to specific v3 alpha version
- Follow v3 changelog closely
- Prepare for minor refactors if API changes

---

### Decision 2: Hybrid State Management ✅

**Architecture:**

```
┌─────────────────────────────────────────────────────────┐
│                    Go Backend (State)                    │
├─────────────────────────────────────────────────────────┤
│ - Scan results: []ScanResult                            │
│ - File tree: *TreeNode hierarchy                        │
│ - Settings: config, preferences                         │
│ - Operation status: scanning/cleaning/idle              │
│ - Disk usage metrics: sizes, counts                     │
│                                                          │
│ Events → React (one-way):                               │
│   - scan:progress                                       │
│   - scan:complete                                       │
│   - tree:updated                                        │
│   - operation:error                                     │
└─────────────────────────────────────────────────────────┘
                        ↓ Events
┌─────────────────────────────────────────────────────────┐
│                  React Frontend (View)                   │
├─────────────────────────────────────────────────────────┤
│ UI State (Zustand/Jotai):                               │
│ - selectedItems: Set<string> (paths)                    │
│ - expandedNodes: Set<string> (tree UI)                  │
│ - filters: { type, sizeRange, search }                  │
│ - viewMode: 'list' | 'treemap' | 'split'               │
│ - uiState: { loading, modal, sidebar }                  │
│                                                          │
│ Bindings ← Go (auto-generated):                         │
│   - Scan(), ScanDirectory()                             │
│   - Clean(), GetSettings()                              │
│   - GetTreeNode(), CalculateSize()                      │
└─────────────────────────────────────────────────────────┘
```

**Why Hybrid Wins:**

1. **Follow Wails best practice:** "State in Go, events to frontend" (official docs)
2. **Performance:** Large datasets (10K+ files) in Go, not React memory
3. **Single source of truth:** Go owns data, React renders
4. **DRY:** Reuse existing scanner logic without TypeScript port
5. **Type safety:** Go structs → TypeScript bindings auto-generated

**React State Limited To:**
- Selection state (user clicking checkboxes)
- UI expansion state (tree nodes expanded/collapsed)
- Filters and search (client-side, doesn't touch Go data)
- View preferences (list vs treemap)

---

### Decision 3: Tree List + Treemap Combo ✅

**Dual Visualization:**

```
┌──────────────────────────────────────────────────────────┐
│  Toolbar: [List] [Treemap] [Split] | Search | Filters   │
├──────────────────────────────────────────────────────────┤
│                                                          │
│  Split Mode (default):                                  │
│  ┌──────────────────┬───────────────────────────────┐   │
│  │   Tree List      │      Treemap Overview         │   │
│  │   (Navigation)   │      (Visual Size)            │   │
│  │                  │                               │   │
│  │  📁 Xcode        │  ┌─────────┬───────┐         │   │
│  │    7.4 GB        │  │ Xcode   │ Node  │         │   │
│  │  📁 Android      │  │ 7.4 GB  │ 1.8GB │         │   │
│  │    9.0 GB        │  ├─────────┴───────┤         │   │
│  │  📁 Node         │  │   Android       │         │   │
│  │    1.8 GB        │  │   9.0 GB        │         │   │
│  │                  │  └─────────────────┘         │   │
│  │  [Clean Selected]│  Click rect → drill down     │   │
│  └──────────────────┴───────────────────────────────┘   │
└──────────────────────────────────────────────────────────┘
```

**Implementation:**

**Tree List (Left):**
- React component: `<FileTreeList>`
- Virtual scrolling for 10K+ items (react-window)
- Lazy load children on expand
- Checkboxes for selection
- Drill-down navigation (like TUI plan)

**Treemap (Right):**
- Library: `recharts` treemap or custom D3
- Rectangles sized by disk usage
- Color by category (xcode=blue, android=green, node=brown)
- Click to drill down (syncs with tree list)
- Hover shows tooltip with size details

**Sync Behavior:**
- Select in tree → Highlight in treemap
- Click treemap rect → Expand tree node
- Bi-directional selection sync

**Why Combo:**
- Tree list for navigation (familiar UX)
- Treemap for visual overview (see space hogs instantly)
- Best of both: detailed + visual

---

### Decision 4: Monorepo Structure ✅

**Project Layout:**

```
mac-dev-cleaner-cli/          # Root monorepo
├── cmd/
│   ├── cli/                   # CLI entry point
│   │   └── main.go            # Current CLI app
│   └── gui/                   # NEW: GUI entry point
│       ├── main.go            # Wails v3 app
│       └── app.go             # App struct with bindings
│
├── internal/                  # Shared Go packages
│   ├── scanner/               # Existing scanner (reused)
│   │   ├── scanner.go
│   │   ├── xcode.go
│   │   ├── android.go
│   │   └── node.go
│   ├── cleaner/               # Existing cleaner (reused)
│   │   ├── cleaner.go
│   │   └── safety.go
│   ├── tui/                   # CLI TUI (bubbles)
│   │   └── tui.go
│   └── services/              # NEW: GUI backend services
│       ├── scan_service.go    # Scanning with events
│       ├── tree_service.go    # Tree navigation
│       ├── clean_service.go   # Cleaning operations
│       └── settings_service.go # Config management
│
├── frontend/                  # NEW: React app
│   ├── src/
│   │   ├── App.tsx            # Main app
│   │   ├── components/
│   │   │   ├── FileTreeList.tsx
│   │   │   ├── Treemap.tsx
│   │   │   ├── Toolbar.tsx
│   │   │   └── CleanDialog.tsx
│   │   ├── hooks/
│   │   │   ├── useScanResults.ts
│   │   │   ├── useTreeState.ts
│   │   │   └── useSelection.ts
│   │   ├── store/
│   │   │   └── uiStore.ts     # Zustand store
│   │   ├── types/
│   │   │   └── bindings.ts    # Auto-generated
│   │   └── lib/
│   │       └── wailsjs/       # Wails bindings
│   ├── package.json
│   ├── vite.config.ts
│   └── tsconfig.json
│
├── pkg/
│   └── types/                 # Shared types
│       ├── types.go           # Existing ScanResult, etc.
│       └── tree.go            # TreeNode (from TUI plan)
│
├── go.mod                     # Go dependencies
├── wails.json                 # Wails config
└── README.md
```

**Build Commands:**

```bash
# CLI (existing)
go build -o dev-cleaner ./cmd/cli

# GUI (new)
wails3 build                   # Production build
wails3 dev                     # Dev mode with hot reload

# Both
make build-all                 # Makefile target
```

**Why Monorepo:**
- ✅ Share scanner/cleaner logic (DRY)
- ✅ Single repo for issues, PRs, versioning
- ✅ Type consistency (Go types → TS bindings)
- ✅ Easier refactoring across CLI/GUI

---

## Go Backend Architecture

### Service Layer Pattern

**Why Services?**
- Encapsulate business logic
- Provide clean API for frontend bindings
- Handle events and progress updates
- Stateful operations (scanning, cleaning)

---

### Service 1: ScanService

**File:** `internal/services/scan_service.go`

```go
package services

import (
    "context"
    "github.com/thanhdevapp/dev-cleaner/internal/scanner"
    "github.com/thanhdevapp/dev-cleaner/pkg/types"
    "github.com/wailsapp/wails/v3/pkg/application"
)

type ScanService struct {
    app      *application.App
    scanner  *scanner.Scanner
    results  []types.ScanResult
    scanning bool
}

// NewScanService creates service
func NewScanService(app *application.App) *ScanService {
    s, _ := scanner.New()
    return &ScanService{
        app:     app,
        scanner: s,
    }
}

// Scan performs scan and emits progress events
func (s *ScanService) Scan(opts types.ScanOptions) error {
    if s.scanning {
        return errors.New("scan already in progress")
    }

    s.scanning = true
    defer func() { s.scanning = false }()

    // Emit start event
    s.app.EmitEvent("scan:started", nil)

    // Scan with progress
    results, err := s.scanner.ScanAll(opts)
    if err != nil {
        s.app.EmitEvent("scan:error", err.Error())
        return err
    }

    s.results = results

    // Emit complete event
    s.app.EmitEvent("scan:complete", results)
    return nil
}

// GetResults returns cached scan results
func (s *ScanService) GetResults() []types.ScanResult {
    return s.results
}

// IsScanning returns scan status
func (s *ScanService) IsScanning() bool {
    return s.scanning
}
```

**Bindings Generated:**

```typescript
// frontend/src/lib/wailsjs/go/services/ScanService.ts
export function Scan(opts: types.ScanOptions): Promise<void>
export function GetResults(): Promise<types.ScanResult[]>
export function IsScanning(): Promise<boolean>
```

---

### Service 2: TreeService

**File:** `internal/services/tree_service.go`

```go
package services

import (
    "github.com/thanhdevapp/dev-cleaner/internal/scanner"
    "github.com/thanhdevapp/dev-cleaner/pkg/types"
    "github.com/wailsapp/wails/v3/pkg/application"
)

type TreeService struct {
    app     *application.App
    scanner *scanner.Scanner
    cache   map[string]*types.TreeNode // path → node
}

func NewTreeService(app *application.App) *TreeService {
    s, _ := scanner.New()
    return &TreeService{
        app:     app,
        scanner: s,
        cache:   make(map[string]*types.TreeNode),
    }
}

// GetTreeNode lazily scans directory
func (t *TreeService) GetTreeNode(path string, depth int) (*types.TreeNode, error) {
    // Check cache
    if node, exists := t.cache[path]; exists && node.Scanned {
        return node, nil
    }

    // Scan
    node, err := t.scanner.ScanDirectory(path, depth, 5)
    if err != nil {
        return nil, err
    }

    // Cache
    t.cache[path] = node

    // Emit event
    t.app.EmitEvent("tree:updated", node)
    return node, nil
}

// ClearCache clears tree cache
func (t *TreeService) ClearCache() {
    t.cache = make(map[string]*types.TreeNode)
}
```

---

### Service 3: CleanService

**File:** `internal/services/clean_service.go`

```go
package services

import (
    "github.com/thanhdevapp/dev-cleaner/internal/cleaner"
    "github.com/thanhdevapp/dev-cleaner/pkg/types"
    "github.com/wailsapp/wails/v3/pkg/application"
)

type CleanService struct {
    app      *application.App
    cleaner  *cleaner.Cleaner
    cleaning bool
}

func NewCleanService(app *application.App, dryRun bool) *CleanService {
    c, _ := cleaner.New(dryRun)
    return &CleanService{
        app:     app,
        cleaner: c,
    }
}

// Clean deletes selected items with progress
func (c *CleanService) Clean(items []types.ScanResult) ([]cleaner.CleanResult, error) {
    if c.cleaning {
        return nil, errors.New("clean already in progress")
    }

    c.cleaning = true
    defer func() { c.cleaning = false }()

    c.app.EmitEvent("clean:started", len(items))

    results, err := c.cleaner.Clean(items)

    if err != nil {
        c.app.EmitEvent("clean:error", err.Error())
        return results, err
    }

    // Calculate freed space
    var freedSpace int64
    for _, r := range results {
        if r.Success {
            freedSpace += r.Size
        }
    }

    c.app.EmitEvent("clean:complete", map[string]interface{}{
        "results":    results,
        "freedSpace": freedSpace,
    })

    return results, nil
}

// IsCleaning returns clean status
func (c *CleanService) IsCleaning() bool {
    return c.cleaning
}
```

---

### Service 4: SettingsService

**File:** `internal/services/settings_service.go`

```go
package services

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type Settings struct {
    Theme           string   `json:"theme"`           // "light" | "dark" | "auto"
    DefaultView     string   `json:"defaultView"`     // "list" | "treemap" | "split"
    AutoScan        bool     `json:"autoScan"`        // Scan on launch
    ConfirmDelete   bool     `json:"confirmDelete"`   // Show confirm dialog
    ScanCategories  []string `json:"scanCategories"`  // ["xcode", "android", "node"]
    MaxDepth        int      `json:"maxDepth"`        // Tree depth limit
}

type SettingsService struct {
    settings Settings
    path     string
}

func NewSettingsService() *SettingsService {
    home, _ := os.UserHomeDir()
    path := filepath.Join(home, ".dev-cleaner-gui.json")

    s := &SettingsService{path: path}
    s.Load()
    return s
}

func (s *SettingsService) Load() error {
    data, err := os.ReadFile(s.path)
    if err != nil {
        // Default settings
        s.settings = Settings{
            Theme:          "auto",
            DefaultView:    "split",
            AutoScan:       true,
            ConfirmDelete:  true,
            ScanCategories: []string{"xcode", "android", "node"},
            MaxDepth:       5,
        }
        return nil
    }

    return json.Unmarshal(data, &s.settings)
}

func (s *SettingsService) Save() error {
    data, _ := json.MarshalIndent(s.settings, "", "  ")
    return os.WriteFile(s.path, data, 0644)
}

func (s *SettingsService) Get() Settings {
    return s.settings
}

func (s *SettingsService) Update(settings Settings) error {
    s.settings = settings
    return s.Save()
}
```

---

### Wails App Setup

**File:** `cmd/gui/app.go`

```go
package main

import (
    "github.com/thanhdevapp/dev-cleaner/internal/services"
    "github.com/wailsapp/wails/v3/pkg/application"
)

type App struct {
    app            *application.App
    scanService    *services.ScanService
    treeService    *services.TreeService
    cleanService   *services.CleanService
    settingsService *services.SettingsService
}

func NewApp() *App {
    return &App{}
}

func (a *App) Startup(app *application.App) {
    a.app = app
    a.scanService = services.NewScanService(app)
    a.treeService = services.NewTreeService(app)
    a.cleanService = services.NewCleanService(app, false)
    a.settingsService = services.NewSettingsService()
}

// Expose services to frontend
func (a *App) ScanService() *services.ScanService {
    return a.scanService
}

func (a *App) TreeService() *services.TreeService {
    return a.treeService
}

func (a *App) CleanService() *services.CleanService {
    return a.cleanService
}

func (a *App) SettingsService() *services.SettingsService {
    return a.settingsService
}
```

**File:** `cmd/gui/main.go`

```go
package main

import (
    "embed"
    "log"
    "github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    app := application.New(application.Options{
        Name:        "Mac Dev Cleaner",
        Description: "Clean development artifacts on macOS",
        Services: []application.Service{
            application.NewService(&App{}),
        },
        Assets: application.AssetOptions{
            Handler: application.AssetFileServerFS(assets),
        },
        Mac: application.MacOptions{
            ApplicationShouldTerminateAfterLastWindowClosed: true,
        },
    })

    // Create main window
    app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
        Title:  "Mac Dev Cleaner",
        Width:  1200,
        Height: 800,
        Mac: application.MacWindow{
            InvisibleTitleBarHeight: 50,
            Backdrop:                application.MacBackdropTranslucent,
        },
    })

    err := app.Run()
    if err != nil {
        log.Fatal(err)
    }
}
```

---

## React Frontend Architecture

### Tech Stack

**Core:**
- React 18 (with Suspense, Transitions)
- TypeScript (strict mode)
- Vite (build tool)

**State Management:**
- Zustand (lightweight, simple)
- Alternative: Jotai (atomic state)

**UI Components:**
- shadcn/ui (Radix + Tailwind)
- Recharts (treemap visualization)
- react-window (virtual scrolling)

**Styling:**
- Tailwind CSS (utility-first)
- CSS variables for theming

---

### State Management (Zustand)

**File:** `frontend/src/store/uiStore.ts`

```typescript
import { create } from 'zustand'
import { devtools } from 'zustand/middleware'

interface UIState {
  // Selection
  selectedPaths: Set<string>
  toggleSelection: (path: string) => void
  clearSelection: () => void
  selectAll: (paths: string[]) => void

  // Tree expansion
  expandedNodes: Set<string>
  toggleExpand: (path: string) => void
  expandPath: (path: string) => void

  // Filters
  filters: {
    search: string
    types: string[]
    minSize: number
    maxSize: number
  }
  updateFilters: (filters: Partial<UIState['filters']>) => void

  // View mode
  viewMode: 'list' | 'treemap' | 'split'
  setViewMode: (mode: UIState['viewMode']) => void

  // UI state
  isSidebarOpen: boolean
  isSettingsOpen: boolean
  toggleSidebar: () => void
  toggleSettings: () => void
}

export const useUIStore = create<UIState>()(
  devtools(
    (set) => ({
      selectedPaths: new Set(),
      toggleSelection: (path) =>
        set((state) => {
          const newSet = new Set(state.selectedPaths)
          if (newSet.has(path)) {
            newSet.delete(path)
          } else {
            newSet.add(path)
          }
          return { selectedPaths: newSet }
        }),
      clearSelection: () => set({ selectedPaths: new Set() }),
      selectAll: (paths) => set({ selectedPaths: new Set(paths) }),

      expandedNodes: new Set(),
      toggleExpand: (path) =>
        set((state) => {
          const newSet = new Set(state.expandedNodes)
          if (newSet.has(path)) {
            newSet.delete(path)
          } else {
            newSet.add(path)
          }
          return { expandedNodes: newSet }
        }),
      expandPath: (path) =>
        set((state) => ({
          expandedNodes: new Set([...state.expandedNodes, path]),
        })),

      filters: {
        search: '',
        types: [],
        minSize: 0,
        maxSize: Infinity,
      },
      updateFilters: (filters) =>
        set((state) => ({
          filters: { ...state.filters, ...filters },
        })),

      viewMode: 'split',
      setViewMode: (mode) => set({ viewMode: mode }),

      isSidebarOpen: true,
      isSettingsOpen: false,
      toggleSidebar: () =>
        set((state) => ({ isSidebarOpen: !state.isSidebarOpen })),
      toggleSettings: () =>
        set((state) => ({ isSettingsOpen: !state.isSettingsOpen })),
    }),
    { name: 'ui-store' }
  )
)
```

---

### Custom Hooks

**File:** `frontend/src/hooks/useScanResults.ts`

```typescript
import { useState, useEffect } from 'react'
import { EventsOn } from '@/lib/wailsjs/runtime/runtime'
import { GetResults } from '@/lib/wailsjs/go/services/ScanService'
import type { types } from '@/types/bindings'

export function useScanResults() {
  const [results, setResults] = useState<types.ScanResult[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    // Listen for scan events
    const unsubComplete = EventsOn('scan:complete', (data: types.ScanResult[]) => {
      setResults(data)
      setLoading(false)
    })

    const unsubError = EventsOn('scan:error', (err: string) => {
      setError(err)
      setLoading(false)
    })

    const unsubStarted = EventsOn('scan:started', () => {
      setLoading(true)
      setError(null)
    })

    // Load initial results
    GetResults().then(setResults)

    return () => {
      unsubComplete()
      unsubError()
      unsubStarted()
    }
  }, [])

  return { results, loading, error }
}
```

**File:** `frontend/src/hooks/useTreeNode.ts`

```typescript
import { useState, useCallback } from 'react'
import { GetTreeNode } from '@/lib/wailsjs/go/services/TreeService'
import type { types } from '@/types/bindings'

export function useTreeNode(path: string) {
  const [node, setNode] = useState<types.TreeNode | null>(null)
  const [loading, setLoading] = useState(false)

  const fetchNode = useCallback(async (depth: number = 0) => {
    setLoading(true)
    try {
      const result = await GetTreeNode(path, depth)
      setNode(result)
    } catch (err) {
      console.error('Failed to fetch tree node:', err)
    } finally {
      setLoading(false)
    }
  }, [path])

  return { node, loading, fetchNode }
}
```

---

### Components

**File:** `frontend/src/components/FileTreeList.tsx`

```typescript
import { FixedSizeList as List } from 'react-window'
import { Checkbox } from '@/components/ui/checkbox'
import { ChevronRight, ChevronDown, Folder, File } from 'lucide-react'
import { useUIStore } from '@/store/uiStore'
import { formatBytes } from '@/lib/utils'
import type { types } from '@/types/bindings'

interface Props {
  results: types.ScanResult[]
  height: number
}

export function FileTreeList({ results, height }: Props) {
  const { selectedPaths, toggleSelection, expandedNodes, toggleExpand } = useUIStore()

  const Row = ({ index, style }: { index: number; style: React.CSSProperties }) => {
    const item = results[index]
    const isExpanded = expandedNodes.has(item.path)
    const isSelected = selectedPaths.has(item.path)

    return (
      <div style={style} className="flex items-center gap-2 px-4 hover:bg-accent">
        <button onClick={() => toggleExpand(item.path)}>
          {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
        </button>

        <Checkbox
          checked={isSelected}
          onCheckedChange={() => toggleSelection(item.path)}
        />

        <Folder size={16} className="text-blue-500" />

        <span className="flex-1">{item.name}</span>

        <span className="text-sm text-muted-foreground">
          {formatBytes(item.size)}
        </span>

        <span className="text-xs text-muted-foreground">
          {item.fileCount} files
        </span>
      </div>
    )
  }

  return (
    <List
      height={height}
      itemCount={results.length}
      itemSize={40}
      width="100%"
    >
      {Row}
    </List>
  )
}
```

**File:** `frontend/src/components/Treemap.tsx`

```typescript
import { Treemap, ResponsiveContainer, Tooltip } from 'recharts'
import { formatBytes } from '@/lib/utils'
import type { types } from '@/types/bindings'

interface Props {
  results: types.ScanResult[]
  onItemClick: (item: types.ScanResult) => void
}

export function TreemapChart({ results, onItemClick }: Props) {
  // Transform data for recharts
  const data = results.map((item) => ({
    name: item.name,
    size: item.size,
    type: item.type,
    path: item.path,
  }))

  const COLORS = {
    xcode: '#147EFB',
    android: '#3DDC84',
    node: '#68A063',
  }

  return (
    <ResponsiveContainer width="100%" height="100%">
      <Treemap
        data={data}
        dataKey="size"
        stroke="#fff"
        fill="#8884d8"
        content={({ x, y, width, height, name, size, type }) => (
          <g>
            <rect
              x={x}
              y={y}
              width={width}
              height={height}
              style={{
                fill: COLORS[type as keyof typeof COLORS],
                stroke: '#fff',
                strokeWidth: 2,
                cursor: 'pointer',
              }}
              onClick={() => {
                const item = results.find((r) => r.name === name)
                if (item) onItemClick(item)
              }}
            />
            {width > 60 && height > 30 && (
              <>
                <text
                  x={x + width / 2}
                  y={y + height / 2 - 10}
                  textAnchor="middle"
                  fill="#fff"
                  fontSize={12}
                >
                  {name}
                </text>
                <text
                  x={x + width / 2}
                  y={y + height / 2 + 10}
                  textAnchor="middle"
                  fill="#fff"
                  fontSize={10}
                  opacity={0.8}
                >
                  {formatBytes(size)}
                </text>
              </>
            )}
          </g>
        )}
      >
        <Tooltip
          content={({ active, payload }) => {
            if (active && payload && payload.length) {
              const data = payload[0].payload
              return (
                <div className="bg-popover p-2 rounded border">
                  <p className="font-semibold">{data.name}</p>
                  <p className="text-sm">{formatBytes(data.size)}</p>
                  <p className="text-xs text-muted-foreground">{data.type}</p>
                </div>
              )
            }
            return null
          }}
        />
      </Treemap>
    </ResponsiveContainer>
  )
}
```

---

## Communication Patterns

### Pattern 1: Method Calls (Frontend → Backend)

```typescript
// React component
import { Scan } from '@/lib/wailsjs/go/services/ScanService'

async function handleScan() {
  await Scan({
    includeXcode: true,
    includeAndroid: true,
    includeNode: true,
    maxDepth: 5,
  })
}
```

**Flow:**
1. User clicks "Scan" button
2. React calls `Scan()` (auto-generated binding)
3. Go `ScanService.Scan()` executes
4. Go emits `scan:complete` event
5. React hook receives event, updates UI

---

### Pattern 2: Events (Backend → Frontend)

```typescript
// React hook
import { EventsOn } from '@/lib/wailsjs/runtime/runtime'

useEffect(() => {
  const unsubscribe = EventsOn('scan:progress', (progress: number) => {
    setProgress(progress)
  })

  return unsubscribe
}, [])
```

**Event Types:**
- `scan:started` - Scan begins
- `scan:progress` - Progress update (0-100)
- `scan:complete` - Scan done, data included
- `scan:error` - Scan failed
- `tree:updated` - Tree node loaded
- `clean:started` - Cleaning begins
- `clean:complete` - Cleaning done

---

### Pattern 3: Lazy Loading Trees

```typescript
// Tree node expansion
async function handleExpand(path: string) {
  // Mark as expanded in UI immediately (optimistic)
  toggleExpand(path)

  // Lazy load children from Go
  const node = await GetTreeNode(path, 0)

  // Update local state with children
  updateTreeData(path, node.children)
}
```

---

## Code Reuse Strategy

### Shared Packages

**Already exist (reuse as-is):**
- `internal/scanner/` - All scanning logic
- `internal/cleaner/` - All cleaning logic
- `pkg/types/` - Type definitions

**Need extraction:**
- Extract TUI tree logic from `internal/tui/tui.go`
- Move to `pkg/tree/` for sharing
- Both TUI and GUI use same tree navigation

**New for GUI:**
- `internal/services/` - Wails services
- `cmd/gui/` - GUI entry point
- `frontend/` - React app

---

### Migration Path

**Phase 1: Minimal changes to existing code**
1. Keep `cmd/cli/` exactly as-is
2. Add new `cmd/gui/` without touching CLI
3. Import `internal/scanner` and `internal/cleaner` directly

**Phase 2: Refactor for sharing** (optional, after GUI works)
1. Extract common tree logic to `pkg/tree/`
2. Both CLI TUI and GUI use `pkg/tree/`
3. DRY achieved without breaking CLI

---

## Implementation Phases (1 Month)

### Week 1: Foundation

**Days 1-2: Wails v3 Setup**
- Install Wails v3: `go install github.com/wailsapp/wails/v3/cmd/wails3@latest`
- Init project: `wails3 init -n gui -t react-ts`
- Copy generated files to `cmd/gui/` and `frontend/`
- Configure `wails.json`
- Verify dev mode: `wails3 dev`

**Days 3-4: Services Layer**
- Implement `ScanService` with events
- Implement `TreeService` with lazy loading
- Implement `CleanService` with progress
- Implement `SettingsService`
- Test bindings generation: `wails3 generate bindings`

**Days 5-7: Basic UI**
- Setup Tailwind + shadcn/ui
- Create `App.tsx` layout
- Implement `Toolbar` component
- Implement `FileTreeList` (basic, no virtualization yet)
- Wire up scan button → service → display results

**Deliverable:** Basic GUI that can scan and display flat list

---

### Week 2: Tree Navigation

**Days 8-10: Tree List Component**
- Implement tree expansion logic
- Add checkbox selection
- Integrate `useTreeNode` hook for lazy loading
- Add virtual scrolling (react-window)
- Handle large datasets (10K+ items)

**Days 11-14: Treemap Visualization**
- Setup Recharts
- Implement `TreemapChart` component
- Color by category
- Add drill-down on click
- Sync selection between tree list and treemap

**Deliverable:** Full tree + treemap visualization working

---

### Week 3: Operations & UX

**Days 15-17: Clean Operations**
- Implement `CleanDialog` component
- Show confirmation with size summary
- Integrate `CleanService`
- Show progress during cleaning
- Display results (success/errors)

**Days 18-19: Settings**
- Implement `SettingsDialog`
- Theme switcher (light/dark)
- Default view preference
- Scan categories config
- Persist settings

**Days 20-21: Polish**
- Add loading states
- Error handling UI
- Empty states
- Tooltips
- Keyboard shortcuts

**Deliverable:** Full-featured GUI with clean + settings

---

### Week 4: Testing & Distribution

**Days 22-24: Testing**
- Manual testing checklist
- Edge cases (permissions, large folders)
- Performance testing (10K+ files)
- Memory leak testing
- Cross-check with CLI results

**Days 25-27: Distribution**
- Build production: `wails3 build`
- Code signing (macOS)
- Create DMG installer
- GitHub Actions workflow
- Release v1.0.0-gui

**Days 28-30: Documentation**
- Update README with GUI screenshots
- Write user guide
- Architecture documentation
- Developer setup guide
- Release notes

**Deliverable:** Production-ready GUI v1.0.0

---

## Technical Risks & Mitigation

### Risk 1: Wails v3 Alpha Instability

**Issue:** API changes, bugs in alpha version

**Mitigation:**
- Pin to specific v3 alpha commit
- Follow Wails changelog/Discord
- Have rollback plan to v2 if critical bugs
- Budget 2-3 days for unexpected v3 issues

---

### Risk 2: Large Dataset Performance

**Issue:** 10K+ files in React state = lag

**Mitigation:**
- Virtual scrolling (react-window)
- Lazy tree loading (don't load all upfront)
- Keep heavy data in Go, send only visible to React
- Pagination if needed

---

### Risk 3: Treemap Rendering Performance

**Issue:** Complex treemap with 1000+ rectangles = slow

**Mitigation:**
- Limit treemap to top 100 items
- Use WebGL rendering if needed (recharts limitation)
- Fallback to simpler bar chart for large datasets
- Progressive loading

---

### Risk 4: State Sync Between Tree & Treemap

**Issue:** Selection state out of sync

**Mitigation:**
- Single source of truth: Zustand store
- Both components read from same store
- No duplicate state
- Unit tests for sync logic

---

### Risk 5: Timeline Slip (1 Month Tight)

**Issue:** Unexpected complexity, bugs

**Mitigation:**
- MVP-first approach: Get basic GUI working Week 1
- Treemap is nice-to-have (Week 2), can cut if needed
- Polish (Week 3) is flexible
- Distribution (Week 4) can extend to Week 5

**Fallback plan:**
- Week 1: Basic scan + list (must have)
- Week 2: Tree navigation (must have)
- Week 3: Clean operations (must have)
- Week 4+: Treemap, settings, polish (nice-to-have)

---

## Success Metrics

**Must achieve:**

1. ✅ GUI launches and scans successfully
2. ✅ Displays scan results in tree list
3. ✅ Lazy tree navigation works (drill down, go back)
4. ✅ Selection and clean operations functional
5. ✅ No regression in CLI functionality
6. ✅ Performance: Handle 10K+ files smoothly
7. ✅ Memory: < 200MB RAM usage
8. ✅ Package size: < 50MB .app bundle

**Nice to have:**

1. Treemap visualization working
2. Settings persisted
3. Dark mode
4. Keyboard shortcuts
5. macOS native feel (translucent, animations)

---

## Alternative Approaches (Rejected)

### Alternative 1: Electron + React

**Why Rejected:**
- 📦 Huge bundle size (>100MB)
- 🐌 Slower startup than Wails
- 💰 More memory usage
- ❌ Less native feel

Wails v3 is superior for macOS desktop apps.

---

### Alternative 2: Pure Web App (Browser-based)

**Why Rejected:**
- 🔒 No file system access (security sandbox)
- ❌ Can't delete files
- 📱 Wrong UX paradigm (not desktop app)
- 🚫 No menubar integration

Must be native desktop app per requirements.

---

### Alternative 3: Go-only State (No React State)

**Why Rejected:**
- 🐌 Every UI state change = Go call = slow
- ❌ Checkbox toggles, tree expansion = backend calls = laggy
- 😞 Poor UX (network latency for UI state)

Hybrid approach balances performance and simplicity.

---

## Unresolved Questions

1. **Multi-window layout:** Should settings be separate window or modal? Wails v3 supports multi-window - utilize it?

2. **Menubar app option:** Should we add menubar icon (runs in background)? Or traditional dock app only?

3. **Auto-update:** Implement auto-update mechanism? Sparkle framework integration?

4. **Analytics:** Add anonymous usage analytics? (Mixpanel, PostHog) Or fully offline?

5. **Treemap library:** Recharts vs D3 custom implementation? Recharts easier but less flexible.

6. **Theme:** Light + Dark only, or support auto (system preference)? Auto adds complexity.

7. **Localization:** English only or multi-language support? i18n adds significant overhead.

8. **Integration with CLI:** Should GUI be able to launch CLI commands? Or completely separate?

---

## Next Steps

1. **Get approval:** Review architecture decisions
2. **Resolve questions:** Answer 8 unresolved questions above
3. **Setup Wails v3:** Install and verify environment
4. **Create feature branch:** `git checkout -b feat/wails-gui`
5. **Start Week 1:** Wails setup + services layer

---

## Sources

Research sources used in this brainstorm:

### Wails Documentation:
- [Wails v3 Alpha Docs](https://v3alpha.wails.io/)
- [Wails v3 What's New](https://v3alpha.wails.io/whats-new/)
- [Wails v3 Bindings](https://v3alpha.wails.io/learn/bindings/)
- [Wails State Management Discussion](https://github.com/wailsapp/wails/discussions/2936)
- [Wails Application Development Guide](https://wails.io/docs/guides/application-development/)
- [Wails React Templates](https://wails.io/docs/community/templates/)

### React Best Practices:
- [React Architecture Patterns 2025 - GeeksforGeeks](https://www.geeksforgeeks.org/reactjs/react-architecture-pattern-and-best-practices/)
- [React Design Patterns 2025 - Telerik](https://www.telerik.com/blogs/react-design-patterns-best-practices)
- [React Best Practices 2025 - Technostacks](https://technostacks.com/blog/react-best-practices/)

### Wails + React Integration:
- [Wails React Tutorial - Markaicode](https://markaicode.com/wails-desktop-app-development-tutorial/)
- [Building Desktop Apps with Go, React and Wails - Medium](https://medium.com/@tomronw/mapping-success-building-a-simple-tracking-desktop-app-with-go-react-and-wails-ac83dbcbccca)
- [Wails + React + Vite Setup Guide - DEV Community](https://dev.to/dera_johnson/setting-up-a-desktop-project-with-wails-react-and-vite-a-step-by-step-guide-1b0m)

### Disk Visualization:
- [Awesome Wails Applications](https://github.com/wailsapp/awesome-wails)
- [Disk Space Visualization Tools](https://www.itechtics.com/15-tools-visualize-file-system-usage-windows/)
- [Treemap Visualization - FolderSizes](https://www.foldersizes.com/screens/disk-space-treemap/)

---

**Report generated:** 2025-12-15
**Timeline:** 1 month (4 weeks)
**Complexity:** High (but achievable with advanced Wails experience)
**Recommendation:** Proceed with Hybrid state + Wails v3 architecture
