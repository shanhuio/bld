package lets

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"shanhu.io/std/docker"
)

type env struct {
	dock *docker.Client

	rootDir     string
	workDir     string
	workSrcPath string

	srcDir string
	outDir string

	// repoName is the canonical name of the self repo, taken from the
	// workspace's repo node. src() redirects paths under repoName/ to
	// rootDir, while dependency paths resolve under srcDir
	// (<rootDir>/_/src).
	repoName string

	workspace *Workspace // Lazily loaded.

	nodeType func(name string) string
	ruleType func(name string) string
}

func (e *env) prepareOut(ps ...string) (string, error) {
	p := e.out(ps...)
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return p, nil
}

func dirFilePath(dir string, ps ...string) string {
	if len(ps) == 0 {
		return dir
	}
	p := path.Join(ps...)
	return filepath.Join(dir, filepath.FromSlash(p))
}

func (e *env) root(ps ...string) string {
	return dirFilePath(e.rootDir, ps...)
}

func (e *env) out(ps ...string) string {
	return dirFilePath(e.outDir, ps...)
}

func (e *env) src(ps ...string) string {
	if len(ps) > 0 {
		full := path.Join(ps...)
		if full == e.repoName || strings.HasPrefix(full, e.repoName+"/") {
			rest := strings.TrimPrefix(full, e.repoName)
			rest = strings.TrimPrefix(rest, "/")
			return dirFilePath(e.rootDir, rest)
		}
	}
	return dirFilePath(e.srcDir, ps...)
}

// srcName converts an absolute filesystem path back to its logical
// source name (the inverse of env.src). It checks srcDir first so that
// dependency files under _/src in single-repo mode are not mistaken for
// self-repo files (srcDir is itself nested inside rootDir).
func (e *env) srcName(physical string) (string, error) {
	if rel, ok := pathUnder(e.srcDir, physical); ok {
		return filepath.ToSlash(rel), nil
	}
	if rel, ok := pathUnder(e.rootDir, physical); ok {
		if rel == "" {
			return e.repoName, nil
		}
		return path.Join(e.repoName, filepath.ToSlash(rel)), nil
	}
	return "", fmt.Errorf("path %q is outside any source root", physical)
}

// setupSingleRepo wires the env to the self repo named in the workspace,
// rooting srcDir and outDir under _/src and _/out and computing
// workSrcPath relative to the self repo. Called from ReadWorkspace.
func (e *env) setupSingleRepo(name string) error {
	e.repoName = name
	e.srcDir = filepath.Join(e.rootDir, "_", "src")
	e.outDir = filepath.Join(e.rootDir, "_", "out")

	// If we're inside _/src, we're under a dependency checkout, so report
	// the path relative to it. Otherwise we're in the self repo and the
	// logical workSrcPath is name + (rel from rootDir).
	if rel, ok := strings.CutPrefix(
		e.workDir, e.srcDir+string(filepath.Separator),
	); ok {
		e.workSrcPath = filepath.ToSlash(rel)
		return nil
	}
	rel, err := filepath.Rel(e.rootDir, e.workDir)
	if err != nil {
		return fmt.Errorf("rel work dir: %w", err)
	}
	switch {
	case rel == ".":
		e.workSrcPath = name
	case strings.HasPrefix(rel, ".."):
		e.workSrcPath = ""
	default:
		e.workSrcPath = path.Join(name, filepath.ToSlash(rel))
	}
	return nil
}
