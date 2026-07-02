package lets

import (
	"path/filepath"
	"testing"
)

// singleRepoEnvWithDep builds an env for the self repo "shanhu.io/bld"
// with one external dependency declared in the workspace repo's Deps. It
// mirrors what ReadWorkspace + setupSingleRepo produce, without touching
// the filesystem.
func singleRepoEnvWithDep(root, self, dep string) *env {
	return &env{
		rootDir:  root,
		srcDir:   filepath.Join(root, "_", "src"),
		outDir:   filepath.Join(root, "_", "out"),
		repoName: self,
		workspace: &Workspace{
			Repo: &Repo{
				Name: self,
				Deps: map[string]string{dep: ""},
			},
		},
	}
}

// TestContainerRunMountDir_dependency checks that when a container_run rule lives
// in a repo that is a dependency of the root workspace, its MountDir target
// resolves to that dependency's own build-file directory (under _/src), not
// a path under the umbrella workspace root.
func TestContainerRunMountDir_dependency(t *testing.T) {
	const (
		root = "/ws"
		self = "shanhu.io/bld"
		dep  = "dep.example.com/lib/foo"
	)
	e := singleRepoEnvWithDep(root, self, dep)

	// A container_run defined in the dependency's dockers/ subdirectory.
	dir := dep + "/dockers"
	r, err := newContainerRun(e, dir, &ContainerRun{
		Name:     "smoke",
		Image:    "app",
		MountDir: "/dir",
		Command:  []string{"true"},
	})
	if err != nil {
		t.Fatalf("newContainerRun: %v", err)
	}

	if r.path != dir {
		t.Fatalf("containerRun.path = %q, want %q", r.path, dir)
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

// TestContainerRunMountDir_selfRepo checks the self-repo case: a rule in the
// root repo mounts a directory directly under the workspace root.
func TestContainerRunMountDir_selfRepo(t *testing.T) {
	const (
		root = "/ws"
		self = "shanhu.io/bld"
		dep  = "dep.example.com/lib/foo"
	)
	e := singleRepoEnvWithDep(root, self, dep)

	dir := self + "/dockers"
	r, err := newContainerRun(e, dir, &ContainerRun{
		Name:     "smoke",
		Image:    "app",
		MountDir: "/dir",
	})
	if err != nil {
		t.Fatalf("newContainerRun: %v", err)
	}

	got := e.src(r.path)
	want := filepath.Join(root, "dockers")
	if got != want {
		t.Errorf("mount dir = %q, want self-repo build dir %q", got, want)
	}
}

func TestMakeCacheVols_sortedAndPrefixed(t *testing.T) {
	vols, err := makeCacheVols(map[string]string{
		"/root/go/pkg":   "go-pkg",
		"/var/cache/apk": "apk-cache",
	})
	if err != nil {
		t.Fatalf("makeCacheVols: %v", err)
	}
	want := []*cacheVol{
		{name: "lets-cache-go-pkg", cont: "/root/go/pkg"},
		{name: "lets-cache-apk-cache", cont: "/var/cache/apk"},
	}
	if len(vols) != len(want) {
		t.Fatalf("got %d volumes, want %d: %+v", len(vols), len(want), vols)
	}
	for i, w := range want {
		if *vols[i] != *w {
			t.Errorf("vols[%d] = %+v, want %+v", i, vols[i], w)
		}
	}
}

// TestMakeCacheVols_sameVolumeMultiplePaths checks that one logical cache
// name may back mounts at more than one container path.
func TestMakeCacheVols_sameVolumeMultiplePaths(t *testing.T) {
	vols, err := makeCacheVols(map[string]string{
		"/a": "shared",
		"/b": "shared",
	})
	if err != nil {
		t.Fatalf("makeCacheVols: %v", err)
	}
	want := []*cacheVol{
		{name: "lets-cache-shared", cont: "/a"},
		{name: "lets-cache-shared", cont: "/b"},
	}
	if len(vols) != len(want) {
		t.Fatalf("got %d volumes, want %d: %+v", len(vols), len(want), vols)
	}
	for i, w := range want {
		if *vols[i] != *w {
			t.Errorf("vols[%d] = %+v, want %+v", i, vols[i], w)
		}
	}
}

func TestMakeCacheVols_empty(t *testing.T) {
	vols, err := makeCacheVols(nil)
	if err != nil {
		t.Fatalf("makeCacheVols: %v", err)
	}
	if vols != nil {
		t.Errorf("vols = %+v, want nil", vols)
	}
}

func TestMakeCacheVols_errors(t *testing.T) {
	for _, c := range []struct {
		name string
		vols map[string]string
	}{
		{
			name: "relative path",
			vols: map[string]string{"relative/path": "c"},
		},
		{
			name: "unclean path",
			vols: map[string]string{"/root/go/../go/pkg/": "c"},
		},
		{
			name: "invalid volume name",
			vols: map[string]string{"/cache": "bad/name"},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, err := makeCacheVols(c.vols); err == nil {
				t.Fatalf("makeCacheVols: want error, got nil")
			}
		})
	}
}
