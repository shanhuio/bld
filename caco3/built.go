package caco3

import (
	"fmt"

	"shanhu.io/bld/dock"
	"shanhu.io/bld/httputil"
)

type built struct {
	Outs    []*fileStat  `json:",omitempty"` // A list of outputs.
	Dockers []*dockerSum `json:",omitempty"` // A contaienr image.
}

func newBuilt(env *env, meta *buildRuleMeta) (*built, error) {
	b := new(built)
	for i, out := range meta.outs {
		if i == 0 && meta.dockerOut {
			sum, err := loadDockerSum(env.out(out))
			if err != nil {
				return nil, fmt.Errorf("read docker sum: %s: %w", out, err)
			}
			b.Dockers = append(b.Dockers, sum)
		}
		stat, err := newOutFileStat(env, out)
		if err != nil {
			return nil, fmt.Errorf("get output stat: %s: %w", out, err)
		}
		b.Outs = append(b.Outs, stat)
	}
	return b, nil
}

func checkSameBuilt(env *env, b *built) (bool, error) {
	for _, out := range b.Outs {
		same, err := sameFileStat(env, out)
		if err != nil {
			return false, fmt.Errorf(
				"check output stat of %q: %w", out.Name, err,
			)
		}
		if !same {
			return false, nil
		}
	}

	for _, d := range b.Dockers {
		repoTag := repoTag(d.Repo, d.Tag)
		info, err := dock.InspectImage(env.dock, repoTag)
		if err != nil {
			if httputil.IsNotFound(err) {
				return false, nil // Image not found.
			}
			return false, fmt.Errorf("inspect docker %s: %w", repoTag, err)
		}
		if info.ID != d.ID {
			return false, nil // Image ID changed.
		}
	}

	return true, nil
}
