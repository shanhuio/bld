package gofiledag

import "testing"

func TestCheckMethodsSameFile(t *testing.T) {
	pkg := parsePkg(t, map[string]string{
		"foo.go": `package test

type Foo struct{}

func (f *Foo) Bar() {}
`,
	})
	if got := checkMethods(pkg); len(got) != 0 {
		t.Errorf("got violations %v, want none", got)
	}
}

func TestCheckMethodsDifferentFile(t *testing.T) {
	pkg := parsePkg(t, map[string]string{
		"type.go":   "package test\n\ntype Foo struct{}\n",
		"method.go": "package test\n\nfunc (f *Foo) Bar() {}\n",
	})
	got := checkMethods(pkg)
	if len(got) != 1 {
		t.Fatalf("got %d violations, want 1: %v", len(got), got)
	}
	v := got[0]
	if v.Kind != "method_misplaced" {
		t.Errorf("kind = %q, want method_misplaced", v.Kind)
	}
	if v.Pos.Filename != "method.go" {
		t.Errorf("pos filename = %q, want method.go", v.Pos.Filename)
	}
}

func TestCheckMethodsPointerAndValueReceivers(t *testing.T) {
	pkg := parsePkg(t, map[string]string{
		"a.go": "package test\n\ntype T struct{}\n",
		"b.go": "package test\n\nfunc (t T) M1() {}\nfunc (t *T) M2() {}\n",
	})
	got := checkMethods(pkg)
	if len(got) != 2 {
		t.Fatalf("got %d violations, want 2: %v", len(got), got)
	}
}

func TestCheckMethodsGenericReceiver(t *testing.T) {
	pkg := parsePkg(t, map[string]string{
		"a.go": "package test\n\ntype Box[T any] struct{ V T }\n",
		"b.go": "package test\n\nfunc (b *Box[T]) Get() T { return b.V }\n",
	})
	got := checkMethods(pkg)
	if len(got) != 1 {
		t.Fatalf("got %d violations, want 1: %v", len(got), got)
	}
	if got[0].Pos.Filename != "b.go" {
		t.Errorf("filename = %q, want b.go", got[0].Pos.Filename)
	}
}
