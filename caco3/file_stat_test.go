package caco3

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func newTestEnv(t *testing.T) (*env, string, string) {
	t.Helper()
	root := t.TempDir()
	srcDir := filepath.Join(root, "src")
	outDir := filepath.Join(root, "out")
	for _, d := range []string{srcDir, outDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	return &env{srcDir: srcDir, outDir: outDir}, srcDir, outDir
}

func TestNewFileStatSrc(t *testing.T) {
	e, srcDir, _ := newTestEnv(t)
	content := []byte("hello")
	if err := os.WriteFile(filepath.Join(srcDir, "f.txt"), content, 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	stat, err := newSrcFileStat(e, "f.txt")
	if err != nil {
		t.Fatalf("newSrcFileStat: %v", err)
	}
	if stat.Name != "f.txt" {
		t.Errorf("Name = %q, want %q", stat.Name, "f.txt")
	}
	if stat.Type != fileTypeSrc {
		t.Errorf("Type = %q, want %q", stat.Type, fileTypeSrc)
	}
	if stat.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", stat.Size, len(content))
	}
	if stat.ModTimestamp == 0 {
		t.Error("ModTimestamp = 0, want nonzero")
	}
	if stat.Symlink != "" {
		t.Errorf("Symlink = %q, want empty", stat.Symlink)
	}
}

func TestNewFileStatOut(t *testing.T) {
	e, _, outDir := newTestEnv(t)
	if err := os.WriteFile(filepath.Join(outDir, "o.bin"), []byte("xy"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	stat, err := newOutFileStat(e, "o.bin")
	if err != nil {
		t.Fatalf("newOutFileStat: %v", err)
	}
	if stat.Type != fileTypeOut {
		t.Errorf("Type = %q, want %q", stat.Type, fileTypeOut)
	}
}

func TestNewFileStatMissing(t *testing.T) {
	e, _, _ := newTestEnv(t)
	_, err := newSrcFileStat(e, "nope.txt")
	if err == nil {
		t.Fatal("want error, got nil")
	}
	if !errors.Is(err, errNotFound) {
		t.Errorf("err = %v, want errors.Is(..., errNotFound)", err)
	}
}

func TestNewFileStatSymlink(t *testing.T) {
	e, srcDir, _ := newTestEnv(t)
	target := filepath.Join(srcDir, "target.txt")
	if err := os.WriteFile(target, []byte("data"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	link := filepath.Join(srcDir, "link")
	if err := os.Symlink("target.txt", link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	stat, err := newSrcFileStat(e, "link")
	if err != nil {
		t.Fatalf("newSrcFileStat: %v", err)
	}
	if stat.Symlink != "target.txt" {
		t.Errorf("Symlink = %q, want %q", stat.Symlink, "target.txt")
	}
}

func TestSameFileStatUnchanged(t *testing.T) {
	e, srcDir, _ := newTestEnv(t)
	if err := os.WriteFile(filepath.Join(srcDir, "a"), []byte("x"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	stat, err := newSrcFileStat(e, "a")
	if err != nil {
		t.Fatalf("setup stat: %v", err)
	}
	same, err := sameFileStat(e, stat)
	if err != nil {
		t.Fatalf("sameFileStat: %v", err)
	}
	if !same {
		t.Error("same = false, want true for unchanged file")
	}
}

func TestSameFileStatSizeChanged(t *testing.T) {
	e, srcDir, _ := newTestEnv(t)
	f := filepath.Join(srcDir, "a")
	if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	stat, err := newSrcFileStat(e, "a")
	if err != nil {
		t.Fatalf("setup stat: %v", err)
	}
	if err := os.WriteFile(f, []byte("xyz"), 0644); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	same, err := sameFileStat(e, stat)
	if err != nil {
		t.Fatalf("sameFileStat: %v", err)
	}
	if same {
		t.Error("same = true, want false after size change")
	}
}

func TestSameFileStatDeleted(t *testing.T) {
	e, srcDir, _ := newTestEnv(t)
	f := filepath.Join(srcDir, "a")
	if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	stat, err := newSrcFileStat(e, "a")
	if err != nil {
		t.Fatalf("setup stat: %v", err)
	}
	if err := os.Remove(f); err != nil {
		t.Fatalf("remove: %v", err)
	}
	same, err := sameFileStat(e, stat)
	if err != nil {
		t.Fatalf("sameFileStat: %v", err)
	}
	if same {
		t.Error("same = true, want false after file deletion")
	}
}
