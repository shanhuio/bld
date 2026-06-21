package lets

import (
	"path/filepath"
	"testing"
)

func TestNewBuilder_workDirIsRoot(t *testing.T) {
	root := t.TempDir()
	if _, err := NewBuilder(root, &Config{Root: root}); err != nil {
		t.Fatalf("NewBuilder: %v", err)
	}
}

func TestNewBuilder_workDirUnderRoot(t *testing.T) {
	root := t.TempDir()
	// A subdirectory under root, including a dependency-style path under
	// _/src, must be accepted.
	sub := filepath.Join(root, "_", "src", "dep.example.com", "lib")
	if _, err := NewBuilder(sub, &Config{Root: root}); err != nil {
		t.Fatalf("NewBuilder: %v", err)
	}
}

func TestNewBuilder_workDirOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir() // sibling temp dir, not under root
	if _, err := NewBuilder(outside, &Config{Root: root}); err == nil {
		t.Fatal("NewBuilder: want error for work dir outside root, got nil")
	}
}
