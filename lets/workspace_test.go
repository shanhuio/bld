package lets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadWorkspace_legacy(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "WORKSPACE.lets")
	src := multiLine(
		`repo_map {`,
		`    Src: {`,
		`        "test.local/proj1/dockers": "git@example.com:p1.git",`,
		`    },`,
		`}`,
	)
	if err := os.WriteFile(f, []byte(src), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	ws, errs := readWorkspace(f)
	if errs != nil {
		t.Fatalf("readWorkspace: %v", errs)
	}
	if ws.Repo != nil {
		t.Errorf("Repo = %+v, want nil for legacy workspace", ws.Repo)
	}
	if ws.RepoMap == nil {
		t.Fatal("RepoMap = nil")
	}
	if got := ws.RepoMap.Src["test.local/proj1/dockers"]; got != "git@example.com:p1.git" {
		t.Errorf("Src entry = %q, want git@example.com:p1.git", got)
	}
}

func TestReadWorkspace_withRepo(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "WORKSPACE.lets")
	src := multiLine(
		`repo {`,
		`    Name: "test.local/proj2/dockers",`,
		`}`,
		``,
		`repo_map {`,
		`    Src: {`,
		`        "test.local/proj1/dockers": "git@example.com:p1.git",`,
		`    },`,
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
	if ws.RepoMap == nil {
		t.Fatal("RepoMap = nil")
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
		filepath.Join(root, "WORKSPACE.lets"), []byte(src), 0644,
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
	f := filepath.Join(dir, "WORKSPACE.lets")
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
	if ws.RepoMap != nil {
		t.Errorf("RepoMap = %+v, want nil", ws.RepoMap)
	}
}
