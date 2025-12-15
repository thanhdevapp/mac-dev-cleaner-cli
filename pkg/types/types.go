// Package types contains shared types for the dev-cleaner CLI
package types

// CleanTargetType represents the category of the clean target
type CleanTargetType string

const (
	TypeXcode       CleanTargetType = "xcode"
	TypeAndroid     CleanTargetType = "android"
	TypeNode        CleanTargetType = "node"
	TypeReactNative CleanTargetType = "react-native"
	TypeCache       CleanTargetType = "cache"
)

// ScanResult represents a single scannable/cleanable directory
type ScanResult struct {
	Path      string          `json:"path"`
	Type      CleanTargetType `json:"type"`
	Size      int64           `json:"size"`
	FileCount int             `json:"fileCount"`
	Name      string          `json:"name"` // Display name
}

// ScanOptions controls scanning behavior
type ScanOptions struct {
	IncludeXcode       bool
	IncludeAndroid     bool
	IncludeNode        bool
	IncludeReactNative bool
	IncludeCache       bool
	MaxDepth           int
	ProjectRoot        string // Optional: scan from specific root
}

// CleanOptions controls cleaning behavior
type CleanOptions struct {
	DryRun  bool
	Confirm bool
	LogPath string
}

// DefaultScanOptions returns options with all categories enabled
func DefaultScanOptions() ScanOptions {
	return ScanOptions{
		IncludeXcode:       true,
		IncludeAndroid:     true,
		IncludeNode:        true,
		IncludeReactNative: true,
		IncludeCache:       true,
		MaxDepth:           3,
	}
}
