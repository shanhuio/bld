//go:build docker_e2e

// Run with: go test -tags=docker_e2e ./caco3/
//
// Requires a reachable Docker daemon at the default socket and network
// access to pull alpine:3.23 from Docker Hub.

package caco3

import (
	"os"
	"os/exec"
	"path/filepath"
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

func TestE2EBuildAndRun(t *testing.T) {
	requireDocker(t)

	// Pre-test cleanup: remove any leftover tags so AlwaysRebuild starts
	// from a clean slate. Re-run cleanup after the test as well.
	dockerRmiTags(e2eTags...)
	t.Cleanup(func() { dockerRmiTags(e2eTags...) })

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "WORKSPACE.caco3"), e2eWorkspace)
	writeFile(t,
		filepath.Join(root, "src/test.local/proj1/dockers/BUILD.caco3"),
		e2eProj1Build,
	)
	writeFile(t,
		filepath.Join(root, "src/test.local/proj2/dockers/BUILD.caco3"),
		e2eProj2Build,
	)
	writeFile(t,
		filepath.Join(root, "src/test.local/proj2/dockers/app/Dockerfile"),
		e2eAppDockerfile,
	)
	writeFile(t,
		filepath.Join(root, "src/test.local/proj2/dockers/payload.txt"),
		"hello from caco3\n",
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

	// Building verify pulls smoke in transitively (verify depends on
	// smoke's result.txt output), which in turn pulls in app and
	// alpine.
	targets := []string{"test.local/proj2/dockers/verify"}
	if errs := b.Build(targets); errs != nil {
		for _, e := range errs {
			t.Error(e)
		}
		t.FailNow()
	}

	outDir := filepath.Join(root, "out/test.local/proj2/dockers")

	// smoke wrote /result.txt out of the first container.
	bs, err := os.ReadFile(filepath.Join(outDir, "result.txt"))
	if err != nil {
		t.Fatalf("read smoke output: %v", err)
	}
	if got, want := string(bs), "hello from caco3\n"; got != want {
		t.Errorf("smoke output = %q, want %q", got, want)
	}

	// verify consumed smoke's result.txt as input and re-emitted it.
	bs, err = os.ReadFile(filepath.Join(outDir, "verified.txt"))
	if err != nil {
		t.Fatalf("read verify output: %v", err)
	}
	if got, want := string(bs), "hello from caco3\n"; got != want {
		t.Errorf("verify output = %q, want %q", got, want)
	}

	// The built image must exist locally.
	if err := exec.Command(
		"docker", "image", "inspect", "test.local/proj2/app:latest",
	).Run(); err != nil {
		t.Errorf("docker image inspect on built image: %v", err)
	}
}

var e2eWorkspace = multiLine(
	`repo_map {`,
	`    Src: {`,
	`        "test.local/proj1/dockers": "",`,
	`        "test.local/proj2/dockers": "",`,
	`    },`,
	`}`,
)

var e2eProj1Build = multiLine(
	`docker_pull {`,
	`    Name: "alpine",`,
	`    Pull: "alpine:3.23",`,
	`}`,
)

var e2eProj2Build = multiLine(
	`docker_build {`,
	`    Name: "app",`,
	`    From: ["/test.local/proj1/dockers/alpine"],`,
	`    Input: ["payload.txt"],`,
	`    PrefixDir: ".",`,
	`}`,
	``,
	`docker_run {`,
	`    Name: "smoke",`,
	`    Image: "app",`,
	`    Command: ["sh", "-c", "cat /payload.txt > /result.txt"],`,
	`    Output: {`,
	`        "result.txt": "/result.txt",`,
	`    },`,
	`}`,
	``,
	`docker_run {`,
	`    Name: "verify",`,
	`    Image: "app",`,
	`    Input: {`,
	`        "result.txt": "/in.txt",`,
	`    },`,
	`    Command: ["sh", "-c", "cat /in.txt > /verified.txt"],`,
	`    Output: {`,
	`        "verified.txt": "/verified.txt",`,
	`    },`,
	`}`,
)

var e2eAppDockerfile = multiLine(
	`FROM test.local/proj1/alpine:latest`,
	`COPY payload.txt /payload.txt`,
)
