package main

import (
	"os"
	"testing"

	"shanhu.io/bld/lets"
)

func TestResolveWorkDir_withRoot(t *testing.T) {
	root := t.TempDir() // already absolute and clean
	config := &lets.Config{Root: root}

	wd, err := resolveWorkDir(config)
	if err != nil {
		t.Fatalf("resolveWorkDir: %v", err)
	}
	if wd != root {
		t.Errorf("work dir = %q, want root %q", wd, root)
	}
	if config.Root != root {
		t.Errorf("config.Root = %q, want %q", config.Root, root)
	}
}

func TestResolveWorkDir_noRoot(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	config := &lets.Config{}

	wd, err := resolveWorkDir(config)
	if err != nil {
		t.Fatalf("resolveWorkDir: %v", err)
	}
	if wd != cwd {
		t.Errorf("work dir = %q, want cwd %q", wd, cwd)
	}
	if config.Root != "" {
		t.Errorf("config.Root = %q, want empty", config.Root)
	}
}
