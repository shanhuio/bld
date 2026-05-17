package caco3

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func currentGitCommit(dir string) (string, error) {
	branches, err := runCmdOutput(dir, "git", "branch")
	if err != nil {
		return "", fmt.Errorf("list branches: %w", err)
	}
	if len(bytes.TrimSpace(branches)) == 0 {
		return "", nil
	}

	ret, err := runCmdOutput(
		dir, "git", "show", "HEAD", "-s", "--format=%H",
	)
	if err != nil {
		return "", fmt.Errorf("get HEAD commit: %w", err)
	}
	return strings.TrimSpace(string(ret)), nil
}

type syncResult struct {
	commit  string
	updated bool
}

func gitSync(name, dir, remote, commit string) (*syncResult, error) {
	if commit == "" {
		latest, err := runCmdOutput(dir, "git", "ls-remote", remote, "HEAD")
		if err != nil {
			return nil, fmt.Errorf("git ls-remote: %w", err)
		}
		line := strings.TrimSpace(string(latest))
		fields := strings.Fields(line)
		if len(fields) == 0 {
			return nil, fmt.Errorf("bad remote commit: %q", line)
		}
		commit = fields[0]
	}

	gitDir := filepath.Join(dir, ".git")
	exist, err := isDir(gitDir)
	if err != nil {
		return nil, fmt.Errorf("check git dir: %w", err)
	}

	const stashBranch = "caco3"

	if !exist {
		if err := runCmd(dir, "git", "init", "-q"); err != nil {
			return nil, fmt.Errorf("git init: %w", err)
		}
		if err := runCmd(
			dir, "git", "remote", "add", "origin", remote,
		); err != nil {
			return nil, fmt.Errorf("git add remote: %w", err)
		}

		log.Printf(
			"[new %s] %s\n", shortDigest(commit), name,
		)
	} else {
		cur, err := currentGitCommit(dir)
		if err != nil {
			return nil, fmt.Errorf("get current comment: %w", err)
		}
		if cur == commit {
			return &syncResult{commit: cur}, nil
		}

		if cur != "" {
			hasCommit, err := callCmd(
				dir, "git", "cat-file", "-e", commit,
			)
			if err != nil {
				return nil, fmt.Errorf("git check commit: %w", err)
			}
			if hasCommit {
				isAncestor, err := callCmd(
					dir, "git", "merge-base", "--is-ancestor", commit, cur,
				)
				if err != nil {
					return nil, fmt.Errorf("git merge check: %w", err)
				}
				if isAncestor {
					// merge will be a noop, just update stash branch.
					if err := runCmd(
						dir, "git", "branch", "-q", "-f", stashBranch, commit,
					); err != nil {
						return nil, fmt.Errorf("git branch: %w", err)
					}
					return &syncResult{commit: cur}, nil
				}
			}

			log.Printf(
				"[%s..%s] %s\n",
				shortDigest(cur), shortDigest(commit), name,
			)
		} else {
			log.Printf(
				"[new %s] %s\n", shortDigest(commit), name,
			)
		}
	}

	// fetch to the stash branch and then merge.
	if err := runCmd(
		dir, "git", "fetch", "-q", remote, "HEAD",
	); err != nil {
		return nil, fmt.Errorf("git fetch: %w", err)
	}
	if err := runCmd(
		dir, "git", "branch", "-q", "-f", stashBranch, commit,
	); err != nil {
		return nil, fmt.Errorf("git branch stash: %w", err)
	}
	if err := runCmd(
		dir, "git", "merge", "-q", stashBranch,
	); err != nil {
		return nil, fmt.Errorf("git merge stash: %w", err)
	}

	return &syncResult{
		commit:  commit,
		updated: true,
	}, nil
}

// SyncOptions contains options for syncing remote repositories.
type SyncOptions struct {
	// Set remotes for existing repositories.
	SetRemotes bool
}

func syncRepos(env *env, sums *RepoSums, opts *SyncOptions) (
	*RepoSums, error,
) {
	ws := env.workspace

	var dirs []string
	for dir := range ws.RepoMap.Src {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	gitHosting := ws.RepoMap.GitHosting
	if gitHosting == nil {
		gitHosting = make(map[string]string)
	}
	repos := make(map[string]string)
	for _, dir := range dirs {
		repo := ws.RepoMap.Src[dir]
		if repo == "" {
			domain, p, ok := strings.Cut(dir, "/")
			if !ok {
				domain = dir
				p = ""
			}
			gitHost := domain
			if alt, found := gitHosting[domain]; found {
				gitHost = alt
			}
			repo = fmt.Sprintf("git@%s:%s.git", gitHost, p)
		}
		repos[dir] = repo
	}

	curSums := &RepoSums{
		RepoCommits: make(map[string]string),
	}

	for _, dir := range dirs {
		git := repos[dir]
		srcDir := env.src(dir)
		if err := os.MkdirAll(srcDir, 0755); err != nil {
			return nil, fmt.Errorf("make dir for %q: %w", dir, err)
		}

		commit := ""
		if sums != nil {
			c, ok := sums.RepoCommits[dir]
			if !ok {
				return nil, fmt.Errorf("commit missing for %q", dir)
			}
			commit = c
		}
		result, err := gitSync(dir, srcDir, git, commit)
		if err != nil {
			return nil, fmt.Errorf("git sync %q: %w", dir, err)
		}
		curSums.RepoCommits[dir] = result.commit
	}

	if opts.SetRemotes {
		for _, dir := range dirs {
			srcDir := env.src(dir)
			remotes, err := listRemotes(srcDir)
			if err != nil {
				return nil, fmt.Errorf("list remotes: %w", err)
			}

			var wants []*gitRemote
			wants = append(wants, &gitRemote{
				name: "origin",
				git:  repos[dir],
			})
			for _, extra := range ws.RepoMap.ExtraRemotes {
				if url, ok := extra.URL[dir]; ok {
					wants = append(wants, &gitRemote{
						name: extra.Name,
						git:  url,
					})
				}
			}

			for _, r := range wants {
				if cur, ok := remotes[r.name]; !ok {
					log.Printf("%q: add remote %q", dir, r.name)
					if err := runCmd(
						srcDir, "git", "remote", "add", r.name, r.git,
					); err != nil {
						return nil, fmt.Errorf(
							"add git remote %q for %q: %w", r.name, dir, err,
						)
					}
				} else if cur.git != r.git {
					log.Printf("%q: set remote %q", dir, r.name)
					if err := runCmd(
						srcDir, "git", "remote", "set-url", r.name, r.git,
					); err != nil {
						return nil, fmt.Errorf(
							"set git remote %q for %q: %w", r.name, dir, err,
						)
					}
				}
			}
		}
	}

	return curSums, nil
}
