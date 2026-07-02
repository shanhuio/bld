package lets

import (
	"path/filepath"
	"testing"
)

func TestImageSumOut(t *testing.T) {
	if got, want := imageSumOut("foo/bar"), "foo/bar.imgsum"; got != want {
		t.Errorf("imageSumOut = %q, want %q", got, want)
	}
}

func TestImageTarOut(t *testing.T) {
	if got, want := imageTarOut("foo/bar"), "foo/bar.tar.gz"; got != want {
		t.Errorf("imageTarOut = %q, want %q", got, want)
	}
}

func TestNewImageSum(t *testing.T) {
	sum := newImageSum("cr.example.com/app", "latest", "sha256:abc")
	if sum.Repo != "cr.example.com/app" {
		t.Errorf("Repo = %q, want cr.example.com/app", sum.Repo)
	}
	if sum.Tag != "latest" {
		t.Errorf("Tag = %q, want latest", sum.Tag)
	}
	if sum.ID != "sha256:abc" {
		t.Errorf("ID = %q, want sha256:abc", sum.ID)
	}
	if sum.Origin != "" {
		t.Errorf("Origin = %q, want empty", sum.Origin)
	}
}

func TestLoadImageSum_roundTrip(t *testing.T) {
	f := filepath.Join(t.TempDir(), "app.imgsum")
	want := &imageSum{
		Repo:   "cr.example.com/app",
		Tag:    "v1",
		ID:     "sha256:abc",
		Origin: "alpine:3.23",
	}
	if err := writeJSONFile(f, want); err != nil {
		t.Fatalf("writeJSONFile: %v", err)
	}

	got, err := loadImageSum(f)
	if err != nil {
		t.Fatalf("loadImageSum: %v", err)
	}
	if *got != *want {
		t.Errorf("loadImageSum = %+v, want %+v", got, want)
	}
}

func TestLoadImageSum_missingFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "does-not-exist.imgsum")
	if _, err := loadImageSum(f); err == nil {
		t.Fatal("loadImageSum: want error for missing file, got nil")
	}
}
