package gofiledag

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

// PassKind identifies which kind of pass a Pass represents.
type PassKind int

const (
	// PassProd is the production-only pass.
	PassProd PassKind = iota
	// PassInternalTest is the pass over production plus internal test
	// files (the `package foo` test files).
	PassInternalTest
	// PassExternalTest is the pass over external test files
	// (the `package foo_test` test files).
	PassExternalTest
)

func (k PassKind) String() string {
	switch k {
	case PassProd:
		return "production"
	case PassInternalTest:
		return "with-tests"
	case PassExternalTest:
		return "external-test"
	}
	return "?"
}

// Pass is one analysis target: a loaded package and its kind.
type Pass struct {
	Kind PassKind
	Pkg  *packages.Package
}

// LoadPasses loads the requested patterns and groups them into analysis passes.
func LoadPasses(cfg *LoadConfig, patterns []string) ([]*Pass, error) {
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
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes |
			packages.NeedImports,
		Tests:      true,
		Dir:        cfg.Dir,
		Env:        env,
		BuildFlags: buildFlags,
	}
	pkgs, err := packages.Load(c, patterns...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	var passes []*Pass
	for _, p := range pkgs {
		kind, skip := classifyPackage(p)
		if skip {
			continue
		}
		passes = append(passes, &Pass{Kind: kind, Pkg: p})
	}
	return passes, nil
}

// classifyPackage decides what kind of pass a loaded package corresponds to,
// or whether it should be skipped (the synthetic test-binary main package).
func classifyPackage(p *packages.Package) (PassKind, bool) {
	if p.Name == "main" && strings.HasSuffix(p.ID, ".test") {
		return 0, true
	}
	if strings.HasSuffix(p.PkgPath, "_test") {
		return PassExternalTest, false
	}
	if strings.Contains(p.ID, " [") {
		return PassInternalTest, false
	}
	return PassProd, false
}
