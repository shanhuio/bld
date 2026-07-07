package gopkgdag

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

// LoadConfig configures package loading.
type LoadConfig struct {
	Dir    string   // working dir for loading
	Tags   []string // build tags
	GOOS   string
	GOARCH string
}

// Load loads the packages matching patterns, returning the root packages
// with their direct imports resolved.
func Load(cfg *LoadConfig, patterns []string) ([]*packages.Package, error) {
	if cfg == nil {
		cfg = new(LoadConfig)
	}
	env := os.Environ()
	if cfg.GOOS != "" {
		env = append(env, "GOOS="+cfg.GOOS)
	}
	if cfg.GOARCH != "" {
		env = append(env, "GOARCH="+cfg.GOARCH)
	}
	var buildFlags []string
	if len(cfg.Tags) > 0 {
		buildFlags = append(buildFlags, "-tags="+strings.Join(cfg.Tags, ","))
	}

	c := &packages.Config{
		Mode:       packages.NeedName | packages.NeedImports,
		Dir:        cfg.Dir,
		Env:        env,
		BuildFlags: buildFlags,
	}
	pkgs, err := packages.Load(c, patterns...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}
	return pkgs, nil
}
