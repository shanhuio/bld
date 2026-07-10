package gofiledag

import (
	"go/token"
	"testing"
)

func TestRelPath(t *testing.T) {
	if got := relPath("/a/b/c.go", ""); got != "/a/b/c.go" {
		t.Errorf("relPath with empty cwd = %q, want unchanged", got)
	}
	if got := relPath("/a/b/c.go", "/a/b"); got != "c.go" {
		t.Errorf("relPath = %q, want c.go", got)
	}
}

func TestRelPos(t *testing.T) {
	pos := token.Position{Filename: "/a/b/c.go", Line: 3, Column: 5}
	if got, want := relPos(pos, "/a/b"), "c.go:3:5"; got != want {
		t.Errorf("relPos = %q, want %q", got, want)
	}
}
