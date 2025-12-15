package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thanhdevapp/dev-cleaner/internal/cleaner"
	"github.com/thanhdevapp/dev-cleaner/internal/scanner"
	"github.com/thanhdevapp/dev-cleaner/internal/ui"
	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

var (
	dryRun       bool
	confirmFlag  bool
	cleanIOS     bool
	cleanAndroid bool
	cleanNode    bool
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean development artifacts",
	Long: `Interactively select and clean development artifacts.

By default, runs in dry-run mode (preview only).
Use --confirm to actually delete files.

Examples:
  dev-cleaner clean              # Interactive selection (dry-run)
  dev-cleaner clean --confirm    # Actually delete selected items
  dev-cleaner clean --ios        # Clean iOS artifacts only
  dev-cleaner clean --dry-run    # Preview what would be deleted`,
	Run: runClean,
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().BoolVar(&dryRun, "dry-run", true, "Preview only, don't delete (default)")
	cleanCmd.Flags().BoolVar(&confirmFlag, "confirm", false, "Actually delete files")
	cleanCmd.Flags().BoolVar(&cleanIOS, "ios", false, "Clean iOS/Xcode artifacts only")
	cleanCmd.Flags().BoolVar(&cleanAndroid, "android", false, "Clean Android/Gradle artifacts only")
	cleanCmd.Flags().BoolVar(&cleanNode, "node", false, "Clean Node.js artifacts only")
}

func runClean(cmd *cobra.Command, args []string) {
	// If --confirm is set, disable dry-run
	if confirmFlag {
		dryRun = false
	}

	s, err := scanner.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing scanner: %v\n", err)
		os.Exit(1)
	}

	// Determine scan options
	opts := types.ScanOptions{
		MaxDepth: 3,
	}

	if cleanIOS || cleanAndroid || cleanNode {
		opts.IncludeXcode = cleanIOS
		opts.IncludeAndroid = cleanAndroid
		opts.IncludeNode = cleanNode
	} else {
		opts.IncludeXcode = true
		opts.IncludeAndroid = true
		opts.IncludeNode = true
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

	// Sort by size
	sortBySize(results)

	// Print results with enhanced UI
	ui.PrintResults(results)
	ui.PrintSummary(results)

	// Interactive selection
	fmt.Println("\n📋 Enter item numbers to clean (comma-separated), 'all' for everything, or 'q' to quit:")
	fmt.Print("   > ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "q" || input == "quit" || input == "" {
		fmt.Println("Cancelled.")
		return
	}

	var selectedResults []types.ScanResult

	if input == "all" || input == "a" {
		selectedResults = results
	} else {
		// Parse comma-separated numbers
		parts := strings.Split(input, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 1 || idx > len(results) {
				fmt.Printf("Invalid selection: %s\n", part)
				continue
			}
			selectedResults = append(selectedResults, results[idx-1])
		}
	}

	if len(selectedResults) == 0 {
		fmt.Println("No valid items selected.")
		return
	}

	// Calculate total size
	var totalSize int64
	for _, r := range selectedResults {
		totalSize += r.Size
	}

	// Show warning
	if dryRun {
		ui.PrintDryRunWarning()
	} else {
		ui.PrintDeleteWarning(len(selectedResults), totalSize)
		fmt.Print("Type 'yes' to confirm: ")

		confirmInput, _ := reader.ReadString('\n')
		confirmInput = strings.TrimSpace(confirmInput)

		if confirmInput != "yes" {
			fmt.Println("Cancelled.")
			return
		}
	}

	// Perform cleaning
	c, err := cleaner.New(dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing cleaner: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	fmt.Println()
	cleanResults, err := c.Clean(selectedResults)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during cleaning: %v\n", err)
		os.Exit(1)
	}

	// Print results
	var successCount int
	var freedSpace int64
	for _, result := range cleanResults {
		if result.Success {
			successCount++
			freedSpace += result.Size
			if result.WasDryRun {
				fmt.Printf("  %s[DRY-RUN]%s Would delete: %s\n", ui.Yellow, ui.Reset, result.Path)
			} else {
				fmt.Printf("  %s✓%s Deleted: %s\n", ui.Green, ui.Reset, result.Path)
			}
		} else {
			fmt.Printf("  %s✗%s Failed: %s (%v)\n", ui.Red, ui.Reset, result.Path, result.Error)
		}
	}

	fmt.Printf("\n%sCompleted!%s %d items processed", ui.Bold, ui.Reset, successCount)
	if dryRun {
		fmt.Printf(" (would free %s)\n", ui.FormatSize(freedSpace))
	} else {
		fmt.Printf(" (%s freed)\n", ui.FormatSize(freedSpace))
	}
}
