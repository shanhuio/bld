package gopkgdag

import (
	"sort"

	"golang.org/x/tools/go/packages"
	"shanhu.io/std/graph"
)

// BuildGraph builds the package dependency graph of pkgs. The graph is
// titled with the module path. Each package is a node named by its import
// path and commented with its package name. Each edge points from an
// imported package to the package that imports it (dependency to
// dependent), so the graph flows from leaf packages upward. Imports outside
// pkgs (standard library, external modules) are omitted. Nodes and edges
// are emitted in import-path order.
func BuildGraph(pkgs []*packages.Package) (*graph.Graph, error) {
	var sorted []*packages.Package
	for _, p := range pkgs {
		if p.PkgPath == "" {
			continue
		}
		sorted = append(sorted, p)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].PkgPath < sorted[j].PkgPath
	})

	inSet := make(map[string]bool, len(sorted))
	for _, p := range sorted {
		inSet[p.PkgPath] = true
	}

	b := graph.NewBuilder()
	b.SetName(moduleName(sorted))
	for _, p := range sorted {
		if b.HasNode(p.PkgPath) {
			continue
		}
		if _, err := b.AddNode(p.PkgPath, p.Name); err != nil {
			return nil, err
		}
	}
	for _, p := range sorted {
		var imps []string
		for path := range p.Imports {
			if inSet[path] {
				imps = append(imps, path)
			}
		}
		sort.Strings(imps)
		for _, imp := range imps {
			// Edge points from the dependency to the dependent.
			if err := b.AddEdge(imp, p.PkgPath); err != nil {
				return nil, err
			}
		}
	}
	return b.Build(), nil
}

// moduleName returns the path of the main module the packages belong to, or
// the first module path found, or "" when no module information is present.
func moduleName(pkgs []*packages.Package) string {
	first := ""
	for _, p := range pkgs {
		if p.Module == nil || p.Module.Path == "" {
			continue
		}
		if p.Module.Main {
			return p.Module.Path
		}
		if first == "" {
			first = p.Module.Path
		}
	}
	return first
}
