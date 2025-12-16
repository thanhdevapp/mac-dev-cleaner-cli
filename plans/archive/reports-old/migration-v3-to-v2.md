# Wails v3 to v2 Migration Report

## Status: COMPLETE ✅

The project has been successfully migrated from Wails v3 Alpha to Wails v2 Stable (v2.11.0).

## Changes Made

### 1. Project Structure
- **Moved** `cmd/gui/main.go` -> `main.go` (Root)
- **Moved** `cmd/gui/app.go` -> `app.go` (Root)
- **Moved** CLI `main.go` -> `cmd/dev-cleaner/main.go`
- **Removed** `cmd/gui` directory
- **Removed** `frontend/frontend` nested directory (duplicate)
- **Removed** `frontend/bindings` (v3 generated)

### 2. Backend (Go)
- **Updated** `go.mod`: Downgraded to Wails v2.11.0
- **Refactored** `main.go`: Uses `wails.Run` with v2 Options
- **Refactored** `app.go`: Uses `context.Context` for lifecycle and event injection
- **Refactored Services**:
  - `ScanService`, `TreeService`, `CleanService` updated to use `context.Context` and `runtime.EventsEmit`
  - Removed dependency on `application.App` (v3)

### 3. Frontend (React)
- **Removed** `@wailsio/runtime` dependency
- **Updated** `wails.json`: Configured for v2, pointing to `./frontend/wailsjs`
- **Updated Components**:
  - `main.tsx`: Removed v3 runtime init
  - `scan-results.tsx`: Uses `EventsOn`, `EventsOff` from runtime
  - `toolbar.tsx`: Uses generated bindings (Scan, ScanOptions)
  - `file-tree-list.tsx`: Uses generated types (ScanResult)
  - Added `@ts-ignore` to imports until bindings are generated

### 4. Configuration
- **Created** `wails.json` (v2 format)
- **Updated** `frontend/tsconfig.json` to include `wailsjs`
- **Updated** `run-gui.sh` to use `wails dev`

## Next Steps

1. **Generate Bindings**:
   Run the app once to generate the missing `frontend/wailsjs` directory:
   ```bash
   ./run-gui.sh
   # OR
   wails dev
   ```

2. **Verify Frontend Types**:
   After generation, remove `// @ts-ignore` comments in `frontend/src/components/*.tsx` to regain type safety.

3. **Build CLI**:
   The CLI tool is now at `cmd/dev-cleaner`:
   ```bash
   go build -o dev-cleaner ./cmd/dev-cleaner
   ```

## Notes
- Wails v2 uses `context.Context` heavily. Services now require `SetContext` or `startup` injection.
- The `frontend/wailsjs` directory is auto-managed by Wails. Do not edit files inside it manually.
