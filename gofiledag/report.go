package gofiledag

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

// PrintCheckResults writes a check-mode summary of results to w and returns
// the number of failing passes. Clean packages produce no output.
func PrintCheckResults(w io.Writer, results []*Result, cwd string) int {
	fails := 0
	for _, r := range results {
		if writeResult(w, r, cwd, false) {
			fails++
		}
	}
	return fails
}

// PrintReportResults writes the graph for each passing package, and the
// violations for each failing package. Returns the number of failing passes.
func PrintReportResults(w io.Writer, results []*Result, cwd string) int {
	fails := 0
	for i, r := range results {
		if i > 0 {
			fmt.Fprintln(w)
		}
		if writeResult(w, r, cwd, true) {
			fails++
		}
	}
	return fails
}

// writeResult writes a single result to w and reports whether it has
// violations. A skipped pass produces a warning. A pass with violations
// prints a header and each violation. For a clean pass, report mode prints
// a header and the file DAG, while check mode prints nothing.
func writeResult(w io.Writer, r *Result, cwd string, report bool) bool {
	header := fmt.Sprintf("%s [%s]", r.Pkg.PkgPath, r.Pass.Kind)
	if r.Skipped != "" {
		fmt.Fprintf(w, "warning: %s: skipped: %s\n", header, r.Skipped)
		return false
	}
	if len(r.Violations) > 0 {
		fmt.Fprintf(w, "%s:\n", header)
		for _, v := range r.Violations {
			writeViolation(w, &v, cwd)
		}
		return true
	}
	if report {
		fmt.Fprintf(w, "%s:\n", header)
		writeReport(w, r.Graph)
	}
	return false
}

func writeViolation(w io.Writer, v *Violation, cwd string) {
	switch v.Kind {
	case "cycle":
		fmt.Fprintf(w, "  cycle: %s\n", v.Message)
		for _, s := range v.Cycle {
			fmt.Fprintf(w, "    %s -> %s  (%s: uses %s, defined at %s:%d)\n",
				s.From, s.To,
				relPath(s.UsePos.Filename, cwd)+":"+itoa(s.UsePos.Line),
				s.Symbol,
				relPath(s.DefPos.Filename, cwd), s.DefPos.Line,
			)
		}
	default:
		fmt.Fprintf(w, "  %s: %s: %s\n",
			relPos(v.Pos, cwd), v.Kind, v.Message)
	}
}

// writeReport writes the file DAG as ranked layers followed by an adjacency
// list. Layers are computed by Kahn-style longest-path ranking.
func writeReport(w io.Writer, g *FileGraph) {
	if g == nil || len(g.Files) == 0 {
		fmt.Fprintln(w, "  (no files)")
		return
	}
	rank := computeRanks(g)
	maxRank := 0
	for _, r := range rank {
		if r > maxRank {
			maxRank = r
		}
	}
	fmt.Fprintln(w, "  layers (top = no deps):")
	for r := 0; r <= maxRank; r++ {
		var layer []string
		for _, f := range g.Files {
			if rank[f] == r {
				layer = append(layer, filepath.Base(f))
			}
		}
		sort.Strings(layer)
		fmt.Fprintf(w, "    [%d] %s\n", r, joinComma(layer))
	}
	fmt.Fprintln(w, "  edges:")
	for _, from := range g.Files {
		tos := g.successors(from)
		if len(tos) == 0 {
			continue
		}
		bases := make([]string, len(tos))
		for i, t := range tos {
			bases[i] = filepath.Base(t)
		}
		fmt.Fprintf(w, "    %s -> %s\n", filepath.Base(from), joinComma(bases))
	}
}

// computeRanks assigns each file a depth equal to the longest path of
// outgoing edges from it. Files with no outgoing edges have rank 0.
// For cyclic graphs the result is best-effort.
func computeRanks(g *FileGraph) map[string]int {
	rank := make(map[string]int)
	var visit func(f string) int
	inProgress := make(map[string]bool)
	visit = func(f string) int {
		if r, ok := rank[f]; ok {
			return r
		}
		if inProgress[f] {
			return 0
		}
		inProgress[f] = true
		r := 0
		for _, w := range g.successors(f) {
			if d := visit(w) + 1; d > r {
				r = d
			}
		}
		inProgress[f] = false
		rank[f] = r
		return r
	}
	for _, f := range g.Files {
		visit(f)
	}
	return rank
}

func itoa(n int) string { return fmt.Sprintf("%d", n) }

func joinComma(ss []string) string { return strings.Join(ss, ", ") }
