package filedag

import "testing"

func TestBuildFileGraphSimpleEdge(t *testing.T) {
	pkg := parsePkg(t, map[string]string{
		"a.go": "package test\n\nfunc UseB() int { return B }\n",
		"b.go": "package test\n\nconst B = 1\n",
	})
	g := buildFileGraph(pkg)
	if got := g.Edges["a.go"]["b.go"]; got == nil {
		t.Fatal("missing edge a.go -> b.go")
	} else if got.Symbol != "B" {
		t.Errorf("edge symbol = %q, want B", got.Symbol)
	}
	if _, ok := g.Edges["b.go"]; ok {
		t.Errorf("unexpected outgoing edges from b.go: %v", g.Edges["b.go"])
	}
}

func TestBuildFileGraphNoSelfEdges(t *testing.T) {
	pkg := parsePkg(t, map[string]string{
		"a.go": "package test\n\nconst X = 1\n\nfunc UseX() int { return X }\n",
	})
	g := buildFileGraph(pkg)
	if len(g.Edges) != 0 {
		t.Errorf("expected no edges, got %v", g.Edges)
	}
}

func TestBuildFileGraphCycle(t *testing.T) {
	pkg := parsePkg(t, map[string]string{
		"a.go": "package test\n\nfunc UseB() int { return B }\n\nconst A = 2\n",
		"b.go": "package test\n\nfunc UseA() int { return A }\n\nconst B = 1\n",
	})
	g := buildFileGraph(pkg)
	if g.Edges["a.go"]["b.go"] == nil {
		t.Error("missing edge a.go -> b.go")
	}
	if g.Edges["b.go"]["a.go"] == nil {
		t.Error("missing edge b.go -> a.go")
	}
}
