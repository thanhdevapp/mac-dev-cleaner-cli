package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thanhdevapp/dev-cleaner/internal/scanner"
	"github.com/thanhdevapp/dev-cleaner/internal/tui"
	"github.com/thanhdevapp/dev-cleaner/internal/ui"
	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

var (
	scanIOS     bool
	scanAndroid bool
	scanNode    bool
	scanFlutter bool
	scanAll     bool
	scanTUI     bool
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan [flags]",
	Short: "Scan for development artifacts",
	Long: `Scan your system for development artifacts that can be cleaned.

By default, scans all supported categories and opens interactive TUI
for browsing, selection, and cleanup. The TUI provides tree navigation,
keyboard shortcuts, and real-time deletion progress.

Categories Scanned:
  • Xcode (DerivedData, Archives, CoreSimulator, CocoaPods)
  • Android (Gradle caches, SDK system images)
  • Node.js (node_modules, npm/yarn/pnpm/bun caches)
  • Flutter (build artifacts, .pub-cache, .dart_tool)

Examples:
  dev-cleaner scan                    # Scan all, launch TUI (default)
  dev-cleaner scan --ios              # Scan iOS/Xcode only
  dev-cleaner scan --android          # Scan Android only
  dev-cleaner scan --node             # Scan Node.js only
  dev-cleaner scan --flutter          # Scan Flutter only
  dev-cleaner scan --no-tui           # Text output without TUI

Flags:
  --ios             Scan iOS/Xcode artifacts only
  --android         Scan Android/Gradle artifacts only
  --node            Scan Node.js artifacts only
  --flutter         Scan Flutter/Dart artifacts only
  --no-tui, -T      Disable TUI, show simple text output
  --all             Scan all categories (default: true)

TUI Features:
  • Navigate with arrow keys or vim bindings (k/j/h/l)
  • Select items with Space, 'a' for all, 'n' for none
  • Quick clean single item with 'c'
  • Batch clean selected items with Enter
  • Drill down into folders with → or 'l'
  • Press '?' for detailed help`,
	Run: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().BoolVar(&scanIOS, "ios", false, "Scan iOS/Xcode artifacts only")
	scanCmd.Flags().BoolVar(&scanAndroid, "android", false, "Scan Android/Gradle artifacts only")
	scanCmd.Flags().BoolVar(&scanNode, "node", false, "Scan Node.js artifacts only")
	scanCmd.Flags().BoolVar(&scanFlutter, "flutter", false, "Scan Flutter/Dart artifacts only")
	scanCmd.Flags().BoolVar(&scanAll, "all", true, "Scan all categories (default)")
	scanCmd.Flags().BoolVar(&scanTUI, "tui", true, "Launch interactive TUI (default)")
	scanCmd.Flags().BoolP("no-tui", "T", false, "Disable TUI, show text output")
}

func runScan(cmd *cobra.Command, args []string) {
	s, err := scanner.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing scanner: %v\n", err)
		os.Exit(1)
	}

	// Determine scan options
	opts := types.ScanOptions{
		MaxDepth: 3,
	}

	// If any specific flag is set, use only those
	if scanIOS || scanAndroid || scanNode || scanFlutter {
		opts.IncludeXcode = scanIOS
		opts.IncludeAndroid = scanAndroid
		opts.IncludeNode = scanNode
		opts.IncludeFlutter = scanFlutter
	} else {
		// Default: scan all
		opts.IncludeXcode = true
		opts.IncludeAndroid = true
		opts.IncludeNode = true
		opts.IncludeFlutter = true
	}

	ui.PrintHeader("Scanning for development artifacts...")

	results, err := s.ScanAll(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println("\n  📭 No cleanable items found.")
		return
	}

	// Sort by size (largest first)
	sortBySize(results)

	// Check for --no-tui flag
	noTUI, _ := cmd.Flags().GetBool("no-tui")
	if noTUI {
		scanTUI = false
	}

	// Launch TUI by default
	if scanTUI {
		if err := tui.Run(results, false, Version); err != nil {
			fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Print results with enhanced UI
	ui.PrintResults(results)
	ui.PrintSummary(results)
	ui.PrintFooter()
}

// sortBySize sorts results by size in descending order
func sortBySize(results []types.ScanResult) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Size > results[i].Size {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}
