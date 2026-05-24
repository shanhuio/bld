package caco3

import "testing"

func TestMakeRelPath(t *testing.T) {
	for _, c := range []struct {
		p, f, want string
	}{
		{"pkg", "a.go", "pkg/a.go"},
		{"pkg/sub", "a.go", "pkg/sub/a.go"},
		{"", "a.go", "a.go"},
		{"pkg", "", "pkg"},
		{"", "", ""},
		{"pkg", "./a.go", "pkg/a.go"},
		{"pkg", "sub/a.go", "pkg/sub/a.go"},
		// Escapes are clamped: f is cleaned from / first, so .. cannot escape p.
		{"pkg", "../escape", "pkg/escape"},
		{"pkg/sub", "../../escape", "pkg/sub/escape"},
		// An absolute f gets re-rooted under p.
		{"pkg", "/abs/path", "pkg/abs/path"},
		// Trailing slashes get cleaned away.
		{"pkg", "sub/", "pkg/sub"},
	} {
		got := makeRelPath(c.p, c.f)
		if got != c.want {
			t.Errorf("makeRelPath(%q, %q) = %q, want %q", c.p, c.f, got, c.want)
		}
	}
}

func TestMakePath(t *testing.T) {
	for _, c := range []struct {
		p, f, want string
	}{
		// Relative f delegates to makeRelPath.
		{"pkg", "a.go", "pkg/a.go"},
		{"pkg", "../escape", "pkg/escape"},
		{"", "f", "f"},

		// Absolute f bypasses p.
		{"pkg", "/abs/file", "abs/file"},
		{"pkg/sub", "/abs", "abs"},
		{"", "/abs", "abs"},
		// Absolute path gets cleaned (dots collapsed).
		{"pkg", "/a/./b/../c", "a/c"},
	} {
		got := makePath(c.p, c.f)
		if got != c.want {
			t.Errorf("makePath(%q, %q) = %q, want %q", c.p, c.f, got, c.want)
		}
	}
}
