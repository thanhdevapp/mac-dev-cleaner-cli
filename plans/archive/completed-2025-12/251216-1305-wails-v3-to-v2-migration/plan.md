# Wails v3 to v2 Migration Plan

**Project:** Mac Dev Cleaner GUI
**Date:** 2025-12-16
**Estimated Effort:** 4-6 hours
**Risk Level:** Medium

---

## Executive Summary

Migrate from Wails v3 alpha (unstable) to Wails v2 stable for improved runtime stability, better Events API, and production-ready bindings generation.

### Current State
- **Wails v3:** `v3.0.0-alpha.47` (Go), `@wailsio/runtime v3.0.0-alpha.77` (npm)
- **Issues:** Runtime initialization problems, Events API instability, bundler compatibility

### Target State
- **Wails v2:** `v2.9.x` or `v2.10.x` (latest stable)
- **Runtime:** Built-in via `wailsjs/` generated directory (no npm package needed)

---

## Phase 1: Environment Preparation

### 1.1 Requirements Verification

| Requirement | Current | Target | Action |
|------------|---------|--------|--------|
| Go version | 1.25.5 | 1.21+ (macOS 15 requires 1.23.3+) | Already compatible |
| Node.js | - | 15+ | Verify with `node -v` |
| Wails CLI | v3 alpha | v2 latest | Reinstall |

### 1.2 Install Wails v2 CLI

```bash
# Remove v3 CLI if installed globally
go clean -cache

# Install Wails v2 CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Verify installation
wails doctor
```

---

## Phase 2: Go Backend Migration

### 2.1 Update go.mod

**Current (v3):**
```go
require (
    github.com/wailsapp/wails/v3 v3.0.0-alpha.47 // indirect
)
```

**Target (v2):**
```go
require (
    github.com/wailsapp/wails/v2 v2.9.2
)
```

**Commands:**
```bash
# Remove v3 dependency
go mod edit -droprequire github.com/wailsapp/wails/v3

# Add v2 dependency
go get github.com/wailsapp/wails/v2@latest

# Clean up
go mod tidy
```

### 2.2 Update main.go

**File:** `cmd/gui/main.go`

**Current (v3):**
```go
package main

import (
    "log"
    "os"
    "github.com/wailsapp/wails/v3/pkg/application"
)

func main() {
    wailsApp := application.New(application.Options{
        Name:        "Mac Dev Cleaner",
        Description: "Clean development artifacts on macOS",
        Assets: application.AssetOptions{
            Handler: application.AssetFileServerFS(os.DirFS("frontend/dist")),
        },
        Mac: application.MacOptions{
            ApplicationShouldTerminateAfterLastWindowClosed: true,
        },
    })

    app := NewAppWithWails(wailsApp)
    wailsApp.RegisterService(application.NewService(app))
    wailsApp.Window.New()

    if err := wailsApp.Run(); err != nil {
        log.Fatal(err)
    }
}
```

**Target (v2):**
```go
package main

import (
    "embed"
    "log"

    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/options/assetserver"
    "github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    app := NewApp()

    err := wails.Run(&options.App{
        Title:     "Mac Dev Cleaner",
        Width:     1200,
        Height:    800,
        MinWidth:  800,
        MinHeight: 600,
        AssetServer: &assetserver.Options{
            Assets: assets,
        },
        OnStartup:  app.startup,
        OnShutdown: app.shutdown,
        Mac: &mac.Options{
            TitleBar: &mac.TitleBar{
                TitlebarAppearsTransparent: true,
                HideTitle:                  false,
                HideTitleBar:               false,
                FullSizeContent:            true,
                UseToolbar:                 false,
            },
            About: &mac.AboutInfo{
                Title:   "Mac Dev Cleaner",
                Message: "Clean development artifacts on macOS - v1.0.0",
            },
        },
        Bind: []interface{}{
            app,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

### 2.3 Update app.go

**File:** `cmd/gui/app.go`

**Key Changes:**

| v3 Pattern | v2 Pattern |
|-----------|-----------|
| `*application.App` stored in struct | `context.Context` stored in struct |
| `app.Event.Emit(...)` | `runtime.EventsEmit(ctx, ...)` |
| `OnStartup(app *application.App)` | `startup(ctx context.Context)` |
| `OnShutdown() error` | `shutdown(ctx context.Context)` |
| Services receive `*application.App` | Services receive `context.Context` |

**Current (v3):**
```go
package main

import (
    "github.com/wailsapp/wails/v3/pkg/application"
    "github.com/thanhdevapp/dev-cleaner/internal/services"
)

type App struct {
    app             *application.App
    scanService     *services.ScanService
    // ...
}

func NewAppWithWails(app *application.App) *App {
    a := &App{app: app}
    a.scanService, _ = services.NewScanService(app)
    // ...
    return a
}

func (a *App) OnStartup(app *application.App) error {
    a.app = app
    // ...
    return nil
}
```

**Target (v2):**
```go
package main

import (
    "context"
    "log"

    "github.com/thanhdevapp/dev-cleaner/internal/cleaner"
    "github.com/thanhdevapp/dev-cleaner/internal/services"
    "github.com/thanhdevapp/dev-cleaner/pkg/types"
)

type App struct {
    ctx             context.Context
    scanService     *services.ScanService
    treeService     *services.TreeService
    cleanService    *services.CleanService
    settingsService *services.SettingsService
}

func NewApp() *App {
    return &App{}
}

func (a *App) startup(ctx context.Context) {
    log.Println("Initializing services...")
    a.ctx = ctx

    var err error
    a.scanService, err = services.NewScanService(ctx)
    if err != nil {
        log.Printf("Failed to create ScanService: %v", err)
        return
    }

    a.treeService, err = services.NewTreeService(ctx)
    if err != nil {
        log.Printf("Failed to create TreeService: %v", err)
        return
    }

    a.cleanService, err = services.NewCleanService(ctx, false)
    if err != nil {
        log.Printf("Failed to create CleanService: %v", err)
        return
    }

    a.settingsService = services.NewSettingsService()
    log.Println("All services initialized successfully!")
}

func (a *App) shutdown(ctx context.Context) {
    log.Println("Shutting down...")
}

// Exposed methods (unchanged signatures)
func (a *App) Scan(opts types.ScanOptions) error {
    if a.scanService == nil {
        return nil
    }
    return a.scanService.Scan(opts)
}

func (a *App) GetScanResults() []types.ScanResult {
    if a.scanService == nil {
        return []types.ScanResult{}
    }
    return a.scanService.GetResults()
}

// ... rest of exposed methods unchanged
```

### 2.4 Update Services to Use context.Context

**Files to modify:**
- `internal/services/scan_service.go`
- `internal/services/tree_service.go`
- `internal/services/clean_service.go`

**Pattern Change:**

**Current (v3):**
```go
package services

import (
    "github.com/wailsapp/wails/v3/pkg/application"
)

type ScanService struct {
    app *application.App
    // ...
}

func NewScanService(app *application.App) (*ScanService, error) {
    return &ScanService{app: app}, nil
}

func (s *ScanService) Scan(opts types.ScanOptions) error {
    s.app.Event.Emit("scan:started")
    // ...
    s.app.Event.Emit("scan:complete", results)
    return nil
}
```

**Target (v2):**
```go
package services

import (
    "context"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

type ScanService struct {
    ctx context.Context
    // ...
}

func NewScanService(ctx context.Context) (*ScanService, error) {
    return &ScanService{ctx: ctx}, nil
}

func (s *ScanService) Scan(opts types.ScanOptions) error {
    runtime.EventsEmit(s.ctx, "scan:started")
    // ...
    runtime.EventsEmit(s.ctx, "scan:complete", results)
    return nil
}
```

---

## Phase 3: Frontend Migration

### 3.1 Update package.json

**Remove:**
```json
{
  "dependencies": {
    "@wailsio/runtime": "latest"  // REMOVE THIS
  }
}
```

**Commands:**
```bash
cd frontend
npm uninstall @wailsio/runtime
```

### 3.2 Update main.tsx

**Current (v3):**
```tsx
import '@wailsio/runtime'
import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
```

**Target (v2):**
```tsx
import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
```

### 3.3 Update Binding Imports

**Bindings Location Change:**

| v3 | v2 |
|----|----|
| `frontend/bindings/github.com/thanhdevapp/dev-cleaner/cmd/gui/app` | `frontend/wailsjs/go/main/App` |
| `frontend/bindings/github.com/thanhdevapp/dev-cleaner/pkg/types/models` | `frontend/wailsjs/go/main/App` (all in one) |

**Current (v3) - toolbar.tsx:**
```tsx
import { Scan } from '../../bindings/github.com/thanhdevapp/dev-cleaner/cmd/gui/app'
import { ScanOptions } from '../../bindings/github.com/thanhdevapp/dev-cleaner/pkg/types/models'
```

**Target (v2) - toolbar.tsx:**
```tsx
import { Scan } from '../wailsjs/go/main/App'
```

### 3.4 Update Events API

**Current (v3) - scan-results.tsx:**
```tsx
import { Events } from '@wailsio/runtime'

useEffect(() => {
    const unsubscribeComplete = Events.On('scan:complete', (event) => {
        if (event.data && Array.isArray(event.data)) {
            setResults(event.data as ScanResult[])
        }
    })
    return () => {
        unsubscribeComplete?.()
    }
}, [])
```

**Target (v2) - scan-results.tsx:**
```tsx
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime'

useEffect(() => {
    // v2: callback receives data directly, not wrapped in event object
    EventsOn('scan:complete', (data: ScanResult[]) => {
        if (data && Array.isArray(data)) {
            setResults(data)
        }
    })
    return () => {
        EventsOff('scan:complete')
    }
}, [])
```

**Key Differences:**
| v3 | v2 |
|----|----|
| `Events.On(name, (event) => event.data)` | `EventsOn(name, (data) => data)` |
| `Events.Emit(name, data)` | `EventsEmit(name, data)` |
| Returns unsubscribe function | Use `EventsOff(name)` to unsubscribe |

### 3.5 Files Requiring Import Updates

| File | Changes Required |
|------|-----------------|
| `frontend/src/main.tsx` | Remove `@wailsio/runtime` import |
| `frontend/src/components/toolbar.tsx` | Update binding imports |
| `frontend/src/components/scan-results.tsx` | Update binding imports + Events API |
| `frontend/src/components/file-tree-list.tsx` | Update if using bindings |
| `frontend/src/components/treemap-chart.tsx` | Update if using bindings |

---

## Phase 4: Configuration Migration

### 4.1 Update wails.json

**Current (v3):**
```json
{
  "name": "Mac Dev Cleaner",
  "outputfilename": "dev-cleaner-gui",
  "frontend:install": "cd frontend && npm install",
  "frontend:build": "cd frontend && npm run build",
  "frontend:dev": "cd frontend && npm run dev",
  "frontend:dev:watcher": "cd frontend && npm run dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "thanhdevapp",
    "email": "thanhdevapp@gmail.com"
  },
  "info": {
    "companyName": "DevTools",
    "productName": "Mac Dev Cleaner",
    "productVersion": "1.0.0",
    "copyright": "Copyright 2025",
    "comments": "Clean development artifacts on macOS"
  }
}
```

**Target (v2):**
```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "Mac Dev Cleaner",
  "outputfilename": "dev-cleaner-gui",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "frontend:dir": "frontend",
  "wailsjsdir": "frontend/wailsjs",
  "author": {
    "name": "thanhdevapp",
    "email": "thanhdevapp@gmail.com"
  },
  "info": {
    "companyName": "DevTools",
    "productName": "Mac Dev Cleaner",
    "productVersion": "1.0.0",
    "copyright": "Copyright 2025",
    "comments": "Clean development artifacts on macOS"
  }
}
```

**Key Changes:**
- Added `$schema` for v2 config validation
- Removed `cd frontend &&` prefix from commands (v2 handles frontend:dir)
- Added `frontend:dir` to specify frontend location
- Added `wailsjsdir` to specify bindings output location

---

## Phase 5: Build & Test

### 5.1 Clean Old Artifacts

```bash
# Remove v3 bindings
rm -rf frontend/bindings

# Remove node_modules to force fresh install
rm -rf frontend/node_modules
rm -f frontend/package-lock.json

# Clean Go cache
go clean -cache
```

### 5.2 Regenerate Bindings

```bash
# Generate v2 bindings
wails generate module
```

This creates `frontend/wailsjs/` with:
```
frontend/wailsjs/
├── go/
│   └── main/
│       ├── App.js       # Generated bindings
│       └── App.d.ts     # TypeScript definitions
└── runtime/
    └── runtime.js       # Wails runtime (Events, Window, etc.)
```

### 5.3 Development Build

```bash
# Install frontend dependencies
cd frontend && npm install && cd ..

# Run in dev mode
wails dev
```

### 5.4 Production Build

```bash
wails build
```

---

## Phase 6: Code Changes Summary

### 6.1 Go Files to Modify

| File | Changes |
|------|---------|
| `go.mod` | Replace v3 with v2 dependency |
| `cmd/gui/main.go` | Complete rewrite for v2 API |
| `cmd/gui/app.go` | Change to context-based pattern |
| `internal/services/scan_service.go` | Replace `*application.App` with `context.Context`, update Events |
| `internal/services/tree_service.go` | Same as scan_service |
| `internal/services/clean_service.go` | Same as scan_service |
| `internal/services/settings_service.go` | No changes needed (no Wails dependency) |

### 6.2 Frontend Files to Modify

| File | Changes |
|------|---------|
| `frontend/package.json` | Remove `@wailsio/runtime` |
| `frontend/src/main.tsx` | Remove runtime import |
| `frontend/src/components/toolbar.tsx` | Update binding imports |
| `frontend/src/components/scan-results.tsx` | Update binding imports + Events API |
| Any file using bindings | Update import paths |

### 6.3 Configuration Files to Modify

| File | Changes |
|------|---------|
| `wails.json` | Add v2 schema, add frontend:dir, update commands |

---

## Risk Assessment

### High Risk
- **Events API Change:** Data structure differs (v2 passes data directly, v3 wraps in event object)
  - **Mitigation:** Test all event handlers thoroughly

### Medium Risk
- **Binding Generation:** Different structure and import paths
  - **Mitigation:** Generate bindings first, then update imports

### Low Risk
- **Context Pattern:** Well-documented migration pattern
- **Configuration:** Minor changes to wails.json

---

## Rollback Plan

1. **Git Branch:** Work on separate branch (`feat/wails-v2-migration`)
2. **Backup:** Current v3 code remains on `feat/wails-gui` branch
3. **Revert Steps:**
   ```bash
   git checkout feat/wails-gui
   go mod tidy
   cd frontend && npm install
   ```

---

## Testing Checklist

- [ ] Application starts without errors
- [ ] Scan functionality works
- [ ] Events propagate correctly (scan:started, scan:complete)
- [ ] Results display in UI
- [ ] Selection/toggle works
- [ ] Clean functionality works
- [ ] Settings persist
- [ ] Window controls work (minimize, close)
- [ ] Production build succeeds
- [ ] macOS-specific features work (title bar, about menu)

---

## Migration Steps (Ordered)

1. Create new branch: `git checkout -b feat/wails-v2-migration`
2. Install Wails v2 CLI
3. Update `go.mod` - replace v3 with v2
4. Update `cmd/gui/main.go` - complete rewrite
5. Update `cmd/gui/app.go` - context pattern
6. Update services - context + Events API
7. Update `wails.json`
8. Remove `@wailsio/runtime` from frontend
9. Delete old bindings: `rm -rf frontend/bindings`
10. Generate v2 bindings: `wails generate module`
11. Update frontend imports
12. Update Events API usage
13. Test with `wails dev`
14. Production build: `wails build`
15. Run test checklist

---

## Appendix: API Reference Quick Comparison

### Events

| Operation | v3 | v2 |
|-----------|----|----|
| Emit (Go) | `app.Event.Emit("name", data)` | `runtime.EventsEmit(ctx, "name", data)` |
| On (JS) | `Events.On("name", (e) => e.data)` | `EventsOn("name", (data) => data)` |
| Off (JS) | `unsubscribe()` | `EventsOff("name")` |

### Application

| Operation | v3 | v2 |
|-----------|----|----|
| Create | `application.New(Options{})` | `wails.Run(&options.App{})` |
| Window | `app.Window.New()` | Configured in options |
| Assets | `application.AssetFileServerFS(fs)` | `embed.FS` via `//go:embed` |

### Lifecycle

| Operation | v3 | v2 |
|-----------|----|----|
| Startup | `OnStartup(app *application.App)` | `startup(ctx context.Context)` |
| Shutdown | `OnShutdown() error` | `shutdown(ctx context.Context)` |

---

## Unresolved Questions

None at this time. The migration path is well-documented and the project structure aligns with standard Wails patterns.
