package caco3

import (
	"fmt"
	"log"
	"strings"

	"shanhu.io/std/docker"
)

type dockerPull struct {
	name    string
	rule    *DockerPull
	repoTag string
	out     string
	tarOut  string
}

func newDockerPull(env *env, p string, r *DockerPull) (*dockerPull, error) {
	name := makeRelPath(p, r.Name)
	repoTag, err := env.nameToRepoTag(name)
	if err != nil {
		return nil, fmt.Errorf("invalid docker pull name: %w", err)
	}
	pull := &dockerPull{
		name:    name,
		rule:    r,
		repoTag: repoTag,
		out:     dockerSumOut(name),
	}
	if r.OutputTar {
		pull.tarOut = dockerTarOut(name)
	}
	return pull, nil
}

func (p *dockerPull) pull(env *env) (*dockerSum, error) {
	r := p.rule

	repo, tag := parseRepoTag(p.repoTag)
	srcRepo, srcTag := repo, tag

	if r.Pull != "" {
		srcRepo, srcTag = parseRepoTag(r.Pull)
	}

	digest := r.Digest

	from := repoTag(srcRepo, srcTag)
	pullTag := srcTag

	if digest != "" {
		from = fmt.Sprintf("%s@%s", srcRepo, digest)
		pullTag = digest
	}

	if err := docker.PullImage(env.dock, srcRepo, pullTag); err != nil {
		return nil, fmt.Errorf("pull image: %w", err)
	}
	if err := docker.TagImage(env.dock, from, srcRepo, srcTag); err != nil {
		return nil, fmt.Errorf("tag image as source: %w", err)
	}
	if !(repo == srcRepo && tag == srcTag) {
		if err := docker.TagImage(env.dock, from, repo, tag); err != nil {
			return nil, fmt.Errorf("re-tag output image: %w", err)
		}
	}
	out := repoTag(repo, tag)
	info, err := docker.InspectImage(env.dock, out)
	if err != nil {
		return nil, fmt.Errorf("inspect image: %w", err)
	}

	var repoDigests []string
	digestPrefix := srcRepo + "@"
	for _, digest := range info.RepoDigests {
		if strings.HasPrefix(digest, digestPrefix) {
			repoDigests = append(repoDigests, digest)
		}
	}
	if digest != "" {
		digestWant := digestPrefix + digest
		found := false
		for _, digest := range repoDigests {
			if digest == digestWant {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf(
				"digest mismatch, got %q, want %q",
				info.RepoDigests, digestWant,
			)
		}
	}

	sum := newDockerSum(repo, tag, info.ID)
	sum.Origin = from
	return sum, nil
}

func (p *dockerPull) build(env *env, opts *buildOpts) error {
	sum, err := p.pull(env)
	if err != nil {
		return err
	}
	out, err := env.prepareOut(p.out)
	if err != nil {
		return fmt.Errorf("prepare sum output: %w", err)
	}
	if err := writeJSONFile(out, sum); err != nil {
		return fmt.Errorf("write image sum: %w", err)
	}

	if p.tarOut != "" {
		log.Printf("Saving %s", p.tarOut)
		out, err := env.prepareOut(p.tarOut)
		if err != nil {
			return fmt.Errorf("prepare tar output: %w", err)
		}
		if err := docker.SaveImageGz(env.dock, sum.ID, out); err != nil {
			return fmt.Errorf("save image as tar: %w", err)
		}
	}
	return nil
}

func (p *dockerPull) meta(env *env) (*buildRuleMeta, error) {
	digest, err := makeDigest(ruleDockerPull, p.name, p.rule)
	if err != nil {
		return nil, fmt.Errorf("digest: %w", err)
	}

	outs := []string{p.out}
	if p.tarOut != "" {
		outs = append(outs, p.tarOut)
	}

	return &buildRuleMeta{
		name:      p.name,
		outs:      outs,
		dockerOut: true,
		digest:    digest,
	}, nil
}
