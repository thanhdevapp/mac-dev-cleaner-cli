package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected scanner, got nil")
	}
	if s.homeDir == "" {
		t.Error("homeDir should not be empty")
	}
}

func TestExpandPath(t *testing.T) {
	s, _ := New()

	tests := []struct {
		name     string
		input    string
		wantHome bool
	}{
		{"tilde path", "~/test", true},
		{"absolute path", "/tmp/test", false},
		{"relative path", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.ExpandPath(tt.input)
			if tt.wantHome {
				if result[0] != '/' {
					t.Errorf("expected absolute path, got: %s", result)
				}
				if result == tt.input {
					t.Errorf("tilde should be expanded: %s", result)
				}
			}
		})
	}
}

func TestPathExists(t *testing.T) {
	s, _ := New()

	// Create temp dir
	dir := t.TempDir()

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"existing dir", dir, true},
		{"non-existent", "/nonexistent/path/12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.PathExists(tt.path); got != tt.want {
				t.Errorf("PathExists(%s) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestCalculateSize(t *testing.T) {
	s, _ := New()

	tests := []struct {
		name      string
		setup     func(dir string)
		wantSize  int64
		wantCount int
	}{
		{
			name:      "empty directory",
			setup:     func(dir string) {},
			wantSize:  0,
			wantCount: 0,
		},
		{
			name: "single file",
			setup: func(dir string) {
				os.WriteFile(filepath.Join(dir, "test.txt"), make([]byte, 100), 0644)
			},
			wantSize:  100,
			wantCount: 1,
		},
		{
			name: "nested directories",
			setup: func(dir string) {
				subdir := filepath.Join(dir, "sub", "nested")
				os.MkdirAll(subdir, 0755)
				os.WriteFile(filepath.Join(dir, "a.txt"), make([]byte, 50), 0644)
				os.WriteFile(filepath.Join(subdir, "b.txt"), make([]byte, 150), 0644)
			},
			wantSize:  200,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(dir)

			size, count, err := s.calculateSize(dir)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if size != tt.wantSize {
				t.Errorf("size = %d, want %d", size, tt.wantSize)
			}
			if count != tt.wantCount {
				t.Errorf("count = %d, want %d", count, tt.wantCount)
			}
		})
	}
}
