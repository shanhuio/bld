package caco3

import (
	"fmt"
	"log"
	"path"
	"sort"
	"strings"

	"shanhu.io/std/docker"
	"shanhu.io/std/tarutil"
)

type dockerRun struct {
	name    string
	rule    *DockerRun
	image   string
	ins     map[string]string
	archIns map[string]string
	deps    []string
	outs    []string
	outMap  map[string]string
	envs    map[string]string
}

func newDockerRun(_ *env, p string, r *DockerRun) *dockerRun {
	name := makeRelPath(p, r.Name)

	image := makePath(p, r.Image)
	var deps []string
	deps = append(deps, dockerSumOut(image))

	depsMap := make(map[string]bool)
	for _, d := range r.Deps {
		depsMap[makePath(p, d)] = true
	}

	ins := make(map[string]string)
	for f, v := range r.Input {
		inPath := makePath(p, f)
		ins[inPath] = v
		depsMap[inPath] = true
	}
	archIns := make(map[string]string)
	for f, v := range r.ArchiveInput {
		inPath := makePath(p, f)
		archIns[inPath] = v
		depsMap[inPath] = true
	}
	deps = append(deps, sortedStrList(depsMap)...)

	var outs []string
	outMap := make(map[string]string)
	for f, v := range r.Output {
		outPath := makeRelPath(p, f)
		outs = append(outs, outPath)
		outMap[outPath] = v
	}
	outs = sortedStrList(makeStrSet(outs))

	return &dockerRun{
		name:    name,
		rule:    r,
		image:   image,
		ins:     ins,
		archIns: archIns,
		deps:    deps,
		outs:    outs,
		outMap:  outMap,
		envs:    makeDockerVars(r.Envs, nil),
	}
}

func (r *dockerRun) meta(env *env) (*buildRuleMeta, error) {
	dat := struct {
		Rule *DockerRun
		Envs map[string]string `json:",omitempty"`
	}{
		Rule: r.rule,
		Envs: r.envs,
	}
	digest, err := makeDigest(ruleDockerRun, r.name, &dat)
	if err != nil {
		return nil, fmt.Errorf("digest: %w", err)
	}

	return &buildRuleMeta{
		name:   r.name,
		outs:   r.outs,
		deps:   r.deps,
		digest: digest,
	}, nil
}

func (r *dockerRun) build(env *env, opts *buildOpts) error {
	contConfig := &docker.ContConfig{
		Cmd:     r.rule.Command,
		WorkDir: r.rule.WorkDir,
		Env:     r.envs,
	}

	if m := r.rule.MountWorkspace; m != "" {
		contConfig.Mounts = append(contConfig.Mounts, &docker.ContMount{
			Host:     env.rootDir,
			Cont:     m,
			ReadOnly: true,
		})
	}

	img, err := env.nameToRepoTag(r.image)
	if err != nil {
		return fmt.Errorf("map image name: %w", err)
	}

	c := env.dock

	cont, err := docker.CreateCont(c, img, contConfig)
	if err != nil {
		return fmt.Errorf("create container: %w", err)
	}
	defer cont.Drop()

	if len(r.ins)+len(r.archIns) > 0 {
		ts := tarutil.NewStream()

		var ins []string
		for in := range r.ins {
			ins = append(ins, in)
		}
		sort.Strings(ins)

		for _, in := range ins {
			var f string
			switch typ := env.nodeType(in); typ {
			case "":
				return fmt.Errorf("input %q not found", in)
			case nodeSrc:
				f = env.src(in)
			case nodeOut:
				f = env.out(in)
			default:
				return fmt.Errorf("unknown type %q", typ)
			}

			dest := r.ins[in]
			ts.AddFile(dest, new(tarutil.Meta), f)
		}

		var archIns []string
		for in := range r.archIns {
			archIns = append(archIns, in)
		}
		sort.Strings(archIns)

		for _, in := range archIns {
			var f string
			switch typ := env.nodeType(in); typ {
			case "":
				return fmt.Errorf("archive input %q not found", in)
			case nodeSrc:
				f = env.src(in)
			case nodeOut:
				f = env.out(in)
			default:
				return fmt.Errorf("unknown type %q", typ)
			}
			dest := r.archIns[in]
			base := path.Base(in)
			if strings.HasSuffix(base, ".zip") {
				ts.AddZipFile(dest, f)
			} else {
				return fmt.Errorf("unknown archive type %q", base)
			}
		}

		if err := docker.CopyInTarStream(cont, ts, "/"); err != nil {
			return fmt.Errorf("copy input: %w", err)
		}
	}

	if err := cont.Start(); err != nil {
		return fmt.Errorf("start container: %w", err)
	}
	if err := cont.FollowLogs(opts.log); err != nil {
		return fmt.Errorf("stream logs: %w", err)
	}

	status, err := cont.Wait(docker.NotRunning)
	if err != nil {
		return fmt.Errorf("wait container: %w", err)
	}
	for _, out := range r.outs {
		from := r.outMap[out]
		to := out

		f, err := env.prepareOut(to)
		if err != nil {
			return fmt.Errorf("prepare output: %s: %w", to, err)
		}

		if err := cont.CopyOutFile(from, f); err != nil {
			if status == 0 {
				return fmt.Errorf("copy %s: %w", to, err)
			}
			log.Printf("copy %s: %s", to, err)
		}
	}

	if status != 0 {
		return fmt.Errorf("exit with %d", status)
	}

	return nil
}
