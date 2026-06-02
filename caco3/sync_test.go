package caco3

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// makeBareRepo creates a bare git repository populated with one commit
// and returns its filesystem path. The returned path is suitable as the
// remote URL for git operations.
func makeBareRepo(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	work := filepath.Join(base, "work")
	bare := filepath.Join(base, "remote.git")

	if err := exec.Command("git", "init", "-q", "-b", "main", work).Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(work, "README"), []byte("hello\n"), 0644,
	); err != nil {
		t.Fatalf("write README: %v", err)
	}
	for _, args := range [][]string{
		{"config", "user.email", "caco3-test@example.com"},
		{"config", "user.name", "caco3-test"},
		{"add", "README"},
		{"commit", "-q", "-m", "initial"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = work
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	cmd := exec.Command("git", "clone", "-q", "--bare", work, bare)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git clone --bare: %v: %s", err, out)
	}
	return bare
}

func TestSyncRepos_skipsSelf(t *testing.T) {
	bare := makeBareRepo(t)
	root := t.TempDir()

	ws := multiLine(
		`repo {`,
		`    Name: "test.local/self/dockers",`,
		`}`,
		``,
		`repo_map {`,
		`    Src: {`,
		`        "test.local/self/dockers": "",`,
		`        "test.local/dep/dockers": "`+bare+`",`,
		`    },`,
		`}`,
	)
	if err := os.WriteFile(
		filepath.Join(root, "WORKSPACE.lets"), []byte(ws), 0644,
	); err != nil {
		t.Fatalf("write workspace: %v", err)
	}

	b, err := NewBuilder(root, &Config{Root: root})
	if err != nil {
		t.Fatalf("NewBuilder: %v", err)
	}
	if _, errs := b.ReadWorkspace(); errs != nil {
		for _, e := range errs {
			t.Error(e)
		}
		t.FailNow()
	}

	sums, err := b.SyncRepos(nil, &SyncOptions{})
	if err != nil {
		t.Fatalf("SyncRepos: %v", err)
	}

	// Self entry must not have been touched.
	if _, ok := sums.RepoCommits["test.local/self/dockers"]; ok {
		t.Error("self entry appears in sums; want skipped")
	}
	if ok, _ := isDir(filepath.Join(root, ".git")); ok {
		t.Error("root .git exists; self repo was touched")
	}
	if ok, _ := isDir(
		filepath.Join(root, "_/src/test.local/self/dockers"),
	); ok {
		t.Error("_/src/<self> exists; self repo was checked out under _/src")
	}

	// Dep should be fully cloned.
	if _, ok := sums.RepoCommits["test.local/dep/dockers"]; !ok {
		t.Fatalf("dep not in sums: %+v", sums.RepoCommits)
	}
	depDir := filepath.Join(root, "_/src/test.local/dep/dockers")
	if ok, _ := isDir(filepath.Join(depDir, ".git")); !ok {
		t.Errorf("dep .git missing at %s", depDir)
	}
	if ok, _ := isRegularFile(filepath.Join(depDir, "README")); !ok {
		t.Error("dep README missing")
	}
}

func TestSyncRepos_noDeps(t *testing.T) {
	root := t.TempDir()
	ws := multiLine(
		`repo {`,
		`    Name: "test.local/self/dockers",`,
		`}`,
	)
	if err := os.WriteFile(
		filepath.Join(root, "WORKSPACE.lets"), []byte(ws), 0644,
	); err != nil {
		t.Fatalf("write workspace: %v", err)
	}

	b, err := NewBuilder(root, &Config{Root: root})
	if err != nil {
		t.Fatalf("NewBuilder: %v", err)
	}
	if _, errs := b.ReadWorkspace(); errs != nil {
		for _, e := range errs {
			t.Error(e)
		}
		t.FailNow()
	}

	sums, err := b.SyncRepos(nil, &SyncOptions{})
	if err != nil {
		t.Fatalf("SyncRepos: %v", err)
	}
	if sums == nil {
		t.Fatal("sums is nil")
	}
	if len(sums.RepoCommits) != 0 {
		t.Errorf("got %d commits, want 0: %+v", len(sums.RepoCommits), sums.RepoCommits)
	}
	if ok, _ := isDir(filepath.Join(root, ".git")); ok {
		t.Error("root .git exists; self repo was touched")
	}
}
