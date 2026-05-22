package gofiledag

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/packages"
)

// parsePkg parses an in-memory package from a map of file name to source. It
// returns a *packages.Package populated with the minimum fields the analyzer
// needs (Fset, Syntax, Types, TypesInfo, Name, ID, PkgPath).
func parsePkg(t *testing.T, files map[string]string) *packages.Package {
	t.Helper()
	fset := token.NewFileSet()
	var astFiles []*ast.File
	for name, src := range files {
		f, err := parser.ParseFile(fset, name, src, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		astFiles = append(astFiles, f)
	}
	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		Scopes:     make(map[ast.Node]*types.Scope),
	}
	conf := types.Config{Importer: importer.Default()}
	tp, err := conf.Check("test", fset, astFiles, info)
	if err != nil {
		t.Fatalf("type check: %v", err)
	}
	return &packages.Package{
		ID:        "test",
		Name:      "test",
		PkgPath:   "test",
		Fset:      fset,
		Syntax:    astFiles,
		Types:     tp,
		TypesInfo: info,
	}
}
