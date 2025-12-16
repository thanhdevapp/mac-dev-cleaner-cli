// Package types contains shared types for the dev-cleaner CLI
package types

// CleanTargetType represents the category of the clean target
type CleanTargetType string

const (
	TypeXcode       CleanTargetType = "xcode"
	TypeAndroid     CleanTargetType = "android"
	TypeNode        CleanTargetType = "node"
	TypeReactNative CleanTargetType = "react-native"
	TypeFlutter     CleanTargetType = "flutter"
	TypeCache       CleanTargetType = "cache"
	TypePython      CleanTargetType = "python"
	TypeRust        CleanTargetType = "rust"
	TypeGo          CleanTargetType = "go"
	TypeHomebrew    CleanTargetType = "homebrew"
	TypeDocker      CleanTargetType = "docker"
	TypeJava        CleanTargetType = "java"
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
	IncludeFlutter     bool
	IncludeCache       bool
	IncludePython      bool
	IncludeRust        bool
	IncludeGo          bool
	IncludeHomebrew    bool
	IncludeDocker      bool
	IncludeJava        bool
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
		IncludeFlutter:     true,
		IncludeCache:       true,
		IncludePython:      true,
		IncludeRust:        true,
		IncludeGo:          true,
		IncludeHomebrew:    true,
		IncludeDocker:      true,
		IncludeJava:        true,
		MaxDepth:           3,
	}
}
