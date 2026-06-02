package gofiledag

import (
	"fmt"
	"go/token"
	"io"
	"path/filepath"
	"sort"
)

// PrintCheckResults writes a check-mode summary of results to w and returns
// the number of failing passes.
func PrintCheckResults(w io.Writer, results []*Result, cwd string) int {
	fails := 0
	for _, r := range results {
		header := fmt.Sprintf("%s [%s]", r.Pkg.PkgPath, r.Pass.Kind)
		if r.Skipped != "" {
			fmt.Fprintf(w, "warning: %s: skipped: %s\n", header, r.Skipped)
			continue
		}
		if len(r.Violations) == 0 {
			continue
		}
		fails++
		fmt.Fprintf(w, "%s:\n", header)
		for _, v := range r.Violations {
			writeViolation(w, &v, cwd)
		}
	}
	return fails
}

// PrintGraphResults writes the graph for each passing package, and the
// violations for each failing package. Returns the number of failing passes.
func PrintGraphResults(w io.Writer, results []*Result, cwd string) int {
	fails := 0
	for i, r := range results {
		if i > 0 {
			fmt.Fprintln(w)
		}
		header := fmt.Sprintf("%s [%s]", r.Pkg.PkgPath, r.Pass.Kind)
		if r.Skipped != "" {
			fmt.Fprintf(w, "warning: %s: skipped: %s\n", header, r.Skipped)
			continue
		}
		fmt.Fprintf(w, "%s:\n", header)
		if len(r.Violations) > 0 {
			fails++
			for _, v := range r.Violations {
				writeViolation(w, &v, cwd)
			}
			continue
		}
		writeGraph(w, r.Graph)
	}
	return fails
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

// writeGraph writes the file DAG as ranked layers followed by an adjacency
// list. Layers are computed by Kahn-style longest-path ranking.
func writeGraph(w io.Writer, g *FileGraph) {
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

func relPath(p, cwd string) string {
	if cwd == "" {
		return p
	}
	r, err := filepath.Rel(cwd, p)
	if err != nil {
		return p
	}
	return r
}

func relPos(p token.Position, cwd string) string {
	return fmt.Sprintf("%s:%d:%d", relPath(p.Filename, cwd), p.Line, p.Column)
}

func itoa(n int) string { return fmt.Sprintf("%d", n) }

func joinComma(ss []string) string {
	s := ""
	for i, v := range ss {
		if i > 0 {
			s += ", "
		}
		s += v
	}
	return s
}
