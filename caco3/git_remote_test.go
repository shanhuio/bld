package caco3

import (
	"os/exec"
	"testing"
)

func gitInitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command("git", "init", "-q", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	return dir
}

func gitAddRemote(t *testing.T, dir, name, url string) {
	t.Helper()
	cmd := exec.Command("git", "remote", "add", name, url)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git remote add %s %s: %v: %s", name, url, err, out)
	}
}

func TestListRemotes_empty(t *testing.T) {
	dir := gitInitRepo(t)
	remotes, err := listRemotes(dir)
	if err != nil {
		t.Fatalf("listRemotes: %v", err)
	}
	if len(remotes) != 0 {
		t.Errorf("got %d remotes, want 0: %+v", len(remotes), remotes)
	}
}

func TestListRemotes_single(t *testing.T) {
	dir := gitInitRepo(t)
	gitAddRemote(t, dir, "origin", "git@example.com:foo/bar.git")

	remotes, err := listRemotes(dir)
	if err != nil {
		t.Fatalf("listRemotes: %v", err)
	}
	if len(remotes) != 1 {
		t.Fatalf("got %d remotes, want 1: %+v", len(remotes), remotes)
	}
	r, ok := remotes["origin"]
	if !ok {
		t.Fatalf("origin not present: %+v", remotes)
	}
	if r.name != "origin" {
		t.Errorf("name = %q, want origin", r.name)
	}
	if want := "git@example.com:foo/bar.git"; r.git != want {
		t.Errorf("git = %q, want %q", r.git, want)
	}
	if !r.fetch {
		t.Error("fetch = false, want true")
	}
	if !r.push {
		t.Error("push = false, want true")
	}
}

func TestListRemotes_multiple(t *testing.T) {
	dir := gitInitRepo(t)
	gitAddRemote(t, dir, "origin", "git@example.com:foo/bar.git")
	gitAddRemote(t, dir, "github", "git@github.com:foo/bar.git")

	remotes, err := listRemotes(dir)
	if err != nil {
		t.Fatalf("listRemotes: %v", err)
	}
	if len(remotes) != 2 {
		t.Fatalf("got %d remotes, want 2: %+v", len(remotes), remotes)
	}
	for _, name := range []string{"origin", "github"} {
		r, ok := remotes[name]
		if !ok {
			t.Errorf("%q not present", name)
			continue
		}
		if !r.fetch || !r.push {
			t.Errorf("%q: fetch=%v push=%v, want both true",
				name, r.fetch, r.push)
		}
	}
	if remotes["origin"].git != "git@example.com:foo/bar.git" {
		t.Errorf("origin url = %q", remotes["origin"].git)
	}
	if remotes["github"].git != "git@github.com:foo/bar.git" {
		t.Errorf("github url = %q", remotes["github"].git)
	}
}

func TestListRemotes_notGitRepo(t *testing.T) {
	dir := t.TempDir()
	if _, err := listRemotes(dir); err == nil {
		t.Error("listRemotes on non-git dir: want error, got nil")
	}
}
