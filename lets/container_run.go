package lets

import (
	"fmt"
	"log"
	"path"
	"regexp"
	"sort"
	"strings"

	"shanhu.io/std/docker"
	"shanhu.io/std/errcode"
	"shanhu.io/std/tarutil"
)

// cacheVolumePrefix namespaces the host-global volume names lets creates
// for cache mounts, so they are recognizable and never clash with
// user-managed volumes.
const cacheVolumePrefix = "lets-cache-"

// cacheVolumeLabel tags every cache volume lets creates, so they can be
// enumerated (e.g. for future pruning) without guessing by name.
const cacheVolumeLabel = "io.shanhu.lets.cache"

// cacheNameRe restricts logical cache names to characters that make a
// valid Docker volume name once prefixed.
var cacheNameRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)

// cacheVol is a resolved cache-volume mount for a containerRun.
type cacheVol struct {
	name string // backend volume name (cacheVolumePrefix + logical name).
	cont string // absolute mount path inside the container.
}

type containerRun struct {
	name    string
	path    string
	rule    *ContainerRun
	image   string
	ins     map[string]string
	archIns map[string]string
	deps    []string
	outs    []string
	outMap  map[string]string
	envs    map[string]string
	caches  []*cacheVol
}

func newContainerRun(_ *env, p string, r *ContainerRun) (*containerRun, error) {
	name := makeRelPath(p, r.Name)

	image := makePath(p, r.Image)
	var deps []string
	deps = append(deps, imageSumOut(image))

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

	caches, err := makeCacheVols(r.CacheVolumes)
	if err != nil {
		return nil, err
	}

	return &containerRun{
		name:    name,
		path:    p,
		rule:    r,
		image:   image,
		ins:     ins,
		archIns: archIns,
		deps:    deps,
		outs:    outs,
		outMap:  outMap,
		envs:    makeEnvVars(r.Envs, nil),
		caches:  caches,
	}, nil
}

// makeCacheVols resolves and validates the CacheVolumes map into a slice
// of cache mounts sorted by mount path. Each key is an absolute, clean
// (see path.Clean) mount path inside the container; each value is a
// logical cache name. The same volume may be mounted at multiple paths.
func makeCacheVols(cacheVolumes map[string]string) ([]*cacheVol, error) {
	if len(cacheVolumes) == 0 {
		return nil, nil
	}

	conts := make([]string, 0, len(cacheVolumes))
	for cont := range cacheVolumes {
		conts = append(conts, cont)
	}
	sort.Strings(conts)

	var vols []*cacheVol
	for _, cont := range conts {
		if !path.IsAbs(cont) {
			return nil, fmt.Errorf("cache volume path %q must be absolute", cont)
		}
		if cont != path.Clean(cont) {
			return nil, fmt.Errorf("cache volume path %q is not clean", cont)
		}
		name := cacheVolumes[cont]
		if !cacheNameRe.MatchString(name) {
			return nil, fmt.Errorf("invalid cache volume name %q", name)
		}
		vols = append(vols, &cacheVol{name: cacheVolumePrefix + name, cont: cont})
	}
	return vols, nil
}

func (r *containerRun) meta(env *env) (*buildRuleMeta, error) {
	dat := struct {
		Rule *ContainerRun
		Envs map[string]string `json:",omitempty"`
	}{
		Rule: r.rule,
		Envs: r.envs,
	}
	digest, err := makeDigest(ruleContainerRun, r.name, &dat)
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

func (r *containerRun) build(env *env, opts *buildOpts) error {
	contConfig := &docker.ContConfig{
		Cmd:     r.rule.Command,
		WorkDir: r.rule.WorkDir,
		Env:     r.envs,
	}

	if m := r.rule.MountDir; m != "" {
		contConfig.Mounts = append(contConfig.Mounts, &docker.ContMount{
			Host:     env.src(r.path),
			Cont:     m,
			ReadOnly: true,
		})
	}

	c := env.dock

	for _, vol := range r.caches {
		// -rebuild clears the cache so the run starts from a cold cache.
		// The volume is host-global, so this affects every rule sharing it.
		if opts.alwaysRebuild {
			if err := docker.RemoveVolume(c, vol.name); err != nil {
				if !errcode.IsNotFound(err) {
					return fmt.Errorf("clear cache volume %q: %w", vol.name, err)
				}
			}
		}
		if _, err := docker.CreateVolumeIfNotExist(c, vol.name, &docker.VolumeConfig{
			Labels: map[string]string{cacheVolumeLabel: "1"},
		}); err != nil {
			return fmt.Errorf("create cache volume %q: %w", vol.name, err)
		}
		contConfig.Mounts = append(contConfig.Mounts, &docker.ContMount{
			Host: vol.name,
			Cont: vol.cont,
			Type: docker.MountVolume,
		})
	}

	img, err := nameToRepoTag(r.image)
	if err != nil {
		return fmt.Errorf("map image name: %w", err)
	}

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
