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
	scanIOS         bool
	scanAndroid     bool
	scanNode        bool
	scanReactNative bool
	scanAll         bool
	scanTUI         bool
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for development artifacts",
	Long: `Scan your system for development artifacts that can be cleaned.

By default, opens interactive TUI for selection.
Use --no-tui for simple text output.

Examples:
  dev-cleaner scan              # Scan + TUI (default)
  dev-cleaner scan --no-tui     # Scan + text output
  dev-cleaner scan --ios        # Scan iOS/Xcode only
  dev-cleaner scan --rn         # Scan React Native caches
  dev-cleaner scan --rn --ios   # Combine flags for RN projects`,
	Run: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().BoolVar(&scanIOS, "ios", false, "Scan iOS/Xcode artifacts only")
	scanCmd.Flags().BoolVar(&scanAndroid, "android", false, "Scan Android/Gradle artifacts only")
	scanCmd.Flags().BoolVar(&scanNode, "node", false, "Scan Node.js artifacts only")
	scanCmd.Flags().BoolVar(&scanReactNative, "react-native", false, "Scan React Native caches")
	scanCmd.Flags().BoolVar(&scanReactNative, "rn", false, "Alias for --react-native")
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
	if scanIOS || scanAndroid || scanNode || scanReactNative {
		opts.IncludeXcode = scanIOS
		opts.IncludeAndroid = scanAndroid
		opts.IncludeNode = scanNode
		opts.IncludeReactNative = scanReactNative
	} else {
		// Default: scan all
		opts.IncludeXcode = true
		opts.IncludeAndroid = true
		opts.IncludeNode = true
		opts.IncludeReactNative = true
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
		if err := tui.Run(results, false); err != nil {
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
