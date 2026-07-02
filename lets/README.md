# lets

`lets` is a small, content-addressed build tool in the spirit of
[Bazel](https://bazel.build/), specialized for building and running
**OCI/Docker container images**. Build rules are declared in `BUILD.lets`
files, wired together into a dependency graph, and executed only when their
inputs change. Most build *actions* either produce a container image or run a
command inside one, so containers are both a first-class output and the
sandbox in which work happens.

The package is importable as a library (`shanhu.io/bld/lets`) and ships a
thin CLI under [`lets/`](./lets).

## Concepts

### Workspace

A workspace is a directory tree rooted at a repo. `lets` locates the root by
walking up from the current directory to the first directory that holds a
`.letsroot` stamp file or a `.git` directory, so commands can be run from any
subdirectory. (A bare `.git` repo therefore works as a root without an extra
stamp.)

The workspace *is* one repo, configured by the `repo` block at the head of the
root `BUILD.lets`. Its own files live at the workspace root, external
dependencies are checked out under `_/src`, and build outputs land under
`_/out`. The `_/` subtree is never scanned as source. The repo's `Deps` list
the dependency repos, which `lets sync` clones/updates.

The root `BUILD.lets` is a [jsonx](https://pkg.go.dev/shanhu.io/std/jsonx)
document (JSON with comments and a typed-entry syntax). The `repo` block must
be its first entry, ahead of any build rules:

```jsonx
// Name the self repo and list its dependency repos. lets builds the self
// repo's rules directly and checks the deps out under _/src.
repo {
    Name: "git.example.com/standalone/dockers",
    Deps: {
        "git.example.com/proj1/dockers": "",
        "git.example.com/proj2/dockers": "",
    },
}

// ... build rules follow ...
```

In `Deps`, an empty repo URL is derived automatically as
`git@<host>:<path>.git` (overridable per host via `GitHosting`).

### BUILD files

Each directory contributes rules through a `BUILD.lets` file — again a `jsonx`
document, parsed as a *series* of typed entries. Each entry's type selects a
rule kind and its body fills in the rule's fields:

```jsonx
image_pull {
    Name: "alpine",
    Pull: "alpine:3.23",
}

image_build {
    Name:      "app",
    From:      ["alpine"],
    Input:     ["payload.txt"],
    PrefixDir: ".",
}

container_run {
    Name:    "smoke",
    Image:   "app",
    Command: ["sh", "-c", "cat /payload.txt > /result.txt"],
    Output: {
        "result.txt": "/result.txt",
    },
}
```

### Names and paths

A rule's `Name` is a path relative to the BUILD file's directory; it cannot
escape that directory. Rules refer to each other (and to source files) by name:

- A **relative** reference (`"alpine"`, `"app/Dockerfile"`) resolves against
  the current BUILD file's directory.
- An **absolute** reference (`"/git.example.com/proj1/dockers/alpine"`) resolves
  from the workspace source root, letting one repo depend on a rule defined in
  another.

Any referenced path that is not itself a rule output is auto-registered as a
**source file** node.

### The build graph

Loading a set of targets produces a graph of nodes:

| node type | meaning                                            |
|-----------|----------------------------------------------------|
| `rule`    | a build rule with an action                        |
| `src`     | a source file on disk                              |
| `out`     | a file produced by some rule                       |
| `sub`     | a `sub_builds` pointer to more BUILD directories   |

The loader registers every rule and its declared outputs, recurses into
`sub_builds` directories, detects circular dependencies, and reports
redeclared or unresolved names with source positions.

### Caching

`lets` is content-addressed. For each rule it computes a **digest** over the
rule's own definition (action type, args, flags, the Dockerfile path, etc.)
combined with the digests of all of its dependencies. The digest keys a
persistent build cache (`CACHE` under the out dir) that records a `built`
manifest: the stat of each output file and, for image rules, the resulting
image repo/tag/ID.

Before running an action, `lets` looks up the digest and verifies the recorded
outputs still match what's on disk (and that image IDs still exist in the
daemon). If everything matches, the action is skipped. A rule whose digest is
empty (e.g. it depends on something that always rebuilds) is always re-executed.
The `-rebuild` flag forces everything to rebuild.

## Rules

| rule           | purpose                                                                 |
|----------------|-------------------------------------------------------------------------|
| `file_set`     | Select a set of source/output files (via explicit list, globs, `**` recursion, with `Ignore` patterns and `Include` of other file sets). Used as input to image builds. |
| `bundle`       | Group other rules under one name; no action of its own.                 |
| `image_pull`   | Pull an image from a registry and tag it locally. Optionally pins a `Digest` and verifies it; optionally saves a `.tar` output. |
| `image_build`  | Build an image from a `Dockerfile` plus a build context assembled from `Input` files, `ArchiveInput` zips, and `From` base-image rules. |
| `container_run`| Run a one-shot container against an image. Copies `Input`/`ArchiveInput` in, runs `Command`, copies `Output` files back out; supports `Env`, `WorkDir`, and a read-only `MountDir` (the rule's own build-file directory). |
| `download`     | Download a URL to an output file and verify its `sha256:` checksum.     |
| `sub_builds`   | List additional directories whose BUILD files should be loaded.         |

Image rules emit a small JSON "sum" (`<name>` image sum recording repo/tag/ID)
as their primary output, so downstream rules can depend on an upstream image by
referencing it.

## Command line

The CLI lives in [`lets/lets`](./lets) and exposes two subcommands:

```
lets build [flags] [targets...]    # build the named rules (relative to cwd)
lets sync  [flags] [targets...]    # clone/update source repos
```

Common `build` flags:

- `-root <dir>` — workspace root (default: discovered by walking up to a
  `.letsroot` stamp file or a `.git` directory).
- `-rebuild` — always rebuild, ignoring the cache.
- `-docker_build_cache` — use the Docker layer build cache (default `true`).

`sync` flags:

- `-pull` — sync to the latest remote `HEAD` instead of the pinned commits.
- `-save` — write the resolved commits to `sums.jsonx`.

`sync` pins each repo to a commit recorded in `sums.jsonx`; without `-pull` it
reproduces exactly those commits, fetching and fast-forwarding via an internal
`lets` stash branch and refusing to touch the self repo at the workspace root.

## Library usage

```go
b, err := lets.NewBuilder(workDir, &lets.Config{
    // Root: "",              // auto-discover if empty
    // AlwaysRebuild: false,
    // UseDockerBuildCache: true,
})
if err != nil { /* ... */ }

if _, errs := b.ReadWorkspace(); errs != nil {
    lexing.FprintErrs(os.Stderr, errs, workDir)
    return
}

if errs := b.Build([]string{"app", "smoke"}); errs != nil {
    lexing.FprintErrs(os.Stderr, errs, workDir)
}
```

The container backend is abstracted behind an internal `runtime` interface
(Docker today, via [`shanhu.io/std/docker`](https://pkg.go.dev/shanhu.io/std/docker)),
which keeps every pull/build/run/inspect/save operation in one place so other
runtimes can be added later.

## Testing

Unit tests run with the standard `go test ./lets/`. End-to-end tests that
exercise a real Docker daemon are guarded behind a build tag:

```
go test -tags=docker_e2e ./lets/
```

These require a reachable Docker daemon and network access to pull a base image
(`alpine:3.23`) from Docker Hub.
