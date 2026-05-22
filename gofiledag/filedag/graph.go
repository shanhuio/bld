package filedag

import (
	"go/token"
	"sort"

	"golang.org/x/tools/go/packages"
)

// FileGraph is a directed graph of file-to-file references within a package.
// Each edge records one example reference (the first one encountered).
type FileGraph struct {
	Files []string                    // package files, sorted
	Edges map[string]map[string]*Edge // from -> to -> example
}

// Edge records one example reason file From depends on file To.
type Edge struct {
	Symbol string         // referenced symbol name
	UsePos token.Position // location of the reference
	DefPos token.Position // location of the symbol definition
}

func buildFileGraph(pkg *packages.Package) *FileGraph {
	fset := pkg.Fset
	fileSet := make(map[string]bool)
	for _, f := range pkg.Syntax {
		fileSet[fset.Position(f.Pos()).Filename] = true
	}

	edges := make(map[string]map[string]*Edge)
	for ident, obj := range pkg.TypesInfo.Uses {
		if obj.Pkg() != pkg.Types {
			continue
		}
		if obj.Pos() == token.NoPos {
			continue
		}
		usePos := fset.Position(ident.Pos())
		defPos := fset.Position(obj.Pos())
		if usePos.Filename == defPos.Filename {
			continue
		}
		if !fileSet[usePos.Filename] || !fileSet[defPos.Filename] {
			continue
		}
		toMap, ok := edges[usePos.Filename]
		if !ok {
			toMap = make(map[string]*Edge)
			edges[usePos.Filename] = toMap
		}
		if _, exists := toMap[defPos.Filename]; exists {
			continue
		}
		toMap[defPos.Filename] = &Edge{
			Symbol: obj.Name(),
			UsePos: usePos,
			DefPos: defPos,
		}
	}

	files := make([]string, 0, len(fileSet))
	for f := range fileSet {
		files = append(files, f)
	}
	sort.Strings(files)
	return &FileGraph{Files: files, Edges: edges}
}

// successors returns a sorted slice of the out-neighbors of v.
func (g *FileGraph) successors(v string) []string {
	var s []string
	for w := range g.Edges[v] {
		s = append(s, w)
	}
	sort.Strings(s)
	return s
}
