package caco3

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsRegularFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("hi"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	for _, c := range []struct {
		name string
		path string
		want bool
	}{
		{"regular file", file, true},
		{"directory", dir, false},
		{"missing", filepath.Join(dir, "nope"), false},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := isRegularFile(c.path)
			if err != nil {
				t.Fatalf("isRegularFile: %v", err)
			}
			if got != c.want {
				t.Errorf("isRegularFile(%q) = %v, want %v", c.path, got, c.want)
			}
		})
	}
}

func TestIsDir(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("hi"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	for _, c := range []struct {
		name string
		path string
		want bool
	}{
		{"directory", dir, true},
		{"regular file", file, false},
		{"missing", filepath.Join(dir, "nope"), false},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := isDir(c.path)
			if err != nil {
				t.Fatalf("isDir: %v", err)
			}
			if got != c.want {
				t.Errorf("isDir(%q) = %v, want %v", c.path, got, c.want)
			}
		})
	}
}
