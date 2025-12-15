// Package ui provides terminal output formatting with lipgloss styling
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// Theme colors
var (
	// Primary colors
	primaryColor = lipgloss.Color("#7C3AED") // Purple
	successColor = lipgloss.Color("#10B981") // Green
	warningColor = lipgloss.Color("#F59E0B") // Amber
	dangerColor  = lipgloss.Color("#EF4444") // Red
	infoColor    = lipgloss.Color("#3B82F6") // Blue
	mutedColor   = lipgloss.Color("#6B7280") // Gray

	// Type colors
	xcodeColor   = lipgloss.Color("#147EFB") // Apple Blue
	androidColor = lipgloss.Color("#3DDC84") // Android Green
	nodeColor    = lipgloss.Color("#68A063") // Node Green
	flutterColor = lipgloss.Color("#02569B") // Flutter Blue
	cacheColor   = lipgloss.Color("#9CA3AF") // Gray
)

// Styles
var (
	// Header styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Padding(0, 2).
			MarginBottom(1)

	// Box styles
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	// Item styles
	indexStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Width(4)

	typeStyleBase = lipgloss.NewStyle().
			Bold(true).
			Width(10)

	sizeStyle = lipgloss.NewStyle().
			Width(10).
			Align(lipgloss.Right)

	nameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

	// Progress bar style
	barFilled = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor)

	barEmpty = lipgloss.NewStyle().
			Foreground(mutedColor).
			Background(lipgloss.Color("#374151"))

	// Summary styles
	summaryStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(successColor).
			MarginTop(1)

	// Warning styles
	warningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(dangerColor)

	dryRunStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(warningColor).
			Background(lipgloss.Color("#422006")).
			Padding(0, 1)

	// Footer style
	footerStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true).
			MarginTop(1)
)

// FormatSize formats bytes to human-readable format
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// PrintHeader prints a styled header
func PrintHeader(text string) {
	emoji := "🧹"
	if strings.Contains(text, "Scanning") {
		emoji = "🔍"
	}
	fmt.Println()
	fmt.Println(headerStyle.Render(fmt.Sprintf(" %s %s ", emoji, text)))
}

// getTypeStyle returns styled type badge
func getTypeStyle(t types.CleanTargetType) lipgloss.Style {
	style := typeStyleBase.Copy()
	switch t {
	case types.TypeXcode:
		return style.Foreground(xcodeColor)
	case types.TypeAndroid:
		return style.Foreground(androidColor)
	case types.TypeNode:
		return style.Foreground(nodeColor)
	case types.TypeFlutter:
		return style.Foreground(flutterColor)
	case types.TypeCache:
		return style.Foreground(cacheColor)
	default:
		return style
	}
}

// getSizeStyle returns styled size based on magnitude
func getSizeStyle(bytes int64) lipgloss.Style {
	style := sizeStyle.Copy()
	if bytes > 1024*1024*1024 { // > 1GB
		return style.Foreground(dangerColor).Bold(true)
	} else if bytes > 100*1024*1024 { // > 100MB
		return style.Foreground(warningColor)
	}
	return style.Foreground(successColor)
}

// renderProgressBar creates a visual progress bar
func renderProgressBar(current, max int64, width int) string {
	if max == 0 {
		return ""
	}

	percentage := float64(current) / float64(max)
	filled := int(percentage * float64(width))
	empty := width - filled

	if filled > width {
		filled = width
		empty = 0
	}

	bar := ""
	if filled > 0 {
		bar += barFilled.Render(strings.Repeat("█", filled))
	}
	if empty > 0 {
		bar += barEmpty.Render(strings.Repeat("░", empty))
	}

	return bar
}

// PrintResult prints a single scan result with enhanced formatting
func PrintResult(result types.ScanResult, index int, maxSize int64) {
	idx := indexStyle.Render(fmt.Sprintf("[%d]", index+1))
	typeStr := getTypeStyle(result.Type).Render(string(result.Type))
	sizeStr := getSizeStyle(result.Size).Render(FormatSize(result.Size))
	bar := renderProgressBar(result.Size, maxSize, 15)
	name := nameStyle.Render(result.Name)

	fmt.Printf("  %s %s %s %s  %s\n", idx, typeStr, sizeStr, bar, name)
}

// PrintResults prints all results in a styled box
func PrintResults(results []types.ScanResult) {
	if len(results) == 0 {
		fmt.Println("\n  📭 No cleanable items found.")
		return
	}

	// Find max size for progress bar scaling
	var maxSize int64
	for _, r := range results {
		if r.Size > maxSize {
			maxSize = r.Size
		}
	}

	// Print separator
	separator := lipgloss.NewStyle().
		Foreground(mutedColor).
		Render(strings.Repeat("─", 70))

	fmt.Println()
	fmt.Println(separator)
	fmt.Println()

	for i, result := range results {
		PrintResult(result, i, maxSize)
	}

	fmt.Println()
	fmt.Println(separator)
}

// PrintSummary prints the scan summary with enhanced styling
func PrintSummary(results []types.ScanResult) {
	var totalSize int64
	typeCounts := make(map[types.CleanTargetType]int)

	for _, r := range results {
		totalSize += r.Size
		typeCounts[r.Type]++
	}

	// Summary line
	summary := fmt.Sprintf("📊 Total: %d items  •  %s",
		len(results),
		FormatSize(totalSize),
	)
	fmt.Println(summaryStyle.Render(summary))

	// Type breakdown
	breakdown := ""
	if c := typeCounts[types.TypeXcode]; c > 0 {
		breakdown += getTypeStyle(types.TypeXcode).Render(fmt.Sprintf(" %d xcode", c))
	}
	if c := typeCounts[types.TypeAndroid]; c > 0 {
		breakdown += getTypeStyle(types.TypeAndroid).Render(fmt.Sprintf(" %d android", c))
	}
	if c := typeCounts[types.TypeNode]; c > 0 {
		breakdown += getTypeStyle(types.TypeNode).Render(fmt.Sprintf(" %d node", c))
	}
	if c := typeCounts[types.TypeFlutter]; c > 0 {
		breakdown += getTypeStyle(types.TypeFlutter).Render(fmt.Sprintf(" %d flutter", c))
	}
	if breakdown != "" {
		fmt.Println(lipgloss.NewStyle().Foreground(mutedColor).Render("   " + breakdown))
	}
}

// PrintDryRunWarning prints a dry-run mode notice
func PrintDryRunWarning() {
	warning := dryRunStyle.Render(" ⚡ DRY-RUN MODE ")
	msg := lipgloss.NewStyle().Foreground(mutedColor).Render(" No files will be deleted")
	fmt.Printf("\n%s%s\n", warning, msg)
	fmt.Println(footerStyle.Render("Use --confirm to actually delete files."))
}

// PrintDeleteWarning prints a deletion warning
func PrintDeleteWarning(count int, size int64) {
	msg := fmt.Sprintf("⚠️  WARNING: About to delete %d items (%s)", count, FormatSize(size))
	fmt.Println()
	fmt.Println(warningStyle.Render(msg))
}

// PrintFooter prints helpful footer message
func PrintFooter() {
	fmt.Println(footerStyle.Render("💡 Run 'dev-cleaner clean' to interactively select items to delete."))
}

// PrintSuccess prints a success message
func PrintSuccess(msg string) {
	style := lipgloss.NewStyle().Foreground(successColor)
	fmt.Println(style.Render("✓ " + msg))
}

// PrintError prints an error message
func PrintError(msg string) {
	style := lipgloss.NewStyle().Foreground(dangerColor)
	fmt.Println(style.Render("✗ " + msg))
}

// Deprecated colors for backward compatibility
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
)
