package lets

import (
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strings"

	"shanhu.io/std/docker"
	"shanhu.io/std/tarutil"
)

type imageBuild struct {
	name           string
	rule           *ImageBuild
	fromRuleSums   []string
	dockerfilePath string
	inputs         []string
	archInputs     []string
	prefixDir      string
	repoTag        string
	args           map[string]string
	out            string
	tarOut         string
}

func newImageBuild(env *env, p string, r *ImageBuild) (
	*imageBuild, error,
) {
	name := makeRelPath(p, r.Name)

	var f string
	if r.Dockerfile == "" {
		f = path.Join(name, "Dockerfile")
	} else {
		f = makePath(p, r.Dockerfile)
	}

	var fromRuleSums []string
	if len(r.From) > 0 {
		for _, from := range r.From {
			rp := makePath(p, from)
			fromRuleSums = append(fromRuleSums, imageSumOut(rp))
		}
	}

	repoTag, err := nameToRepoTag(name)
	if err != nil {
		return nil, fmt.Errorf("invalid name for docker build: %w", err)
	}

	args := makeDockerVars(r.Args, nil)

	inputMap := make(map[string]bool)
	for _, input := range r.Input {
		inputMap[makePath(p, input)] = true
	}
	archInputMap := make(map[string]bool)
	for _, input := range r.ArchiveInput {
		archInputMap[makePath(p, input)] = true
	}

	prefixDir := r.PrefixDir
	if prefixDir == "." {
		prefixDir = p
	}

	var tarOut string
	if r.OutputTar {
		tarOut = imageTarOut(name)
	}

	return &imageBuild{
		name:           name,
		rule:           r,
		dockerfilePath: f,
		fromRuleSums:   fromRuleSums,
		inputs:         sortedStrList(inputMap),
		archInputs:     sortedStrList(archInputMap),
		prefixDir:      prefixDir,
		repoTag:        repoTag,
		args:           args,
		out:            imageSumOut(name),
		tarOut:         tarOut,
	}, nil
}

func (b *imageBuild) meta(env *env) (*buildRuleMeta, error) {
	dat := struct {
		Dockerfile string            // Know which one is the Dockerfile
		Args       map[string]string `json:",omitempty"`
		PrefixDir  string            `json:",omitempty"`
		OutputTar  bool              `json:",omitempty"`
	}{
		Dockerfile: b.dockerfilePath,
		Args:       b.args,
		PrefixDir:  b.prefixDir,
		OutputTar:  b.rule.OutputTar,
	}

	digest, err := makeDigest(ruleImageBuild, b.name, &dat)
	if err != nil {
		return nil, fmt.Errorf("digest: %w", err)
	}

	var deps []string
	deps = append(deps, b.dockerfilePath)
	deps = append(deps, b.fromRuleSums...)
	deps = append(deps, b.inputs...)
	deps = append(deps, b.archInputs...)

	outs := []string{b.out}
	if b.tarOut != "" {
		outs = append(outs, b.tarOut)
	}
	return &buildRuleMeta{
		name:      b.name,
		deps:      sortedStrList(makeStrSet(deps)),
		outs:      outs,
		imageOut: true,
		digest:    digest,
	}, nil
}

func (b *imageBuild) build(env *env, opts *buildOpts) error {
	dockerfileBytes, err := os.ReadFile(env.src(b.dockerfilePath))
	if err != nil {
		return fmt.Errorf("read Dockerfile: %w", err)
	}
	df := string(dockerfileBytes)

	ts := docker.NewTarStream(df)
	files := make(map[string]string)

	for _, in := range b.inputs {
		switch typ := env.nodeType(in); typ {
		case "":
			return fmt.Errorf("file %q not found", in)
		case nodeSrc:
			files[in] = env.src(in)
		case nodeOut:
			files[in] = env.out(in)
		case nodeRule:
			fileSet, err := referenceFileSetOut(env, in)
			if err != nil {
				return fmt.Errorf("input %q: %w", in, err)
			}
			fileSetFile := env.out(fileSet)
			var list []*fileStat
			if err := readJSONFile(fileSetFile, &list); err != nil {
				return fmt.Errorf("read file set %q: %w", in, err)
			}
			for _, f := range list {
				var fp string
				switch f.Type {
				case fileTypeSrc:
					fp = env.src(f.Name)
				case fileTypeOut:
					fp = env.out(f.Name)
				default:
					return fmt.Errorf(
						"invalid file type %q of %q ini set %q",
						f.Type, f.Name, in,
					)
				}
				files[f.Name] = fp
			}
		default:
			return fmt.Errorf("unknown type %q", typ)
		}
	}

	var names []string
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	prefixDir := b.prefixDir
	if prefixDir != "" && !strings.HasPrefix(prefixDir, "/") {
		prefixDir = prefixDir + "/"
	}

	for _, name := range names {
		tarName := name
		if prefixDir != "" {
			if !strings.HasPrefix(name, prefixDir) {
				continue
			}
			tarName = strings.TrimPrefix(name, prefixDir)
		}

		f := files[name]
		stat, err := os.Stat(f)
		if err != nil {
			return fmt.Errorf("stat file %q: %w", name, err)
		}
		mode := stat.Mode()
		if !mode.IsRegular() {
			return fmt.Errorf("%q is not a regular file", name)
		}
		ts.AddFile(tarName, tarutil.ModeMeta(int64(mode)&0777), f)
	}

	for _, ar := range b.archInputs {
		var fp string
		switch typ := env.nodeType(ar); typ {
		case nodeSrc:
			fp = env.src(ar)
		case nodeOut:
			fp = env.out(ar)
		default:
			return fmt.Errorf("unknown type %q", typ)
		}
		base := path.Base(ar)
		dir := path.Dir(ar)
		if dir == "." {
			dir = ""
		}
		if strings.HasSuffix(base, ".zip") {
			ts.AddZipFile(dir, fp)
		} else {
			return fmt.Errorf("unknown archive type %q", base)
		}
	}

	repo, tag := parseRepoTag(b.repoTag)
	rt := repoTag(repo, tag)

	config := &docker.BuildConfig{
		Files:    ts,
		Args:     b.args,
		UseCache: true, // TODO(h8liu): read from option.
	}
	if err := docker.BuildImageConfig(env.dock, rt, config); err != nil {
		return err
	}
	log.Printf("Built image: %s", rt)

	info, err := docker.InspectImage(env.dock, rt)
	if err != nil {
		return fmt.Errorf("inspect built image: %w", err)
	}

	sum := newImageSum(repo, tag, info.ID)

	out, err := env.prepareOut(b.out)
	if err != nil {
		return fmt.Errorf("prepare sum output: %w", err)
	}
	if err := writeJSONFile(out, sum); err != nil {
		return fmt.Errorf("write image sum: %w", err)
	}

	if b.tarOut != "" {
		log.Printf("Saving %s", b.tarOut)
		out, err := env.prepareOut(b.tarOut)
		if err != nil {
			return fmt.Errorf("prepare tar output: %w", err)
		}
		if err := docker.SaveImageGz(env.dock, sum.ID, out); err != nil {
			return fmt.Errorf("save image as tar: %w", err)
		}
	}

	return nil
}
