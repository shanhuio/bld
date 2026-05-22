package gofiledag

import (
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestClassifyPackage(t *testing.T) {
	for _, c := range []struct {
		name     string
		pkg      *packages.Package
		wantSkip bool
		wantKind PassKind
	}{
		{
			name: "production",
			pkg:  &packages.Package{ID: "foo", Name: "foo", PkgPath: "foo"},
			wantKind: PassProd,
		},
		{
			name: "internal test variant",
			pkg:  &packages.Package{ID: "foo [foo.test]", Name: "foo", PkgPath: "foo"},
			wantKind: PassInternalTest,
		},
		{
			name: "external test package",
			pkg: &packages.Package{
				ID: "foo_test [foo.test]", Name: "foo_test", PkgPath: "foo_test",
			},
			wantKind: PassExternalTest,
		},
		{
			name: "synthetic test binary main",
			pkg: &packages.Package{
				ID: "foo.test", Name: "main", PkgPath: "foo.test",
			},
			wantSkip: true,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			gotKind, gotSkip := classifyPackage(c.pkg)
			if gotSkip != c.wantSkip {
				t.Errorf("skip = %v, want %v", gotSkip, c.wantSkip)
			}
			if !c.wantSkip && gotKind != c.wantKind {
				t.Errorf("kind = %v, want %v", gotKind, c.wantKind)
			}
		})
	}
}
