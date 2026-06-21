package lets

import (
	"errors"
	"os"

	"shanhu.io/std/jsonx"
	"shanhu.io/std/lexing"
)

// repoEntry is the entry type that, at the head of the root BUILD.lets,
// declares the workspace.
const repoEntry = "repo"

// letsRootFile is a stamp file that explicitly marks a workspace root.
const letsRootFile = ".letsroot"

// Workspace specifies how to build a project. It is declared by the leading
// repo block of the workspace root's BUILD.lets file.
type Workspace struct {
	Repo *Repo
}

// Repo names the self repo of the workspace and lists its dependency
// repos. It is required: lets builds this repo's own rules directly from
// the workspace root and resolves cross-repo dependencies under _/src.
type Repo struct {
	Name string

	// Deps lists the dependency repos to check out under _/src, mapping
	// each repo's import path to its git remote URL. An empty URL is
	// derived as git@<host>:<path>.git (see GitHosting).
	Deps map[string]string `json:",omitempty"`

	// GitHosting overrides the git host for a domain when deriving an
	// empty Deps URL.
	GitHosting map[string]string `json:",omitempty"`
}

// readWorkspace reads the workspace declaration from the leading repo block
// of the root BUILD.lets file f. Build rules are validated when the build
// file is loaded, so here non-repo entries are parsed leniently and ignored.
func readWorkspace(f string) (*Workspace, []*lexing.Error) {
	tm := func(t string) any {
		if t == repoEntry {
			return new(Repo)
		}
		m := map[string]any{}
		return &m
	}
	entries, errs := jsonx.ReadSeriesFile(f, tm)
	if errs != nil {
		return nil, errs
	}

	ws := new(Workspace)
	for i, entry := range entries {
		repo, ok := entry.V.(*Repo)
		if !ok {
			continue
		}
		if i != 0 {
			return nil, []*lexing.Error{{
				Pos: entry.Pos,
				Err: errors.New("repo must be the first entry in BUILD.lets"),
			}}
		}
		ws.Repo = repo
	}
	return ws, nil
}

// RepoSums records the checkums and git commits of a build.
type RepoSums struct {
	RepoCommits map[string]string
}

// ReadRepoSums reads in the workspaces's repo checksum file.
func ReadRepoSums(f string) (*RepoSums, error) {
	b := new(RepoSums)
	if err := jsonx.ReadFile(f, b); err != nil {
		return nil, err
	}
	return b, nil
}

// SaveRepoSums saves sums to f. When there are no recorded commits, it
// writes nothing and removes any existing sums file, so a workspace without
// dependencies does not carry an empty sums file.
func SaveRepoSums(f string, sums *RepoSums) error {
	if sums == nil || len(sums.RepoCommits) == 0 {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	return jsonx.WriteFile(f, sums)
}
