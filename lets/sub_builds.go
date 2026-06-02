package lets

type subBuilds struct {
	dirs []string
	rule *SubBuilds
}

func newSubBuilds(_ *env, p string, v *SubBuilds) *subBuilds {
	var dirs []string
	for _, d := range v.Dirs {
		dirs = append(dirs, makeRelPath(p, d))
	}

	return &subBuilds{dirs: dirs, rule: v}
}

func (b *subBuilds) Dirs() []string { return b.dirs }
