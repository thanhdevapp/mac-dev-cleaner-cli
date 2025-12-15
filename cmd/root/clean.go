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
	"github.com/thanhdevapp/dev-cleaner/internal/tui"
	"github.com/thanhdevapp/dev-cleaner/internal/ui"
	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

var (
	dryRun        bool
	confirmFlag   bool
	cleanIOS      bool
	cleanAndroid  bool
	cleanNode     bool
	cleanFlutter  bool
	cleanPython   bool
	cleanRust     bool
	cleanGo       bool
	cleanHomebrew bool
	cleanDocker   bool
	cleanJava     bool
	useTUI        bool
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean [flags]",
	Short: "Clean development artifacts",
	Long: `Interactively select and clean development artifacts.

By default, runs in TUI mode with interactive selection and dry-run
enabled (preview only). Use --confirm to actually delete files.

The TUI provides:
  • Real-time deletion progress with package-manager style output
  • Tree navigation for exploring nested folders
  • Quick single-item cleanup or batch operations
  • All operations logged to ~/.dev-cleaner.log

Safety Features:
  ✓ Dry-run mode by default (files are safe)
  ✓ Confirmation required before deletion
  ✓ Path validation (never touches system files)
  ✓ All actions logged for audit trail

Examples:
  dev-cleaner clean                   # Interactive TUI (dry-run)
  dev-cleaner clean --confirm         # Interactive TUI (actually delete)
  dev-cleaner clean --no-tui          # Simple text mode
  dev-cleaner clean --ios --confirm   # Clean iOS artifacts only
  dev-cleaner clean --node            # Preview Node.js cleanup (dry-run)

Flags:
  --confirm         Actually delete files (disables dry-run)
  --dry-run         Preview only, don't delete (default: true)
  --ios             Clean iOS/Xcode artifacts only
  --android         Clean Android/Gradle artifacts only
  --node            Clean Node.js artifacts only
  --flutter         Clean Flutter/Dart artifacts only
  --python          Clean Python caches
  --rust            Clean Rust/Cargo caches
  --go              Clean Go caches
  --homebrew        Clean Homebrew caches
  --docker          Clean Docker images, containers, volumes
  --java            Clean Maven/Gradle caches
  --no-tui, -T      Disable TUI, use simple text mode
  --tui             Use interactive TUI mode (default: true)

TUI Keyboard Shortcuts:
  c            Quick clean current item (ignores selections)
  Enter        Clean all selected items (batch mode)
  Space        Toggle selection
  a/n          Select all / none
  →/l          Enter tree mode (explore folders)
  ?            Show help screen

Important:
  • 'c' clears all selections and cleans ONLY the current item
  • 'Enter' cleans ALL selected items (batch operation)
  • Tree mode allows deletion at any folder level`,
	Run: runClean,
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().BoolVar(&dryRun, "dry-run", true, "Preview only, don't delete (default)")
	cleanCmd.Flags().BoolVar(&confirmFlag, "confirm", false, "Actually delete files")
	cleanCmd.Flags().BoolVar(&cleanIOS, "ios", false, "Clean iOS/Xcode artifacts only")
	cleanCmd.Flags().BoolVar(&cleanAndroid, "android", false, "Clean Android/Gradle artifacts only")
	cleanCmd.Flags().BoolVar(&cleanNode, "node", false, "Clean Node.js artifacts only")
	cleanCmd.Flags().BoolVar(&cleanFlutter, "flutter", false, "Clean Flutter/Dart artifacts only")
	cleanCmd.Flags().BoolVar(&cleanPython, "python", false, "Clean Python caches")
	cleanCmd.Flags().BoolVar(&cleanRust, "rust", false, "Clean Rust/Cargo caches")
	cleanCmd.Flags().BoolVar(&cleanGo, "go", false, "Clean Go caches")
	cleanCmd.Flags().BoolVar(&cleanHomebrew, "homebrew", false, "Clean Homebrew caches")
	cleanCmd.Flags().BoolVar(&cleanDocker, "docker", false, "Clean Docker images, containers, volumes")
	cleanCmd.Flags().BoolVar(&cleanJava, "java", false, "Clean Maven/Gradle caches")
	cleanCmd.Flags().BoolVar(&useTUI, "tui", true, "Use interactive TUI mode (default)")
	cleanCmd.Flags().BoolP("no-tui", "T", false, "Disable TUI, use simple text mode")
}

func runClean(cmd *cobra.Command, args []string) {
	// If --confirm is set, disable dry-run
	if confirmFlag {
		dryRun = false
	}

	// Check for --no-tui flag
	noTUI, _ := cmd.Flags().GetBool("no-tui")
	if noTUI {
		useTUI = false
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

	specificFlagSet := cleanIOS || cleanAndroid || cleanNode || cleanFlutter ||
		cleanPython || cleanRust || cleanGo || cleanHomebrew ||
		cleanDocker || cleanJava

	if specificFlagSet {
		opts.IncludeXcode = cleanIOS
		opts.IncludeAndroid = cleanAndroid
		opts.IncludeNode = cleanNode
		opts.IncludeFlutter = cleanFlutter
		opts.IncludePython = cleanPython
		opts.IncludeRust = cleanRust
		opts.IncludeGo = cleanGo
		opts.IncludeHomebrew = cleanHomebrew
		opts.IncludeDocker = cleanDocker
		opts.IncludeJava = cleanJava
	} else {
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

	// Sort by size
	sortBySize(results)

	// Use TUI or simple mode
	if useTUI {
		if err := tui.Run(results, dryRun, Version); err != nil {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			os.Exit(1)
		}
	} else {
		runSimpleMode(results)
	}
}

func runSimpleMode(results []types.ScanResult) {
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
