package caco3

import (
	"errors"
	"path/filepath"
	"testing"
)

func newTestKVTable(t *testing.T) *kvTable {
	t.Helper()
	f := filepath.Join(t.TempDir(), "kv.sqlite")
	tab, err := openKVTable(f)
	if err != nil {
		t.Fatalf("openKVTable: %v", err)
	}
	t.Cleanup(func() {
		if err := tab.db.Close(); err != nil {
			t.Errorf("close: %v", err)
		}
	})
	return tab
}

func TestKVTableReplaceAndGet(t *testing.T) {
	tab := newTestKVTable(t)

	type entry struct{ Name string }
	in := &entry{Name: "alpha"}
	if err := tab.replace("k1", in); err != nil {
		t.Fatalf("replace: %v", err)
	}

	out := new(entry)
	if err := tab.get("k1", out); err != nil {
		t.Fatalf("get: %v", err)
	}
	if out.Name != "alpha" {
		t.Errorf("get value = %q, want alpha", out.Name)
	}
}

func TestKVTableReplaceOverwrites(t *testing.T) {
	tab := newTestKVTable(t)
	type entry struct{ N int }

	if err := tab.replace("k", &entry{N: 1}); err != nil {
		t.Fatalf("first replace: %v", err)
	}
	if err := tab.replace("k", &entry{N: 2}); err != nil {
		t.Fatalf("second replace: %v", err)
	}

	out := new(entry)
	if err := tab.get("k", out); err != nil {
		t.Fatalf("get: %v", err)
	}
	if out.N != 2 {
		t.Errorf("N = %d, want 2 (latest replace)", out.N)
	}
}

func TestKVTableGetMissing(t *testing.T) {
	tab := newTestKVTable(t)
	var v struct{}
	err := tab.get("nope", &v)
	if !errors.Is(err, errKeyNotFound) {
		t.Errorf("want errKeyNotFound, got %v", err)
	}
}

func TestKVTableRemoveMissing(t *testing.T) {
	tab := newTestKVTable(t)
	if err := tab.remove("nope"); !errors.Is(err, errKeyNotFound) {
		t.Errorf("remove missing: want errKeyNotFound, got %v", err)
	}
}

func TestKVTableRemoveExisting(t *testing.T) {
	tab := newTestKVTable(t)
	type entry struct{ X int }

	if err := tab.replace("k", &entry{X: 1}); err != nil {
		t.Fatalf("replace: %v", err)
	}
	if err := tab.remove("k"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	// After remove, get must return errKeyNotFound.
	if err := tab.get("k", new(entry)); !errors.Is(err, errKeyNotFound) {
		t.Errorf("after remove, get returned %v, want errKeyNotFound", err)
	}
}

func TestKVTablePersistAcrossOpen(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "kv.sqlite")

	tab1, err := openKVTable(f)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	type entry struct{ Greeting string }
	if err := tab1.replace("g", &entry{Greeting: "hi"}); err != nil {
		t.Fatalf("replace: %v", err)
	}
	if err := tab1.db.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	tab2, err := openKVTable(f)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	defer tab2.db.Close()
	out := new(entry)
	if err := tab2.get("g", out); err != nil {
		t.Fatalf("get after reopen: %v", err)
	}
	if out.Greeting != "hi" {
		t.Errorf("Greeting = %q, want hi", out.Greeting)
	}
}
