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

const e2eProj1Build = `docker_pull {
    Name: "alpine",
    Pull: "alpine:3.23",
}
`

const e2eProj2Build = `docker_build {
    Name: "app",
    From: ["/test.local/proj1/dockers/alpine"],
    Input: ["payload.txt"],
    PrefixDir: ".",
}

docker_run {
    Name: "smoke",
    Image: "app",
    Command: ["sh", "-c", "cat /payload.txt > /result.txt"],
    Output: {
        "result.txt": "/result.txt",
    },
}

docker_run {
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

const singleRepoBuild = `docker_pull {
    Name: "alpine",
    Pull: "alpine:3.23",
}

docker_build {
    Name: "hello",
    From: ["alpine"],
    Input: ["payload.txt"],
    PrefixDir: ".",
}

docker_run {
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
// The dep provides docker_pull alpine; the self repo's docker_build app
// references it via an absolute "From" path; docker_run smoke/verify
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
