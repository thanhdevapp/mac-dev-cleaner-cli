package types

import (
	"path/filepath"
)

// TreeNode represents a file/directory in hierarchical tree navigation
type TreeNode struct {
	Path      string          // Full path
	Name      string          // Display name (basename)
	Size      int64           // Size in bytes
	IsDir     bool            // Directory flag
	Type      CleanTargetType // xcode/android/node
	Children  []*TreeNode     // Child nodes (nil = not scanned)
	Scanned   bool            // Lazy scan flag
	Depth     int             // Current depth in tree
	FileCount int             // Number of files
}

// AddChild appends child to node's children
func (n *TreeNode) AddChild(child *TreeNode) {
	if n.Children == nil {
		n.Children = make([]*TreeNode, 0)
	}
	n.Children = append(n.Children, child)
}

// NeedsScanning returns true if node hasn't been scanned yet
func (n *TreeNode) NeedsScanning() bool {
	return !n.Scanned && n.IsDir
}

// HasChildren returns true if node has children
func (n *TreeNode) HasChildren() bool {
	return n.Children != nil && len(n.Children) > 0
}

// GetBasename returns the base name from path
func GetBasename(path string) string {
	return filepath.Base(path)
}

// ScanResultToTreeNode converts ScanResult to initial TreeNode
func ScanResultToTreeNode(result ScanResult) *TreeNode {
	return &TreeNode{
		Path:      result.Path,
		Name:      result.Name,
		Size:      result.Size,
		IsDir:     true, // Scan results are always directories
		Type:      result.Type,
		FileCount: result.FileCount,
		Scanned:   false,
		Depth:     0,
	}
}
