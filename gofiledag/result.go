package gofiledag

import (
	"go/token"

	"golang.org/x/tools/go/packages"
)

// Result is the outcome of analyzing one Pass.
type Result struct {
	Pkg        *packages.Package
	Pass       *Pass
	Skipped    string // non-empty if the pass was skipped, explains why
	Violations []Violation
	Graph      *FileGraph
}

// Violation is one finding from the analyzer.
type Violation struct {
	Kind    string         // "method_misplaced" or "cycle"
	PkgID   string         // package ID
	Pos     token.Position // primary position
	Message string         // one-line summary
	Cycle   []CycleStep    // populated when Kind == "cycle"
}

// CycleStep describes one edge in a reported cycle, with one example
// symbol that demonstrates the dependency.
type CycleStep struct {
	From   string         // source file (basename for display)
	To     string         // target file (basename for display)
	Symbol string         // example referenced symbol
	UsePos token.Position // where the symbol is referenced
	DefPos token.Position // where the symbol is defined
}
