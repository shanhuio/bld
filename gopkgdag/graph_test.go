package gopkgdag

import (
	"reflect"
	"testing"

	"golang.org/x/tools/go/packages"
	"shanhu.io/std/graph"
)

// pkg builds a fake loaded package with the given import path, package name,
// and resolved imports (keyed by import path).
func pkg(path, name string, imports ...*packages.Package) *packages.Package {
	p := &packages.Package{PkgPath: path, Name: name}
	if len(imports) > 0 {
		p.Imports = make(map[string]*packages.Package)
		for _, imp := range imports {
			p.Imports[imp.PkgPath] = imp
		}
	}
	return p
}

func TestBuildGraph_intraModuleEdges(t *testing.T) {
	b := pkg("m/b", "b")
	c := pkg("m/c", "c")
	stdlib := pkg("fmt", "fmt")
	a := pkg("m/a", "a", b, c, stdlib)

	g, err := BuildGraph([]*packages.Package{a, b, c})
	if err != nil {
		t.Fatalf("BuildGraph: %v", err)
	}
	v, err := graph.NewViewer(g)
	if err != nil {
		t.Fatalf("NewViewer: %v", err)
	}

	if v.Len() != 3 {
		t.Fatalf("nodes = %d, want 3: %+v", v.Len(), g.Nodes)
	}
	// The external import (fmt) is not a loaded root, so it is not a node.
	if v.HasNode("fmt") {
		t.Error("external import fmt should not be a node")
	}
	if n := v.Node("m/a"); n == nil || n.Comment != "a" {
		t.Errorf("node m/a = %+v, want comment a", n)
	}
	// Edges to fmt are dropped; only intra-set edges remain, sorted.
	if got := v.Outs("m/a"); !reflect.DeepEqual(got, []string{"m/b", "m/c"}) {
		t.Errorf("Outs(m/a) = %v, want [m/b m/c]", got)
	}
	if got := v.Outs("m/b"); len(got) != 0 {
		t.Errorf("Outs(m/b) = %v, want none", got)
	}
}

func TestBuildGraph_skipsEmptyPkgPath(t *testing.T) {
	good := pkg("m/a", "a")
	broken := pkg("", "")

	g, err := BuildGraph([]*packages.Package{good, broken})
	if err != nil {
		t.Fatalf("BuildGraph: %v", err)
	}
	if len(g.Nodes) != 1 || g.Nodes[0].Name != "m/a" {
		t.Errorf("nodes = %+v, want only m/a", g.Nodes)
	}
}

func TestBuildGraph_empty(t *testing.T) {
	g, err := BuildGraph(nil)
	if err != nil {
		t.Fatalf("BuildGraph: %v", err)
	}
	if len(g.Nodes) != 0 || len(g.Edges) != 0 {
		t.Errorf("graph = %+v, want empty", g)
	}
}
