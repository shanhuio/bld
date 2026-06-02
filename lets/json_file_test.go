package lets

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestJSONFile_roundTrip(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "data.json")

	type payload struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	in := payload{Name: "alpha", Count: 7}
	if err := writeJSONFile(file, in); err != nil {
		t.Fatalf("writeJSONFile: %v", err)
	}

	var out payload
	if err := readJSONFile(file, &out); err != nil {
		t.Fatalf("readJSONFile: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("roundtrip mismatch: got %+v, want %+v", out, in)
	}
}

func TestReadJSONFile_missing(t *testing.T) {
	var out struct{}
	err := readJSONFile(filepath.Join(t.TempDir(), "nope.json"), &out)
	if err == nil {
		t.Fatal("readJSONFile on missing file: want error, got nil")
	}
	if !os.IsNotExist(err) {
		t.Errorf("expected os.IsNotExist error, got %v", err)
	}
}

func TestReadJSONFile_malformed(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(file, []byte("not json"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	var out struct{}
	if err := readJSONFile(file, &out); err == nil {
		t.Error("readJSONFile on malformed input: want error, got nil")
	}
}

func TestWriteJSONFile_unmarshalable(t *testing.T) {
	file := filepath.Join(t.TempDir(), "out.json")
	// Channels cannot be marshaled to JSON.
	err := writeJSONFile(file, make(chan int))
	if err == nil {
		t.Error("writeJSONFile with unmarshalable value: want error, got nil")
	}
	if _, statErr := os.Stat(file); statErr == nil {
		t.Error("file should not have been created on marshal failure")
	}
}
