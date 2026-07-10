package gofiledag

import (
	"bytes"
	"go/token"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
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
