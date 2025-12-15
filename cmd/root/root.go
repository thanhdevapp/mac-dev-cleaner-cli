// Package cmd contains Cobra CLI commands
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time
	Version = "dev"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "dev-cleaner",
	Short: "Clean development artifacts on macOS",
	Long: `Mac Dev Cleaner - A CLI tool to clean development project artifacts

Quickly free up disk space by removing:
  • Xcode DerivedData and caches
  • Android Gradle caches
  • Node.js node_modules directories
  • Package manager caches (npm, yarn, pnpm, bun)

Examples:
  dev-cleaner scan              # Scan and show all cleanable items
  dev-cleaner scan --ios        # Scan iOS/Xcode only
  dev-cleaner clean             # Interactive clean (dry-run by default)
  dev-cleaner clean --confirm   # Actually delete selected items`,
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
