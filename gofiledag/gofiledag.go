package gofiledag

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"regexp"
	"sort"

	"golang.org/x/tools/go/packages"
)

// Analyze runs all checks on a pass.
func Analyze(p *Pass) *Result {
	r := &Result{Pass: p, Pkg: p.Pkg}
	if hasGenerated(p.Pkg) {
		r.Skipped = "contains generated files"
		return r
	}
	r.Violations = append(r.Violations, checkMethods(p.Pkg)...)
	r.Graph = buildFileGraph(p.Pkg)
	if cycle := findFirstCycle(r.Graph); cycle != nil {
		r.Violations = append(r.Violations, cycleViolation(p.Pkg, r.Graph, cycle))
	}
	sort.SliceStable(r.Violations, func(i, j int) bool {
		a, b := r.Violations[i].Pos, r.Violations[j].Pos
		if a.Filename != b.Filename {
			return a.Filename < b.Filename
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Column < b.Column
	})
	return r
}

func cycleViolation(pkg *packages.Package, g *FileGraph, chain []string) Violation {
	var steps []CycleStep
	for i := 0; i+1 < len(chain); i++ {
		from, to := chain[i], chain[i+1]
		edge := g.Edges[from][to]
		step := CycleStep{
			From: filepath.Base(from),
			To:   filepath.Base(to),
		}
		if edge != nil {
			step.Symbol = edge.Symbol
			step.UsePos = edge.UsePos
			step.DefPos = edge.DefPos
		}
		steps = append(steps, step)
	}
	pos := pkg.Fset.Position(token.NoPos)
	if len(steps) > 0 {
		pos = steps[0].UsePos
	}
	msg := fmt.Sprintf("file cycle of %d files in package %s", len(chain)-1, pkg.PkgPath)
	return Violation{
		Kind:    "cycle",
		PkgID:   pkg.ID,
		Pos:     pos,
		Message: msg,
		Cycle:   steps,
	}
}

// hasGenerated returns true if any source file in the package is generated.
func hasGenerated(pkg *packages.Package) bool {
	for _, f := range pkg.Syntax {
		if isGenerated(f) {
			return true
		}
	}
	return false
}

// generatedRE matches the standard Go generated-file marker. The marker must
// appear as a standalone single-line comment before the package clause.
var generatedRE = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)

func isGenerated(f *ast.File) bool {
	for _, cg := range f.Comments {
		if cg.End() > f.Package {
			break
		}
		for _, c := range cg.List {
			if generatedRE.MatchString(c.Text) {
				return true
			}
		}
	}
	return false
}
