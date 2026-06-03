package lets

import (
	"path/filepath"
	"testing"
)

// singleRepoEnvWithDep builds an env shaped like single-repo mode for the
// self repo "shanhu.io/bld" with one external dependency declared in the
// workspace repo_map. It mirrors what ReadWorkspace + setupSingleRepo
// produce, without touching the filesystem.
func singleRepoEnvWithDep(root, self, dep string) *env {
	return &env{
		rootDir:  root,
		srcDir:   filepath.Join(root, "_", "src"),
		outDir:   filepath.Join(root, "_", "out"),
		repoName: self,
		workspace: &Workspace{
			RepoMap: &RepoMap{
				Src: map[string]string{dep: ""},
			},
		},
	}
}

// TestDockerRunMountDir_dependency checks that when a docker_run rule lives
// in a repo that is a dependency of the root workspace, its MountDir target
// resolves to that dependency's own build-file directory (under _/src), not
// a path under the umbrella workspace root.
func TestDockerRunMountDir_dependency(t *testing.T) {
	const (
		root = "/ws"
		self = "shanhu.io/bld"
		dep  = "dep.example.com/lib/foo"
	)
	e := singleRepoEnvWithDep(root, self, dep)

	// A docker_run defined in the dependency's dockers/ subdirectory.
	dir := dep + "/dockers"
	r := newDockerRun(e, dir, &DockerRun{
		Name:     "smoke",
		Image:    "app",
		MountDir: "/dir",
		Command:  []string{"true"},
	})

	if r.path != dir {
		t.Fatalf("dockerRun.path = %q, want %q", r.path, dir)
	}

	got := e.src(r.path)
	want := filepath.Join(root, "_", "src", dep, "dockers")
	if got != want {
		t.Errorf("mount dir = %q, want dependency build dir %q", got, want)
	}
	// It must land under the dependency checkout, not the workspace root.
	if got == filepath.Join(root, "dockers") {
		t.Errorf("mounted workspace-relative dir %q for a dependency rule", got)
	}
}

// TestDockerRunMountDir_selfRepo checks the self-repo case: a rule in the
// root repo mounts a directory directly under the workspace root.
func TestDockerRunMountDir_selfRepo(t *testing.T) {
	const (
		root = "/ws"
		self = "shanhu.io/bld"
		dep  = "dep.example.com/lib/foo"
	)
	e := singleRepoEnvWithDep(root, self, dep)

	dir := self + "/dockers"
	r := newDockerRun(e, dir, &DockerRun{
		Name:     "smoke",
		Image:    "app",
		MountDir: "/dir",
	})

	got := e.src(r.path)
	want := filepath.Join(root, "dockers")
	if got != want {
		t.Errorf("mount dir = %q, want self-repo build dir %q", got, want)
	}
}
