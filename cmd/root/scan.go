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
	scanIOS      bool
	scanAndroid  bool
	scanNode     bool
	scanFlutter  bool
	scanPython   bool
	scanRust     bool
	scanGo       bool
	scanHomebrew bool
	scanDocker   bool
	scanJava     bool
	scanAll      bool
	scanTUI      bool
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
  • Python (pip/poetry/uv caches, venv, __pycache__)
  • Rust (Cargo registry/git, target directories)
  • Go (build cache, module cache)
  • Homebrew (download caches)
  • Docker (unused images, containers, volumes, build cache)
  • Java/Kotlin (Maven .m2, Gradle caches, build directories)

Examples:
  dev-cleaner scan                    # Scan all, launch TUI (default)
  dev-cleaner scan --ios              # Scan iOS/Xcode only
  dev-cleaner scan --android          # Scan Android only
  dev-cleaner scan --node             # Scan Node.js only
  dev-cleaner scan --flutter          # Scan Flutter only
  dev-cleaner scan --python           # Scan Python only
  dev-cleaner scan --rust             # Scan Rust/Cargo only
  dev-cleaner scan --go               # Scan Go only
  dev-cleaner scan --homebrew         # Scan Homebrew only
  dev-cleaner scan --docker           # Scan Docker only
  dev-cleaner scan --java             # Scan Java/Maven/Gradle only
  dev-cleaner scan --no-tui           # Text output without TUI

Flags:
  --ios             Scan iOS/Xcode artifacts only
  --android         Scan Android/Gradle artifacts only
  --node            Scan Node.js artifacts only
  --flutter         Scan Flutter/Dart artifacts only
  --python          Scan Python caches and virtualenvs
  --rust            Scan Rust/Cargo caches and targets
  --go              Scan Go build and module caches
  --homebrew        Scan Homebrew caches
  --docker          Scan Docker images, containers, volumes
  --java            Scan Maven/Gradle caches and build dirs
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
	scanCmd.Flags().BoolVar(&scanPython, "python", false, "Scan Python caches (pip, poetry, venv, __pycache__)")
	scanCmd.Flags().BoolVar(&scanRust, "rust", false, "Scan Rust/Cargo caches and target directories")
	scanCmd.Flags().BoolVar(&scanGo, "go", false, "Scan Go build and module caches")
	scanCmd.Flags().BoolVar(&scanHomebrew, "homebrew", false, "Scan Homebrew caches")
	scanCmd.Flags().BoolVar(&scanDocker, "docker", false, "Scan Docker images, containers, volumes")
	scanCmd.Flags().BoolVar(&scanJava, "java", false, "Scan Maven/Gradle caches and build dirs")
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
	specificFlagSet := scanIOS || scanAndroid || scanNode || scanFlutter ||
		scanPython || scanRust || scanGo || scanHomebrew ||
		scanDocker || scanJava

	if specificFlagSet {
		opts.IncludeXcode = scanIOS
		opts.IncludeAndroid = scanAndroid
		opts.IncludeNode = scanNode
		opts.IncludeFlutter = scanFlutter
		opts.IncludePython = scanPython
		opts.IncludeRust = scanRust
		opts.IncludeGo = scanGo
		opts.IncludeHomebrew = scanHomebrew
		opts.IncludeDocker = scanDocker
		opts.IncludeJava = scanJava
	} else {
		// Default: scan all
		opts = types.DefaultScanOptions()
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
