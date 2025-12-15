package cleaner

import (
	"fmt"
	"os"
	"strings"
)

// dangerousPaths are system paths that should never be deleted
var dangerousPaths = []string{
	"/System",
	"/Library/System",
	"/usr",
	"/bin",
	"/sbin",
	"/etc",
	"/var",
	"/private",
	"/Applications",
	"/opt",
}

// protectedPatterns are patterns that should never be deleted
var protectedPatterns = []string{
	".ssh",
	".gnupg",
	".aws",
	".kube",
	"Keychain",
	"Keychains",
}

// ValidatePath checks if a path is safe to delete
func ValidatePath(path string) error {
	// Must be an absolute path
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must be absolute: %s", path)
	}

	// Check against dangerous system paths
	for _, dangerous := range dangerousPaths {
		if strings.HasPrefix(path, dangerous) {
			return fmt.Errorf("refusing to delete system path: %s", path)
		}
	}

	// Check for protected patterns
	for _, pattern := range protectedPatterns {
		if strings.Contains(path, pattern) {
			return fmt.Errorf("refusing to delete protected path containing '%s': %s", pattern, path)
		}
	}

	// Must be in home directory or known safe locations
	home := os.Getenv("HOME")
	if home == "" {
		return fmt.Errorf("HOME environment variable not set")
	}

	// Allow paths under home directory
	if strings.HasPrefix(path, home) {
		return nil
	}

	// Allow /tmp if needed (for testing)
	if strings.HasPrefix(path, "/tmp") {
		return nil
	}

	return fmt.Errorf("path outside home directory: %s", path)
}

// IsSafeToDelete is a convenience wrapper for ValidatePath
func IsSafeToDelete(path string) bool {
	return ValidatePath(path) == nil
}
