//go:build docker_e2e

// Run with: go test -tags=docker_e2e ./lets/
//
// Requires a reachable Docker daemon at the default socket and network
// access to pull alpine:3.23 from Docker Hub.

package lets

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// e2eTags are the local image tags this test creates. Cleaned up both
// before and after the test runs.
var e2eTags = []string{
	"test.local/proj1/alpine:latest",
	"test.local/proj2/app:latest",
}

func requireDocker(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker CLI not found")
	}
	if err := exec.Command("docker", "version").Run(); err != nil {
		t.Skipf("docker daemon not reachable: %v", err)
	}
}

func dockerRmiTags(refs ...string) {
	for _, ref := range refs {
		_ = exec.Command("docker", "rmi", "-f", ref).Run()
	}
}

const e2eProj1Build = `image_pull {
    Name: "alpine",
    Pull: "alpine:3.23",
}
`

const e2eProj2Build = `image_build {
    Name: "app",
    From: ["/test.local/proj1/dockers/alpine"],
    Input: ["payload.txt"],
    PrefixDir: ".",
}

container_run {
    Name: "smoke",
    Image: "app",
    Command: ["sh", "-c", "cat /payload.txt > /result.txt"],
    Output: {
        "result.txt": "/result.txt",
    },
}

container_run {
    Name: "verify",
    Image: "app",
    Input: {
        "result.txt": "/in.txt",
    },
    Command: ["sh", "-c", "cat /in.txt > /verified.txt"],
    Output: {
        "verified.txt": "/verified.txt",
    },
}
`

const e2eAppDockerfile = `FROM test.local/proj1/alpine:latest
COPY payload.txt /payload.txt
`

// e2eSingleRepoTags are the local image tags TestE2E_singleRepoNoDeps creates.
var e2eSingleRepoTags = []string{
	"test.local/standalone/alpine:latest",
	"test.local/standalone/hello:latest",
}

// TestE2E_singleRepoNoDeps exercises Stage 2 of single-repo mode: the
// workspace declares a `repo` block, has no external dependencies, and
// builds its own rules entirely out of files at the repo root. All
// outputs land under _/out.
func TestE2E_singleRepoNoDeps(t *testing.T) {
	requireDocker(t)

	dockerRmiTags(e2eSingleRepoTags...)
	t.Cleanup(func() { dockerRmiTags(e2eSingleRepoTags...) })

	root := t.TempDir()
	writeFile(t,
		filepath.Join(root, "BUILD.lets"),
		singleRepoWorkspace+"\n"+singleRepoBuild,
	)
	writeFile(t, filepath.Join(root, "hello/Dockerfile"), singleRepoDockerfile)
	writeFile(t, filepath.Join(root, "payload.txt"), "single-repo says hi\n")

	b, err := NewBuilder(root, &Config{Root: root, AlwaysRebuild: true})
	if err != nil {
		t.Fatalf("NewBuilder: %v", err)
	}
	if _, errs := b.ReadWorkspace(); errs != nil {
		for _, e := range errs {
			t.Error(e)
		}
		t.FailNow()
	}

	// workSrcPath should resolve "smoke" to the full single-repo name.
	targets := []string{"smoke"}
	if errs := b.Build(targets); errs != nil {
		for _, e := range errs {
			t.Error(e)
		}
		t.FailNow()
	}

	outFile := filepath.Join(
		root, "_/out/test.local/standalone/dockers/result.txt",
	)
	bs, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read smoke output: %v", err)
	}
	if got, want := string(bs), "single-repo says hi\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}

	if err := exec.Command(
		"docker", "image", "inspect", "test.local/standalone/hello:latest",
	).Run(); err != nil {
		t.Errorf("docker image inspect on built image: %v", err)
	}
}

const singleRepoWorkspace = `repo {
    Name: "test.local/standalone/dockers",
}
`

const singleRepoBuild = `image_pull {
    Name: "alpine",
    Pull: "alpine:3.23",
}

image_build {
    Name: "hello",
    From: ["alpine"],
    Input: ["payload.txt"],
    PrefixDir: ".",
}

container_run {
    Name: "smoke",
    Image: "hello",
    Command: ["sh", "-c", "cat /payload.txt > /result.txt"],
    Output: {
        "result.txt": "/result.txt",
    },
}
`

const singleRepoDockerfile = `FROM test.local/standalone/alpine:latest
COPY payload.txt /payload.txt
`

// TestE2E_singleRepoWithDep exercises Stage 3 of single-repo mode: the
// workspace's self repo (test.local/proj2/dockers) declares one external
// dependency (test.local/proj1/dockers), pre-checked-out under _/src.
// The dep provides image_pull alpine; the self repo's image_build app
// references it via an absolute "From" path; container_run smoke/verify
// chain on top.
func TestE2E_singleRepoWithDep(t *testing.T) {
	requireDocker(t)

	dockerRmiTags(e2eTags...)
	t.Cleanup(func() { dockerRmiTags(e2eTags...) })

	root := t.TempDir()

	// Self-repo files at the repo root: the BUILD.lets begins with the repo
	// block (workspace declaration) followed by the build rules.
	writeFile(t,
		filepath.Join(root, "BUILD.lets"),
		singleRepoWithDepWorkspace+"\n"+e2eProj2Build,
	)
	writeFile(t, filepath.Join(root, "app/Dockerfile"), e2eAppDockerfile)
	writeFile(t, filepath.Join(root, "payload.txt"), "hello from lets\n")
	// Dependency pre-checked-out under _/src.
	writeFile(t,
		filepath.Join(root, "_/src/test.local/proj1/dockers/BUILD.lets"),
		e2eProj1Build,
	)

	b, err := NewBuilder(root, &Config{Root: root, AlwaysRebuild: true})
	if err != nil {
		t.Fatalf("NewBuilder: %v", err)
	}
	if _, errs := b.ReadWorkspace(); errs != nil {
		for _, e := range errs {
			t.Error(e)
		}
		t.FailNow()
	}

	// "verify" resolves to test.local/proj2/dockers/verify via workSrcPath.
	if errs := b.Build([]string{"verify"}); errs != nil {
		for _, e := range errs {
			t.Error(e)
		}
		t.FailNow()
	}

	outDir := filepath.Join(root, "_/out/test.local/proj2/dockers")

	bs, err := os.ReadFile(filepath.Join(outDir, "result.txt"))
	if err != nil {
		t.Fatalf("read smoke output: %v", err)
	}
	if got, want := string(bs), "hello from lets\n"; got != want {
		t.Errorf("smoke output = %q, want %q", got, want)
	}

	bs, err = os.ReadFile(filepath.Join(outDir, "verified.txt"))
	if err != nil {
		t.Fatalf("read verify output: %v", err)
	}
	if got, want := string(bs), "hello from lets\n"; got != want {
		t.Errorf("verify output = %q, want %q", got, want)
	}

	if err := exec.Command(
		"docker", "image", "inspect", "test.local/proj2/app:latest",
	).Run(); err != nil {
		t.Errorf("docker image inspect on built image: %v", err)
	}
}

const singleRepoWithDepWorkspace = `repo {
    Name: "test.local/proj2/dockers",
    Deps: {
        "test.local/proj1/dockers": "",
    },
}
`

func dockerRmVolume(name string) {
	_ = exec.Command("docker", "volume", "rm", "-f", name).Run()
}

// TestE2E_containerRunCacheVolume exercises CacheVolumes: a container_run
// mounts a host-global named volume, writes a marker into it, and the
// marker persists into a second run from a fresh workspace (proving the
// cache survives container teardown). A third run with AlwaysRebuild
// clears the volume, so the marker is gone again.
func TestE2E_containerRunCacheVolume(t *testing.T) {
	requireDocker(t)

	const (
		imgTag  = "test.local/cachetest/alpine:latest"
		volName = "lets-cache-e2e-cache-test"
	)
	dockerRmiTags(imgTag)
	dockerRmVolume(volName)
	t.Cleanup(func() {
		dockerRmiTags(imgTag)
		dockerRmVolume(volName)
	})

	// run builds the probe rule in a fresh workspace (so the rule's own
	// output cache never lets it skip the run) and returns the marker
	// state the container observed.
	run := func(alwaysRebuild bool) string {
		t.Helper()
		root := t.TempDir()
		writeFile(t,
			filepath.Join(root, "BUILD.lets"),
			cacheVolumeWorkspace+"\n"+cacheVolumeBuild,
		)

		b, err := NewBuilder(root, &Config{
			Root: root, AlwaysRebuild: alwaysRebuild,
		})
		if err != nil {
			t.Fatalf("NewBuilder: %v", err)
		}
		if _, errs := b.ReadWorkspace(); errs != nil {
			for _, e := range errs {
				t.Error(e)
			}
			t.FailNow()
		}
		if errs := b.Build([]string{"probe"}); errs != nil {
			for _, e := range errs {
				t.Error(e)
			}
			t.FailNow()
		}
		bs, err := os.ReadFile(filepath.Join(
			root, "_/out/test.local/cachetest/dockers/result.txt",
		))
		if err != nil {
			t.Fatalf("read probe output: %v", err)
		}
		return strings.TrimSpace(string(bs))
	}

	// First run sees a cold cache and writes the marker.
	if got := run(false); got != "cold" {
		t.Fatalf("run 1 = %q, want cold", got)
	}
	// Second run, fresh workspace: the global volume persists, so the
	// marker written by run 1 is still there.
	if got := run(false); got != "warm" {
		t.Fatalf("run 2 = %q, want warm (cache did not persist)", got)
	}
	// The volume must survive container teardown.
	if err := exec.Command(
		"docker", "volume", "inspect", volName,
	).Run(); err != nil {
		t.Errorf("cache volume missing after runs: %v", err)
	}
	// Third run with -rebuild clears the volume, so the marker is gone.
	if got := run(true); got != "cold" {
		t.Fatalf("run 3 (rebuild) = %q, want cold (cache not cleared)", got)
	}
}

const cacheVolumeWorkspace = `repo {
    Name: "test.local/cachetest/dockers",
}
`

const cacheVolumeBuild = `image_pull {
    Name: "alpine",
    Pull: "alpine:3.23",
}

container_run {
    Name: "probe",
    Image: "alpine",
    Command: [
        "sh", "-c",
        "if [ -f /cache/marker ]; then echo warm > /result.txt; else echo cold > /result.txt; fi; echo x > /cache/marker",
    ],
    CacheVolumes: {
        "/cache": "e2e-cache-test",
    },
    Output: {
        "result.txt": "/result.txt",
    },
}
`
