package gofiledag

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// checkMethods returns one violation for every method whose receiver type is
// defined in a different file from the method declaration.
func checkMethods(pkg *packages.Package) []Violation {
	var vs []Violation
	fset := pkg.Fset
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv == nil || len(fd.Recv.List) == 0 {
				continue
			}
			recvIdent := receiverTypeIdent(fd.Recv.List[0].Type)
			if recvIdent == nil {
				continue
			}
			obj, ok := pkg.TypesInfo.Uses[recvIdent].(*types.TypeName)
			if !ok {
				continue
			}
			if obj.Pkg() != pkg.Types {
				continue // alias or built-in; shouldn't happen for methods
			}
			methodPos := fset.Position(fd.Pos())
			typePos := fset.Position(obj.Pos())
			if methodPos.Filename == typePos.Filename {
				continue
			}
			vs = append(vs, Violation{
				Kind:    "method_misplaced",
				PkgID:   pkg.ID,
				Pos:     methodPos,
				Message: methodMisplacedMessage(fd, obj, methodPos, typePos),
			})
		}
	}
	return vs
}

// receiverTypeIdent peels *T, T[X], *T[X], T[X,Y], *T[X,Y] down to the
// underlying type-name ident T. Returns nil for unrecognized forms.
func receiverTypeIdent(expr ast.Expr) *ast.Ident {
	for {
		switch e := expr.(type) {
		case *ast.StarExpr:
			expr = e.X
		case *ast.IndexExpr:
			expr = e.X
		case *ast.IndexListExpr:
			expr = e.X
		case *ast.Ident:
			return e
		default:
			return nil
		}
	}
}

func methodMisplacedMessage(
	fd *ast.FuncDecl, typeObj types.Object,
	methodPos, typePos token.Position,
) string {
	recv := exprString(fd.Recv.List[0].Type)
	return fmt.Sprintf(
		"method (%s).%s in %s; type %s defined at %s:%d",
		recv, fd.Name.Name,
		filepath.Base(methodPos.Filename),
		typeObj.Name(),
		filepath.Base(typePos.Filename), typePos.Line,
	)
}

// exprString renders a receiver type expression like "*Foo" or "Foo[T]".
func exprString(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.StarExpr:
		return "*" + exprString(v.X)
	case *ast.IndexExpr:
		return exprString(v.X) + "[" + exprString(v.Index) + "]"
	case *ast.IndexListExpr:
		s := exprString(v.X) + "["
		for i, ix := range v.Indices {
			if i > 0 {
				s += ", "
			}
			s += exprString(ix)
		}
		return s + "]"
	default:
		return "?"
	}
}
