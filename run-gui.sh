#!/bin/bash
# Run Wails v2 in dev mode
if command -v wails &> /dev/null; then
    wails dev
elif [ -f "$HOME/go/bin/wails" ]; then
    "$HOME/go/bin/wails" dev
else
    echo "Error: wails not found in PATH or ~/go/bin/wails"
    exit 1
fi
