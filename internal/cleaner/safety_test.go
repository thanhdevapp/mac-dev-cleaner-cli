package cleaner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePath(t *testing.T) {
	home := os.Getenv("HOME")

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		// Valid paths
		{"home subdir", filepath.Join(home, "Library/test"), false},
		{"tmp path", "/tmp/test", false},

		// Invalid paths
		{"system path", "/System/Library", true},
		{"usr path", "/usr/bin", true},
		{"relative path", "relative/path", true},
		{"ssh dir", filepath.Join(home, ".ssh/keys"), true},
		{"outside home", "/opt/test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath(%s) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestIsSafeToDelete(t *testing.T) {
	home := os.Getenv("HOME")

	tests := []struct {
		path string
		want bool
	}{
		{filepath.Join(home, "Library/Developer/test"), true},
		{"/System/Library", false},
		{"/usr/bin", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsSafeToDelete(tt.path); got != tt.want {
				t.Errorf("IsSafeToDelete(%s) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
