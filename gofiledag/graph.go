package gofiledag

import (
	"fmt"
	"strings"

	"shanhu.io/std/graph"
)

// nodeName is the graph node name for a file: its path relative to cwd with
// the ".go" suffix dropped.
func nodeName(f, cwd string) string {
	return strings.TrimSuffix(relPath(f, cwd), ".go")
}

// buildGraph builds the file DAG of a single package as a graph.Graph
// titled with its import path. Nodes are files named by their path relative
// to cwd without the ".go" suffix. Each edge points from a referenced file
// to the file that references it (dependency to dependent), so the graph
// flows from leaf files upward. The package's several passes (production and
// internal-test) are merged, and skipped results (no graph) are ignored. It
// returns an error if the results span more than one package, since a single
// graph describes a single package.
func buildGraph(results []*Result, cwd string) (*graph.Graph, error) {
	var pkgPath string
	havePkg := false
	var graphed []*Result
	for _, r := range results {
		if r.Graph == nil {
			continue
		}
		if !havePkg {
			pkgPath, havePkg = r.Pkg.PkgPath, true
		} else if r.Pkg.PkgPath != pkgPath {
			return nil, fmt.Errorf(
				"graph output requires a single package, found %q and %q",
				pkgPath, r.Pkg.PkgPath,
			)
		}
		graphed = append(graphed, r)
	}

	b := graph.NewBuilder()
	b.SetName(pkgPath)
	for _, r := range graphed {
		for _, f := range r.Graph.Files {
			name := nodeName(f, cwd)
			if b.HasNode(name) {
				continue
			}
			if _, err := b.AddNode(name, ""); err != nil {
				return nil, err
			}
		}
	}
	for _, r := range graphed {
		for _, from := range r.Graph.Files {
			fromName := nodeName(from, cwd)
			for _, to := range r.Graph.successors(from) {
				// Edge points from the dependency to the dependent.
				if err := b.AddEdge(nodeName(to, cwd), fromName); err != nil {
					return nil, err
				}
			}
		}
	}
	return b.Build(), nil
}
