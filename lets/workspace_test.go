package lets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadWorkspace_noRepo(t *testing.T) {
	root := t.TempDir()
	src := multiLine(
		`bundle {`,
		`    Name: "all",`,
		`}`,
	)
	if err := os.WriteFile(
		filepath.Join(root, buildFileName), []byte(src), 0644,
	); err != nil {
		t.Fatalf("write: %v", err)
	}

	b, err := NewBuilder(root, &Config{Root: root})
	if err != nil {
		t.Fatalf("NewBuilder: %v", err)
	}
	if _, errs := b.ReadWorkspace(); errs == nil {
		t.Fatal("ReadWorkspace: want error for workspace without repo node, got nil")
	}
}

func TestReadWorkspace_repoNotFirst(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, buildFileName)
	src := multiLine(
		`bundle {`,
		`    Name: "all",`,
		`}`,
		``,
		`repo {`,
		`    Name: "test.local/proj/dockers",`,
		`}`,
	)
	if err := os.WriteFile(f, []byte(src), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, errs := readWorkspace(f); errs == nil {
		t.Fatal("readWorkspace: want error for repo not first, got nil")
	}
}

func TestReadWorkspace_withDeps(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, buildFileName)
	src := multiLine(
		`repo {`,
		`    Name: "test.local/proj2/dockers",`,
		`    Deps: {`,
		`        "test.local/proj1/dockers": "git@example.com:p1.git",`,
		`    },`,
		`}`,
		``,
		`bundle {`,
		`    Name: "all",`,
		`}`,
	)
	if err := os.WriteFile(f, []byte(src), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	ws, errs := readWorkspace(f)
	if errs != nil {
		t.Fatalf("readWorkspace: %v", errs)
	}
	if ws.Repo == nil {
		t.Fatal("Repo = nil, want non-nil")
	}
	if got, want := ws.Repo.Name, "test.local/proj2/dockers"; got != want {
		t.Errorf("Repo.Name = %q, want %q", got, want)
	}
	got := ws.Repo.Deps["test.local/proj1/dockers"]
	if want := "git@example.com:p1.git"; got != want {
		t.Errorf("Deps entry = %q, want %q", got, want)
	}
}

func TestReadWorkspace_emptyRepoName(t *testing.T) {
	root := t.TempDir()
	src := multiLine(
		`repo {`,
		`    Name: "",`,
		`}`,
	)
	if err := os.WriteFile(
		filepath.Join(root, buildFileName), []byte(src), 0644,
	); err != nil {
		t.Fatalf("write: %v", err)
	}

	b, err := NewBuilder(root, &Config{Root: root})
	if err != nil {
		t.Fatalf("NewBuilder: %v", err)
	}
	_, errs := b.ReadWorkspace()
	if errs == nil {
		t.Fatal("ReadWorkspace: want error for empty repo.Name, got nil")
	}
}

func TestReadWorkspace_repoOnly(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, buildFileName)
	src := multiLine(
		`repo {`,
		`    Name: "test.local/standalone",`,
		`}`,
	)
	if err := os.WriteFile(f, []byte(src), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	ws, errs := readWorkspace(f)
	if errs != nil {
		t.Fatalf("readWorkspace: %v", errs)
	}
	if ws.Repo == nil || ws.Repo.Name != "test.local/standalone" {
		t.Errorf("Repo = %+v, want Name=test.local/standalone", ws.Repo)
	}
	if ws.Repo.Deps != nil {
		t.Errorf("Deps = %+v, want nil", ws.Repo.Deps)
	}
}

func TestFindRoot_letsRootStamp(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, letsRootFile), "")

	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	got, err := findRoot(sub)
	if err != nil {
		t.Fatalf("findRoot: %v", err)
	}
	if got != root {
		t.Errorf("findRoot(%q) = %q, want %q", sub, got, root)
	}
}

func TestFindRoot_gitDir(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}

	sub := filepath.Join(root, "pkg")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	got, err := findRoot(sub)
	if err != nil {
		t.Fatalf("findRoot: %v", err)
	}
	if got != root {
		t.Errorf("findRoot(%q) = %q, want %q", sub, got, root)
	}
}

func TestSaveRepoSums_empty_removesExistingFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "sums.jsonx")
	if err := os.WriteFile(f, []byte("{ RepoCommits: {} }\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := SaveRepoSums(f, &RepoSums{RepoCommits: map[string]string{}}); err != nil {
		t.Fatalf("SaveRepoSums: %v", err)
	}
	if ok, _ := isRegularFile(f); ok {
		t.Error("sums file still exists after saving empty sums")
	}
}

func TestSaveRepoSums_empty_writesNoFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "sums.jsonx")
	if err := SaveRepoSums(f, &RepoSums{}); err != nil {
		t.Fatalf("SaveRepoSums: %v", err)
	}
	if ok, _ := isRegularFile(f); ok {
		t.Error("sums file created for empty sums")
	}
}

func TestSaveRepoSums_nonEmpty_writes(t *testing.T) {
	f := filepath.Join(t.TempDir(), "sums.jsonx")
	want := &RepoSums{RepoCommits: map[string]string{"a/b": "deadbeef"}}
	if err := SaveRepoSums(f, want); err != nil {
		t.Fatalf("SaveRepoSums: %v", err)
	}
	got, err := ReadRepoSums(f)
	if err != nil {
		t.Fatalf("ReadRepoSums: %v", err)
	}
	if got.RepoCommits["a/b"] != "deadbeef" {
		t.Errorf("RepoCommits = %+v, want a/b=deadbeef", got.RepoCommits)
	}
}
