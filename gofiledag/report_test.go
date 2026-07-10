package gofiledag

import (
	"bytes"
	"go/token"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
	"shanhu.io/std/graph"
)

// makeResult builds a Result in memory, without loading real packages, so
// the report formatting can be unit-tested in isolation. graphOf (from
// cycle_test.go) builds the FileGraph values these tests use.
func makeResult(pkgPath string, kind PassKind) *Result {
	return &Result{
		Pkg:  &packages.Package{PkgPath: pkgPath},
		Pass: &Pass{Kind: kind},
	}
}

func TestPrintCheckResults_methodViolation(t *testing.T) {
	r := makeResult("example.com/pkg", PassProd)
	r.Violations = []Violation{{
		Kind:    "method_misplaced",
		Pos:     token.Position{Filename: "/ws/foo.go", Line: 10, Column: 2},
		Message: "method Foo.Bar should be in foo.go",
	}}

	var buf bytes.Buffer
	fails := PrintCheckResults(&buf, []*Result{r}, "/ws")
	if fails != 1 {
		t.Fatalf("fails = %d, want 1", fails)
	}
	out := buf.String()
	for _, want := range []string{
		"example.com/pkg [production]:",
		"foo.go:10:2:",
		"method_misplaced",
		"method Foo.Bar should be in foo.go",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrintCheckResults_cycleViolation(t *testing.T) {
	r := makeResult("example.com/pkg", PassProd)
	r.Violations = []Violation{{
		Kind:    "cycle",
		Message: "file cycle of 2 files",
		Cycle: []CycleStep{{
			From:   "a.go",
			To:     "b.go",
			Symbol: "B",
			UsePos: token.Position{Filename: "/ws/a.go", Line: 5},
			DefPos: token.Position{Filename: "/ws/b.go", Line: 3},
		}},
	}}

	var buf bytes.Buffer
	fails := PrintCheckResults(&buf, []*Result{r}, "/ws")
	if fails != 1 {
		t.Fatalf("fails = %d, want 1", fails)
	}
	out := buf.String()
	for _, want := range []string{
		"cycle: file cycle of 2 files",
		"a.go -> b.go",
		"uses B",
		"a.go:5",
		"b.go:3",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrintCheckResults_skippedWarning(t *testing.T) {
	r := makeResult("example.com/pkg", PassProd)
	r.Skipped = "generated code"

	var buf bytes.Buffer
	fails := PrintCheckResults(&buf, []*Result{r}, "")
	if fails != 0 {
		t.Fatalf("fails = %d, want 0", fails)
	}
	out := buf.String()
	if !strings.Contains(out, "warning:") || !strings.Contains(out, "generated code") {
		t.Errorf("expected a skip warning, got:\n%s", out)
	}
}

func TestPrintReportResults_withViolation(t *testing.T) {
	r := makeResult("example.com/pkg", PassProd)
	r.Violations = []Violation{{
		Kind:    "method_misplaced",
		Pos:     token.Position{Filename: "/ws/foo.go", Line: 1, Column: 1},
		Message: "misplaced",
	}}

	var buf bytes.Buffer
	fails := PrintReportResults(&buf, []*Result{r}, "/ws")
	if fails != 1 {
		t.Fatalf("fails = %d, want 1", fails)
	}
	out := buf.String()
	if !strings.Contains(out, "method_misplaced") {
		t.Errorf("output missing violation:\n%s", out)
	}
	if strings.Contains(out, "layers") {
		t.Errorf("a failing pass should not print a graph:\n%s", out)
	}
}

func TestPrintReportResults_separatesResults(t *testing.T) {
	g := graphOf(
		[]string{"/ws/a.go", "/ws/b.go"},
		map[string][]string{"/ws/a.go": {"/ws/b.go"}},
	)
	r1 := makeResult("example.com/one", PassProd)
	r1.Graph = g
	r2 := makeResult("example.com/two", PassProd)
	r2.Graph = g

	var buf bytes.Buffer
	PrintReportResults(&buf, []*Result{r1, r2}, "")
	out := buf.String()
	for _, want := range []string{"example.com/one", "example.com/two"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	// A blank line must separate the two package sections.
	if !strings.Contains(out, "\n\n") {
		t.Errorf("results not separated by a blank line:\n%s", out)
	}
}

func TestWriteReport_layersAndEdges(t *testing.T) {
	g := graphOf(
		[]string{"/ws/a.go", "/ws/b.go"},
		map[string][]string{"/ws/a.go": {"/ws/b.go"}},
	)

	var buf bytes.Buffer
	writeReport(&buf, g)
	out := buf.String()
	for _, want := range []string{
		"layers (top = no deps):",
		"[0] b.go",
		"[1] a.go",
		"edges:",
		"a.go -> b.go",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestWriteReport_empty(t *testing.T) {
	for _, g := range []*FileGraph{nil, {}} {
		var buf bytes.Buffer
		writeReport(&buf, g)
		if !strings.Contains(buf.String(), "(no files)") {
			t.Errorf("empty graph should print (no files), got:\n%s", buf.String())
		}
	}
}

func TestComputeRanks_dag(t *testing.T) {
	g := graphOf(
		[]string{"a", "b", "c"},
		map[string][]string{"a": {"b"}, "b": {"c"}},
	)
	rank := computeRanks(g)
	for f, want := range map[string]int{"a": 2, "b": 1, "c": 0} {
		if rank[f] != want {
			t.Errorf("rank[%s] = %d, want %d", f, rank[f], want)
		}
	}
}

func TestComputeRanks_cycleTerminates(t *testing.T) {
	g := graphOf(
		[]string{"a", "b"},
		map[string][]string{"a": {"b"}, "b": {"a"}},
	)
	rank := computeRanks(g)
	if len(rank) != 2 {
		t.Errorf("got %d ranks, want 2: %+v", len(rank), rank)
	}
}

func TestRelPath(t *testing.T) {
	if got := relPath("/a/b/c.go", ""); got != "/a/b/c.go" {
		t.Errorf("relPath with empty cwd = %q, want unchanged", got)
	}
	if got := relPath("/a/b/c.go", "/a/b"); got != "c.go" {
		t.Errorf("relPath = %q, want c.go", got)
	}
}

func TestRelPos(t *testing.T) {
	pos := token.Position{Filename: "/a/b/c.go", Line: 3, Column: 5}
	if got, want := relPos(pos, "/a/b"), "c.go:3:5"; got != want {
		t.Errorf("relPos = %q, want %q", got, want)
	}
}

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
	// Node names drop the ".go" suffix.
	if got := v.Outs("a"); !reflect.DeepEqual(got, []string{"b"}) {
		t.Errorf("Outs(a) = %v, want [b]", got)
	}
	if got := v.Outs("b"); len(got) != 0 {
		t.Errorf("Outs(b) = %v, want none", got)
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
	if got := v.Outs("a_test"); !reflect.DeepEqual(got, []string{"a"}) {
		t.Errorf("Outs(a_test) = %v, want [a]", got)
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
