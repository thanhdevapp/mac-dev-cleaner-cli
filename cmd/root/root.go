// Package cmd contains Cobra CLI commands
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time
	Version = "1.0.2"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "dev-cleaner",
	Short: "Clean development artifacts on macOS",
	Long: `Mac Dev Cleaner - A CLI tool to clean development project artifacts

Quickly free up disk space by removing:
  • Xcode DerivedData, Archives, and caches
  • Android Gradle caches and SDK artifacts
  • Node.js node_modules directories
  • Package manager caches (npm, yarn, pnpm, bun)
  • Flutter/Dart build artifacts and pub-cache

Features:
  ✨ Interactive TUI with keyboard navigation
  ✨ Tree mode for exploring nested folders
  ✨ Dry-run mode by default (safe preview)
  ✨ Quick clean with single keypress
  ✨ Batch operations for multiple items
  ✨ Real-time deletion progress tracking

Scan Examples:
  dev-cleaner scan                    # Scan all + launch TUI
  dev-cleaner scan --ios              # Scan iOS/Xcode only
  dev-cleaner scan --android          # Scan Android/Gradle only
  dev-cleaner scan --node             # Scan Node.js artifacts only
  dev-cleaner scan --flutter          # Scan Flutter/Dart only
  dev-cleaner scan --no-tui           # Text output without TUI

Clean Examples:
  dev-cleaner clean                   # Interactive TUI (dry-run)
  dev-cleaner clean --confirm         # Interactive TUI (actually delete)
  dev-cleaner clean --ios --confirm   # Clean iOS artifacts only
  dev-cleaner clean --no-tui          # Simple text mode cleanup

TUI Keyboard Shortcuts:
  ↑/↓, k/j     Navigate up/down
  Space        Toggle selection
  a            Select all items
  n            Deselect all items
  c            Quick clean current item (single-item mode)
  Enter        Clean all selected items (batch mode)
  →/l          Drill down into folder (tree mode)
  ←/h          Go back to parent (in tree mode)
  ?            Show detailed help screen
  q            Quit

Tip: Use 'dev-cleaner scan' to start the interactive TUI mode!`,
	Version: Version,
}

// Execute adds all child commands to the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here if needed
}
