package caco3

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"shanhu.io/std/docker"
	"shanhu.io/std/lexing"
)

// Config provide the configuration to start a builder.
type Config struct {
	Root string // Root directory

	AlwaysRebuild bool // Always rebuild everything.

	UseDockerBuildCache bool // Use docker build cache.
}

// Builder builds stuff.
type Builder struct {
	env  *env
	opts *buildOpts
}

const workspaceFile = "WORKSPACE.lets"

func findRoot(cur string) (string, error) {
	for {
		f := filepath.Join(cur, workspaceFile)
		ok, err := isRegularFile(f)
		if err != nil {
			return "", fmt.Errorf("check %q: %w", f, err)
		}
		if ok {
			return cur, nil
		}
		if cur == "" || cur == "/" {
			break
		}
		cur = filepath.Dir(cur)
	}
	return "", errors.New("not in a workspace")
}

// NewBuilder creates a new builder that builds stuff.
func NewBuilder(workDir string, config *Config) (*Builder, error) {
	if !filepath.IsAbs(workDir) {
		return nil, fmt.Errorf("work dir %q is relative", workDir)
	}
	if filepath.Clean(workDir) != workDir {
		return nil, fmt.Errorf("work dir %q is dirty", workDir)
	}

	root := config.Root
	if root == "" {
		dir, err := findRoot(workDir)
		if err != nil {
			return nil, fmt.Errorf("find root: %w", err)
		}
		root = dir
	}

	srcDir := filepath.Join(root, "src")
	workSrcPath := ""
	if strings.HasPrefix(workDir, srcDir+"/") {
		workSrcPath = filepath.ToSlash(strings.TrimPrefix(workDir, srcDir+"/"))
	}

	env := &env{
		dock:        docker.NewUnixClient(""),
		rootDir:     root,
		workDir:     workDir,
		workSrcPath: workSrcPath,
		srcDir:      srcDir,
		outDir:      filepath.Join(root, "out"),
	}
	opts := &buildOpts{
		log:           os.Stderr,
		alwaysRebuild: config.AlwaysRebuild,
		docker: &dockerOpts{
			useBuildCache: config.UseDockerBuildCache,
		},
	}
	return &Builder{
		env:  env,
		opts: opts,
	}, nil
}

// ReadWorkspace reads and loads the WORKSPACE file into the build env.
// When the workspace declares a `repo` node, the env is switched into
// single-repo mode: srcDir / outDir move under _/src and _/out, and
// the env learns to redirect paths under the self-repo name back to
// the workspace root.
func (b *Builder) ReadWorkspace() (*Workspace, []*lexing.Error) {
	if b.env.workspace != nil {
		return b.env.workspace, nil
	}

	ws, errs := readWorkspace(b.env.root(workspaceFile))
	if errs != nil {
		return nil, errs
	}
	b.env.workspace = ws

	if ws.Repo != nil {
		if ws.Repo.Name == "" {
			return nil, lexing.SingleErr(
				errors.New("repo.Name must not be empty"),
			)
		}
		if err := b.env.setupSingleRepo(ws.Repo.Name); err != nil {
			return nil, lexing.SingleErr(err)
		}
	}
	return ws, nil
}

// Src returns the filesystem path to a source file.
func (b *Builder) Src(f string) string { return b.env.src(f) }

// Out returns the filesystme path to an output file.
func (b *Builder) Out(f string) string { return b.env.out(f) }

// SyncRepos synchronizes the repositories. When sums is nil, it pulls
// from the latest HEAD.
func (b *Builder) SyncRepos(sums *RepoSums, opts *SyncOptions) (
	*RepoSums, error,
) {
	return syncRepos(b.env, sums, opts)
}

// Build builds the given rules.
func (b *Builder) Build(rules []string) []*lexing.Error {
	if w := b.env.workSrcPath; w != "" {
		var absPaths []string
		for _, r := range rules {
			p := makePath(w, r)
			absPaths = append(absPaths, p)
		}
		rules = absPaths
	}

	nodes, nodeMap, errs := loadNodes(b.env, rules)
	if errs != nil {
		return errs
	}
	cacheFile, err := b.env.prepareOut("CACHE")
	if err != nil {
		return lexing.SingleErr(fmt.Errorf("prepare CACHE: %w", err))
	}
	cache, err := newBuildCache(cacheFile)
	if err != nil {
		err := fmt.Errorf("create build cache: %w", err)
		return lexing.SingleErr(err)
	}

	ctx := &buildContext{
		nodes: nodeMap,
		built: make(map[string]string),
		cache: cache,
	}
	return b.buildNodes(ctx, nodes)
}

func (b *Builder) buildNodes(
	ctx *buildContext, nodes []*buildNode,
) []*lexing.Error {
	b.env.nodeType = ctx.nodeType
	b.env.ruleType = ctx.ruleType

	for _, n := range nodes {
		if n.typ == nodeSrc {
			log.Printf("%s is a source file", n.name)
			continue
		}
		if _, err := b.buildNode(ctx, n); err != nil {
			return lexing.SingleErr(err)
		}
	}
	return nil
}

func (b *Builder) buildNode(ctx *buildContext, n *buildNode) (
	string, error,
) {
	if digest, ok := ctx.built[n.name]; ok {
		return digest, nil
	}
	digest := ""
	defer func() { ctx.built[n.name] = digest }()

	deps := make(map[string]string)
	for _, dep := range n.deps {
		depNode := ctx.nodes[dep]
		if depNode == nil {
			return "", fmt.Errorf("dep %q for %q not found", dep, n.name)
		}
		d, err := b.buildNode(ctx, depNode)
		if err != nil {
			return "", err
		}
		if d == "" {
			// If any dep is always rebuilding, then this node
			// is also always rebuilding.
			deps = nil
		} else if deps != nil {
			deps[dep] = d
		}
	}

	if deps != nil { // Not always rebuilding, so calculate the digest
		d, err := buildNodeDigest(b.env, n, deps)
		if err != nil {
			return "", fmt.Errorf("digest: %w", err)
		}
		digest = d
	}

	outputChanged := true
	if digest != "" {
		built, err := ctx.cache.get(digest)
		if err != nil {
			if !errors.Is(err, errCacheMiss) {
				return "", fmt.Errorf("check from build cache: %w", err)
			}
		} else {
			same, err := checkSameBuilt(b.env, built)
			if err != nil {
				return "", fmt.Errorf("check built: %w", err)
			}
			outputChanged = !same
		}
	}

	// Build.
	if !outputChanged && !b.opts.alwaysRebuild { // Cache hit.
		return digest, nil
	}
	if err := ctx.cache.remove(digest); err != nil {
		return "", fmt.Errorf("invalidate cache: %w", err)
	}

	if n.typ == nodeRule && n.rule != nil {
		log.Printf("BUILD %s", n.name)
		if err := n.rule.build(b.env, b.opts); err != nil {
			return "", fmt.Errorf("build %s: %w", n.name, err)
		}

		built, err := newBuilt(b.env, n.ruleMeta)
		if err != nil {
			return "", fmt.Errorf("make built: %w", err)
		}
		if err := ctx.cache.put(digest, built); err != nil {
			return "", fmt.Errorf("save in build cache: %w", err)
		}
	}

	return digest, nil
}
