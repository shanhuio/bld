package lets

import "io"

// runtime abstracts the container backend (Docker today; other
// runtimes in the future) used by build rules. The interface is broad on
// purpose for v0: the goal is to capture every operation lets needs from
// the runtime in one place, even if some methods could later be split
// into narrower capabilities.
type runtime interface {
	// Pull pulls one image from a registry and applies one or more
	// local tags to it. Implementations must be idempotent: pulling
	// an image that is already present is a no-op.
	//
	// If req.ExpectDigest is non-empty, the resulting image's repo
	// digests must include req.SrcRepo + "@" + req.ExpectDigest, or
	// the call fails.
	Pull(req *pullRequest) error

	// Inspect returns runtime-agnostic metadata for a local image
	// identified by repo:tag, repo@digest, or image ID.
	// Implementations must return an error that satisfies
	// errcode.IsNotFound when the image is not present locally.
	Inspect(ref string) (*imageInfo, error)

	// Save writes the image identified by ref as a gzipped tarball
	// at the given host path.
	Save(ref, file string) error

	// Build builds an image from a Dockerfile description and a
	// build context (regular files plus zip archives), then tags it
	// locally as ref.
	Build(ref string, req *buildRequest) error

	// Run executes a one-shot container against image ref. Inputs
	// are copied in before start; outputs are copied out after the
	// command exits (best effort: partial copy failures are reported
	// but do not mask a non-zero exit status). Returns the command's
	// exit status.
	Run(ref string, req *runRequest) (int, error)
}

// imageTag is a local (repo, tag) pair.
type imageTag struct {
	Repo string
	Tag  string
}

// pullRequest describes one PullAndTagImage operation.
type pullRequest struct {
	// SrcRepo is the registry-side repository to pull from
	// (e.g., "alpine" or "cr.shanhu.io/foo/bar").
	SrcRepo string
	// SrcTag is the tag (or digest, when ExpectDigest is set) to
	// pull. An empty value is treated as "latest".
	SrcTag string
	// ExpectDigest, if non-empty, is a sha256 digest the pulled
	// image must report under SrcRepo's RepoDigests.
	ExpectDigest string
	// Tags are the local (repo, tag) pairs to apply to the pulled
	// image after the pull succeeds. Must be non-empty.
	Tags []imageTag
}

// imageInfo holds the runtime-agnostic image metadata lets needs.
type imageInfo struct {
	// ID is the locally-assigned image identifier (typically a
	// sha256-prefixed hash).
	ID string
	// RepoDigests is the list of "repo@digest" strings reported by
	// the runtime. Used by callers for digest verification.
	RepoDigests []string
}

// buildRequest describes one BuildImage operation.
type buildRequest struct {
	// Dockerfile is the verbatim Dockerfile contents.
	Dockerfile string
	// Files are regular files included in the build context, in
	// the order they should appear.
	Files []buildFile
	// Archives are zip archives whose contents are extracted into
	// the build context.
	Archives []buildArchive
	// Args is the --build-arg key/value map.
	Args map[string]string
	// UseCache toggles whether the runtime's image-layer cache is
	// consulted during the build.
	UseCache bool
}

// buildFile is one regular file in a build context.
type buildFile struct {
	Source string // Host path on disk.
	Dest   string // Path inside the build-context tarball.
	Mode   int64  // Unix mode bits; if 0, taken from the source file.
}

// buildArchive is a zip archive expanded into the build context.
type buildArchive struct {
	Source string // Host path to the .zip file.
	Dest   string // Destination directory inside the build context.
}

// runRequest describes one RunContainer operation.
type runRequest struct {
	Cmd     []string          // Command to run; nil = image default.
	Env     map[string]string // Environment variables.
	WorkDir string            // Working directory inside the container.

	// Mounts are read-only host->container bind mounts attached
	// before start.
	Mounts []containerMount

	// Caches are named, persistent, read-write volumes mounted for
	// caching across runs. They are not hermetic: their contents never
	// affect a rule's outputs or digest. When ClearCaches is set, each
	// is emptied before the run.
	Caches      []containerCache
	ClearCaches bool

	// Inputs are regular files copied into the container before
	// start.
	Inputs []containerInput
	// Archives are zip files extracted into a directory inside the
	// container before start.
	Archives []containerArchive
	// Outputs are files (or directories) copied out of the
	// container after the command exits.
	Outputs []containerOutput

	// Log receives the merged stdout+stderr of the container while
	// it runs. Nil discards.
	Log io.Writer
}

// containerMount is a read-only host -> container bind mount.
type containerMount struct {
	Host string
	Cont string
}

// containerCache is a named, persistent read-write volume mounted for
// caching across runs.
type containerCache struct {
	Name string // backend volume name.
	Cont string // absolute mount path inside the container.
}

// containerInput is a regular file copied into the container.
type containerInput struct {
	Source string // Host path.
	Dest   string // Container path.
}

// containerArchive is a zip archive extracted into the container.
type containerArchive struct {
	Source string // Host path of the .zip file.
	Dest   string // Container directory to extract into.
}

// containerOutput is a file copied out of the container after the run.
type containerOutput struct {
	Cont string // Container path.
	Host string // Host path.
}
