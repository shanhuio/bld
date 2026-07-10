package gofiledag

import (
	"reflect"
	"testing"

	"shanhu.io/std/graph"
)

func TestBuildGraph(t *testing.T) {
	r := makeResult("example.com/pkg", PassProd)
	r.Graph = graphOf(
		[]string{"/ws/a.go", "/ws/b.go"},
		map[string][]string{"/ws/a.go": {"/ws/b.go"}},
	)

	g, err := buildGraph([]*Result{r}, "/ws")
	if err != nil {
		t.Fatalf("buildGraph: %v", err)
	}
	if g.Name != "example.com/pkg" {
		t.Errorf("graph Name = %q, want example.com/pkg", g.Name)
	}
	v, err := graph.NewViewer(g)
	if err != nil {
		t.Fatalf("NewViewer: %v", err)
	}
	if v.Len() != 2 {
		t.Fatalf("nodes = %d, want 2", v.Len())
	}
	// Node names drop the ".go" suffix. a.go uses b.go, so the edge points
	// from the dependency b to the dependent a.
	if got := v.Outs("b"); !reflect.DeepEqual(got, []string{"a"}) {
		t.Errorf("Outs(b) = %v, want [a]", got)
	}
	if got := v.Outs("a"); len(got) != 0 {
		t.Errorf("Outs(a) = %v, want none", got)
	}
}

func TestBuildGraph_multiplePackagesError(t *testing.T) {
	a := makeResult("example.com/a", PassProd)
	a.Graph = graphOf([]string{"/ws/a/x.go"}, nil)
	b := makeResult("example.com/b", PassProd)
	b.Graph = graphOf([]string{"/ws/b/y.go"}, nil)

	if _, err := buildGraph([]*Result{a, b}, "/ws"); err == nil {
		t.Fatal("buildGraph: want error for multiple packages, got nil")
	}
}

// TestBuildGraph_dedupsAcrossPasses checks that a file shared by the
// production and internal-test passes of one package becomes a single node.
func TestBuildGraph_dedupsAcrossPasses(t *testing.T) {
	prod := makeResult("example.com/pkg", PassProd)
	prod.Graph = graphOf([]string{"/ws/a.go"}, nil)

	test := makeResult("example.com/pkg", PassInternalTest)
	test.Graph = graphOf(
		[]string{"/ws/a.go", "/ws/a_test.go"},
		map[string][]string{"/ws/a_test.go": {"/ws/a.go"}},
	)

	g, err := buildGraph([]*Result{prod, test}, "/ws")
	if err != nil {
		t.Fatalf("buildGraph: %v", err)
	}
	if g.Name != "example.com/pkg" {
		t.Errorf("graph Name = %q, want example.com/pkg", g.Name)
	}
	v, err := graph.NewViewer(g)
	if err != nil {
		t.Fatalf("NewViewer: %v", err)
	}
	if v.Len() != 2 {
		t.Fatalf("nodes = %d, want 2 (a.go merged): %+v", v.Len(), g.Nodes)
	}
	// a_test.go uses a.go, so the edge points from a to a_test.
	if got := v.Outs("a"); !reflect.DeepEqual(got, []string{"a_test"}) {
		t.Errorf("Outs(a) = %v, want [a_test]", got)
	}
}

func TestBuildGraph_skipsNilGraph(t *testing.T) {
	skipped := makeResult("example.com/pkg", PassProd)
	skipped.Skipped = "generated"

	g, err := buildGraph([]*Result{skipped}, "/ws")
	if err != nil {
		t.Fatalf("buildGraph: %v", err)
	}
	if len(g.Nodes) != 0 || len(g.Edges) != 0 {
		t.Errorf("graph = %+v, want empty for skipped result", g)
	}
}
