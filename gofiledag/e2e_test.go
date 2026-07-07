package gofiledag

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// loadFixture loads the named testdata fixture and returns all passes.
func loadFixture(t *testing.T, name string) []*Pass {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	passes, err := LoadPasses(&LoadConfig{Dir: dir}, []string{"."})
	if err != nil {
		t.Fatalf("LoadPasses(%s): %v", name, err)
	}
	return passes
}

// analyzeFixture returns the production-pass Result for the named fixture.
// Fails the test if the fixture has no production pass.
func analyzeFixture(t *testing.T, name string) *Result {
	t.Helper()
	for _, p := range loadFixture(t, name) {
		if p.Kind == PassProd {
			return Analyze(p)
		}
	}
	t.Fatalf("%s: no production pass", name)
	return nil
}

func TestEndToEnd_acyclic(t *testing.T) {
	r := analyzeFixture(t, "acyclic")
	if r.Skipped != "" {
		t.Fatalf("unexpected skip: %s", r.Skipped)
	}
	if len(r.Violations) != 0 {
		t.Fatalf("got %d violations, want 0: %+v", len(r.Violations), r.Violations)
	}
	if r.Graph == nil || len(r.Graph.Files) != 2 {
		t.Fatalf("graph: got %v, want 2 files", r.Graph)
	}
	// a.go must depend on b.go (uses B), but not the reverse.
	a := findFile(r.Graph.Files, "a.go")
	b := findFile(r.Graph.Files, "b.go")
	if r.Graph.Edges[a][b] == nil {
		t.Errorf("missing edge a.go -> b.go")
	}
	if r.Graph.Edges[b][a] != nil {
		t.Errorf("unexpected edge b.go -> a.go")
	}
}

func TestEndToEnd_cyclic(t *testing.T) {
	r := analyzeFixture(t, "cyclic")
	if len(r.Violations) != 1 {
		t.Fatalf("got %d violations, want 1: %+v", len(r.Violations), r.Violations)
	}
	v := r.Violations[0]
	if v.Kind != "cycle" {
		t.Errorf("kind = %q, want cycle", v.Kind)
	}
	if len(v.Cycle) != 2 {
		t.Errorf("cycle has %d steps, want 2", len(v.Cycle))
	}
	// Every step has a populated example symbol.
	for i, s := range v.Cycle {
		if s.Symbol == "" {
			t.Errorf("step %d missing symbol", i)
		}
	}
}

func TestEndToEnd_badMethod(t *testing.T) {
	r := analyzeFixture(t, "badmethod")
	if len(r.Violations) != 1 {
		t.Fatalf("got %d violations, want 1: %+v", len(r.Violations), r.Violations)
	}
	v := r.Violations[0]
	if v.Kind != "method_misplaced" {
		t.Errorf("kind = %q, want method_misplaced", v.Kind)
	}
	if !strings.HasSuffix(v.Pos.Filename, "method.go") {
		t.Errorf("violation filename = %q, want .../method.go", v.Pos.Filename)
	}
	if !strings.Contains(v.Message, "Foo") {
		t.Errorf("message %q does not mention Foo", v.Message)
	}
}

func TestEndToEnd_generated(t *testing.T) {
	r := analyzeFixture(t, "generated")
	if r.Skipped == "" {
		t.Fatalf("expected skip, got violations=%v", r.Violations)
	}
	if !strings.Contains(r.Skipped, "generated") {
		t.Errorf("skip reason = %q, want to mention 'generated'", r.Skipped)
	}
	if len(r.Violations) != 0 {
		t.Errorf("got %d violations on skipped pass, want 0", len(r.Violations))
	}
}

func TestEndToEnd_withTestsHasTwoPasses(t *testing.T) {
	passes := loadFixture(t, "withtests")
	kinds := make(map[PassKind]bool)
	for _, p := range passes {
		kinds[p.Kind] = true
	}
	if !kinds[PassProd] {
		t.Error("missing production pass")
	}
	if !kinds[PassInternalTest] {
		t.Error("missing with-tests pass")
	}
	// Each pass should analyze clean.
	for _, p := range passes {
		r := Analyze(p)
		if len(r.Violations) != 0 {
			t.Errorf("pass %s: got violations %+v", p.Kind, r.Violations)
		}
	}
}

func TestPrintCheckResults_cycle(t *testing.T) {
	r := analyzeFixture(t, "cyclic")
	var buf bytes.Buffer
	fails := PrintCheckResults(&buf, []*Result{r}, "")
	if fails != 1 {
		t.Errorf("fails = %d, want 1", fails)
	}
	out := buf.String()
	for _, want := range []string{"cycle", "a.go", "b.go", "uses"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrintCheckResults_clean(t *testing.T) {
	r := analyzeFixture(t, "acyclic")
	var buf bytes.Buffer
	fails := PrintCheckResults(&buf, []*Result{r}, "")
	if fails != 0 {
		t.Errorf("fails = %d, want 0", fails)
	}
	if buf.Len() != 0 {
		t.Errorf("check mode on clean pass should be silent, got:\n%s", buf.String())
	}
}

func TestPrintReportResults_clean(t *testing.T) {
	r := analyzeFixture(t, "acyclic")
	var buf bytes.Buffer
	fails := PrintReportResults(&buf, []*Result{r}, "")
	if fails != 0 {
		t.Errorf("fails = %d, want 0", fails)
	}
	out := buf.String()
	for _, want := range []string{"layers", "edges", "a.go", "b.go"} {
		if !strings.Contains(out, want) {
			t.Errorf("graph output missing %q:\n%s", want, out)
		}
	}
}

func TestPrintCheckResults_skipped(t *testing.T) {
	r := analyzeFixture(t, "generated")
	var buf bytes.Buffer
	fails := PrintCheckResults(&buf, []*Result{r}, "")
	if fails != 0 {
		t.Errorf("fails = %d, want 0 (skipped should not count)", fails)
	}
	out := buf.String()
	if !strings.Contains(out, "warning") || !strings.Contains(out, "generated") {
		t.Errorf("expected a generated-skip warning, got:\n%s", out)
	}
}

func TestAnalyzePasses_skipsWithTestsWhenProdFails(t *testing.T) {
	passes := loadFixture(t, "cyclicwithtest")

	// Sanity: the loader should produce both a production and a with-tests
	// pass for this fixture.
	var hasProd, hasWithTests bool
	for _, p := range passes {
		switch p.Kind {
		case PassProd:
			hasProd = true
		case PassInternalTest:
			hasWithTests = true
		}
	}
	if !hasProd || !hasWithTests {
		t.Fatalf("loader produced prod=%v with-tests=%v", hasProd, hasWithTests)
	}

	results := AnalyzePasses(passes)
	prodCount, withTestsCount := 0, 0
	for _, r := range results {
		switch r.Pass.Kind {
		case PassProd:
			prodCount++
			if len(r.Violations) == 0 {
				t.Errorf("prod pass should have a cycle violation")
			}
		case PassInternalTest:
			withTestsCount++
		}
	}
	if prodCount != 1 {
		t.Errorf("got %d prod results, want 1", prodCount)
	}
	if withTestsCount != 0 {
		t.Errorf("got %d with-tests results, want 0 (should be deduped)", withTestsCount)
	}
}

func TestAnalyzePasses_keepsWithTestsWhenProdClean(t *testing.T) {
	results := AnalyzePasses(loadFixture(t, "withtests"))
	var hasProd, hasWithTests bool
	for _, r := range results {
		switch r.Pass.Kind {
		case PassProd:
			hasProd = true
		case PassInternalTest:
			hasWithTests = true
		}
	}
	if !hasProd || !hasWithTests {
		t.Errorf("expected both passes when prod is clean; got prod=%v with-tests=%v",
			hasProd, hasWithTests)
	}
}

// findFile returns the full path in files whose basename matches name.
func findFile(files []string, name string) string {
	for _, f := range files {
		if filepath.Base(f) == name {
			return f
		}
	}
	return ""
}
