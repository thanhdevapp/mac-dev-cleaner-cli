# Development Guide (Wails v2)

## Prerequisites
- Go 1.21+
- Node.js 16+
- Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

## Running the App
```bash
# Start backend + frontend with hot reload
wails dev
# OR
./run-gui.sh
```

## Building
```bash
# Build Mac App
wails build
```

## CLI Tool
The CLI entry point has moved to `cmd/dev-cleaner`.
```bash
go run ./cmd/dev-cleaner [command]
```
