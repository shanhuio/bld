package lets

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		{"config", "user.email", "lets-test@example.com"},
		{"config", "user.name", "lets-test"},
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
		`    Deps: {`,
		`        "test.local/self/dockers": "",`,
		`        "test.local/dep/dockers": "`+bare+`",`,
		`    },`,
		`}`,
	)
	if err := os.WriteFile(
		filepath.Join(root, buildFileName), []byte(ws), 0644,
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

	sums, err := b.SyncRepos(nil)
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
		filepath.Join(root, buildFileName), []byte(ws), 0644,
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

	sums, err := b.SyncRepos(nil)
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

// TestGitSync_shallowUpdate checks that gitSync creates a minimal-depth
// (shallow) checkout and can still advance it to a newer commit, which the
// old fetch-HEAD-then-merge flow could not do across shallow grafts.
func TestGitSync_shallowUpdate(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found")
	}

	base := t.TempDir()
	bare := filepath.Join(base, "remote.git")
	work := filepath.Join(base, "work")

	git := func(dir string, args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}

	// A bare remote plus a work clone to push commits from.
	git(base, "init", "-q", "--bare", "-b", "main", bare)
	git(base, "clone", "-q", bare, work)
	git(work, "config", "user.email", "lets-test@example.com")
	git(work, "config", "user.name", "lets-test")
	writeFile(t, filepath.Join(work, "README"), "v1\n")
	git(work, "add", "README")
	git(work, "commit", "-q", "-m", "v1")
	git(work, "push", "-q", "origin", "main")
	commit1 := git(work, "rev-parse", "HEAD")

	// First sync clones the dependency into a fresh dir.
	dir := filepath.Join(base, "checkout")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	res, err := gitSync("dep", dir, bare, commit1)
	if err != nil {
		t.Fatalf("gitSync 1: %v", err)
	}
	if res.commit != commit1 {
		t.Fatalf("sync 1 commit = %q, want %q", res.commit, commit1)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "README")); string(got) != "v1\n" {
		t.Errorf("README = %q, want v1", got)
	}
	if git(dir, "rev-parse", "--is-shallow-repository") != "true" {
		t.Error("checkout is not shallow after first sync")
	}

	// Push a newer commit, then sync again: the checkout must advance.
	writeFile(t, filepath.Join(work, "README"), "v2\n")
	git(work, "add", "README")
	git(work, "commit", "-q", "-m", "v2")
	git(work, "push", "-q", "origin", "main")
	commit2 := git(work, "rev-parse", "HEAD")

	res, err = gitSync("dep", dir, bare, commit2)
	if err != nil {
		t.Fatalf("gitSync 2: %v", err)
	}
	if res.commit != commit2 {
		t.Fatalf("sync 2 commit = %q, want %q", res.commit, commit2)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "README")); string(got) != "v2\n" {
		t.Errorf("README after update = %q, want v2", got)
	}
	if head := git(dir, "rev-parse", "HEAD"); head != commit2 {
		t.Errorf("HEAD = %q, want %q", head, commit2)
	}
	if git(dir, "rev-parse", "--is-shallow-repository") != "true" {
		t.Error("checkout is not shallow after update")
	}
}
