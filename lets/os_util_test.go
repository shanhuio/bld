package lets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathUnder(t *testing.T) {
	sep := string(filepath.Separator)
	base := filepath.Join("/", "a", "b")
	inside := filepath.Join(base, "c")
	deeper := filepath.Join(base, "c", "d")
	sibling := filepath.Join("/", "a", "bxyz")
	parent := filepath.Join("/", "a")
	unrelated := filepath.Join("/", "x")

	for _, c := range []struct {
		name     string
		base     string
		physical string
		wantRel  string
		wantOK   bool
	}{
		{"empty base", "", inside, "", false},
		{"equal paths", base, base, "", true},
		{"one level under", base, inside, "c", true},
		{"two levels under", base, deeper, "c" + sep + "d", true},
		{"boundary not crossed", base, sibling, "", false},
		{"physical is parent", base, parent, "", false},
		{"unrelated", base, unrelated, "", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			gotRel, gotOK := pathUnder(c.base, c.physical)
			if gotRel != c.wantRel || gotOK != c.wantOK {
				t.Errorf("pathUnder(%q, %q) = (%q, %v), want (%q, %v)",
					c.base, c.physical, gotRel, gotOK, c.wantRel, c.wantOK)
			}
		})
	}
}

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

func TestPathExists(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("hi"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(file, link); err != nil {
		t.Fatalf("setup symlink: %v", err)
	}

	for _, c := range []struct {
		name string
		path string
		want bool
	}{
		{"directory", dir, true},
		{"regular file", file, true},
		{"symlink", link, true},
		{"missing", filepath.Join(dir, "nope"), false},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := pathExists(c.path)
			if err != nil {
				t.Fatalf("pathExists: %v", err)
			}
			if got != c.want {
				t.Errorf("pathExists(%q) = %v, want %v", c.path, got, c.want)
			}
		})
	}
}
