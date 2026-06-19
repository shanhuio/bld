package lets

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func writeFile(t *testing.T, p, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(p), err)
	}
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
}

func TestFileSetOut(t *testing.T) {
	if got, want := fileSetOut("foo/bar"), "foo/bar.fileset"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestListAllFiles_basic(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.txt"), "a")
	writeFile(t, filepath.Join(dir, "sub", "b.txt"), "b")

	got, err := listAllFiles(dir)
	if err != nil {
		t.Fatalf("listAllFiles: %v", err)
	}
	sort.Strings(got)
	want := []string{
		filepath.Join(dir, "a.txt"),
		filepath.Join(dir, "sub", "b.txt"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestListAllFiles_ignores(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "keep.txt"), "k")
	writeFile(t, filepath.Join(dir, ".gitignore"), "g")
	writeFile(t, filepath.Join(dir, "COPYING"), "c")
	writeFile(t, filepath.Join(dir, "tags"), "t")
	writeFile(t, filepath.Join(dir, ".DS_Store"), "d")
	writeFile(t, filepath.Join(dir, "BUILD.lets"), "b")
	writeFile(t, filepath.Join(dir, ".letsroot"), "")
	writeFile(t, filepath.Join(dir, ".git", "config"), "x")
	writeFile(t, filepath.Join(dir, "_", "src", "dep", "d.go"), "x")
	writeFile(t, filepath.Join(dir, "sub", "regular.go"), "r")

	got, err := listAllFiles(dir)
	if err != nil {
		t.Fatalf("listAllFiles: %v", err)
	}
	sort.Strings(got)
	want := []string{
		filepath.Join(dir, "keep.txt"),
		filepath.Join(dir, "sub", "regular.go"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestListAllFiles_includesSymlinks(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "real.txt"), "r")
	if err := os.Symlink("real.txt", filepath.Join(dir, "link")); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	got, err := listAllFiles(dir)
	if err != nil {
		t.Fatalf("listAllFiles: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d entries, want 2 (real + link): %v", len(got), got)
	}
}

func TestFileSet_selectGlob(t *testing.T) {
	e, srcDir, _ := newTestEnv(t)
	writeFile(t, filepath.Join(srcDir, "pkg", "a.go"), "a")
	writeFile(t, filepath.Join(srcDir, "pkg", "b.go"), "b")
	writeFile(t, filepath.Join(srcDir, "pkg", "c.txt"), "c")

	fs, err := newFileSet(e, "pkg", &FileSet{
		Name:   "sources",
		Select: []string{"*.go"},
	})
	if err != nil {
		t.Fatalf("newFileSet: %v", err)
	}
	want := []string{"pkg/a.go", "pkg/b.go"}
	if !reflect.DeepEqual(fs.files, want) {
		t.Errorf("files = %v, want %v", fs.files, want)
	}
	if fs.name != "pkg/sources" {
		t.Errorf("name = %q, want pkg/sources", fs.name)
	}
	if fs.out != "pkg/sources.fileset" {
		t.Errorf("out = %q, want pkg/sources.fileset", fs.out)
	}
}

func TestFileSet_selectAllRecursive(t *testing.T) {
	e, srcDir, _ := newTestEnv(t)
	writeFile(t, filepath.Join(srcDir, "pkg", "a.go"), "a")
	writeFile(t, filepath.Join(srcDir, "pkg", "sub", "b.go"), "b")
	writeFile(t, filepath.Join(srcDir, "pkg", ".git", "hidden"), "x")

	fs, err := newFileSet(e, "pkg", &FileSet{
		Name:   "all",
		Select: []string{"**"},
	})
	if err != nil {
		t.Fatalf("newFileSet: %v", err)
	}
	want := []string{"pkg/a.go", "pkg/sub/b.go"}
	if !reflect.DeepEqual(fs.files, want) {
		t.Errorf("files = %v, want %v", fs.files, want)
	}
}

func TestFileSet_ignoreGlobAndDir(t *testing.T) {
	e, srcDir, _ := newTestEnv(t)
	writeFile(t, filepath.Join(srcDir, "pkg", "a.go"), "a")
	writeFile(t, filepath.Join(srcDir, "pkg", "a_test.go"), "x")
	writeFile(t, filepath.Join(srcDir, "pkg", "vendor", "v.go"), "v")

	fs, err := newFileSet(e, "pkg", &FileSet{
		Name:   "src",
		Select: []string{"**"},
		Ignore: []string{"*_test.go", "vendor/"},
	})
	if err != nil {
		t.Fatalf("newFileSet: %v", err)
	}
	want := []string{"pkg/a.go"}
	if !reflect.DeepEqual(fs.files, want) {
		t.Errorf("files = %v, want %v", fs.files, want)
	}
}

func TestFileSet_explicitFiles(t *testing.T) {
	e, srcDir, _ := newTestEnv(t)
	writeFile(t, filepath.Join(srcDir, "pkg", "a.go"), "a")
	writeFile(t, filepath.Join(srcDir, "pkg", "b.go"), "b")

	fs, err := newFileSet(e, "pkg", &FileSet{
		Name:  "two",
		Files: []string{"a.go", "b.go"},
	})
	if err != nil {
		t.Fatalf("newFileSet: %v", err)
	}
	want := []string{"pkg/a.go", "pkg/b.go"}
	if !reflect.DeepEqual(fs.files, want) {
		t.Errorf("files = %v, want %v", fs.files, want)
	}
}

func TestFileSet_selectNoMatch(t *testing.T) {
	e, srcDir, _ := newTestEnv(t)
	writeFile(t, filepath.Join(srcDir, "pkg", "a.go"), "a")

	_, err := newFileSet(e, "pkg", &FileSet{
		Name:   "none",
		Select: []string{"*.c"},
	})
	if err == nil {
		t.Fatal("want error for select with no matches, got nil")
	}
}

func TestFileSet_metaDigestAndDeps(t *testing.T) {
	e, srcDir, _ := newTestEnv(t)
	writeFile(t, filepath.Join(srcDir, "pkg", "a.go"), "a")
	writeFile(t, filepath.Join(srcDir, "pkg", "b.go"), "b")

	fs, err := newFileSet(e, "pkg", &FileSet{
		Name:    "src",
		Select:  []string{"*.go"},
		Include: []string{"other:set"},
	})
	if err != nil {
		t.Fatalf("newFileSet: %v", err)
	}
	meta, err := fs.meta(e)
	if err != nil {
		t.Fatalf("meta: %v", err)
	}
	if meta.name != "pkg/src" {
		t.Errorf("meta.name = %q, want pkg/src", meta.name)
	}
	wantOuts := []string{"pkg/src.fileset"}
	if !reflect.DeepEqual(meta.outs, wantOuts) {
		t.Errorf("meta.outs = %v, want %v", meta.outs, wantOuts)
	}
	wantDeps := []string{"pkg/a.go", "pkg/b.go", "other:set"}
	if !reflect.DeepEqual(meta.deps, wantDeps) {
		t.Errorf("meta.deps = %v, want %v", meta.deps, wantDeps)
	}
	if meta.digest == "" {
		t.Error("meta.digest empty, want non-empty")
	}
}

func TestReferenceFileSetOut(t *testing.T) {
	e := &env{
		nodeType: func(name string) string {
			switch name {
			case "ok":
				return nodeRule
			case "not_rule":
				return nodeSrc
			}
			return ""
		},
		ruleType: func(name string) string {
			if name == "ok" {
				return ruleFileSet
			}
			return "docker_build"
		},
	}

	for _, c := range []struct {
		name    string
		want    string
		wantErr bool
	}{
		{name: "ok", want: "ok.fileset"},
		{name: "not_rule", wantErr: true},
		{name: "missing", wantErr: true},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := referenceFileSetOut(e, c.name)
			if c.wantErr {
				if err == nil {
					t.Errorf("want error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Errorf("got err %v", err)
			}
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}

	// And one case where the rule is not a fileSet.
	e.ruleType = func(name string) string { return "docker_build" }
	if _, err := referenceFileSetOut(e, "ok"); err == nil {
		t.Error("want error when rule type is not file_set")
	}
}

func TestFileSet_buildAggregatesStats(t *testing.T) {
	e, srcDir, outDir := newTestEnv(t)
	writeFile(t, filepath.Join(srcDir, "pkg", "a.go"), "a")
	writeFile(t, filepath.Join(srcDir, "pkg", "b.go"), "bb")

	fs, err := newFileSet(e, "pkg", &FileSet{
		Name:   "src",
		Select: []string{"*.go"},
	})
	if err != nil {
		t.Fatalf("newFileSet: %v", err)
	}
	// build() consults env.nodeType to resolve each file.
	e.nodeType = func(name string) string {
		switch name {
		case "pkg/a.go", "pkg/b.go":
			return nodeSrc
		}
		return ""
	}

	if err := fs.build(e, &buildOpts{}); err != nil {
		t.Fatalf("build: %v", err)
	}

	// The output JSON list should contain stats for both files.
	outFile := filepath.Join(outDir, "pkg", "src.fileset")
	if _, err := os.Stat(outFile); err != nil {
		t.Fatalf("stat output: %v", err)
	}
	var list []*fileStat
	if err := readJSONFile(outFile, &list); err != nil {
		t.Fatalf("read output: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("got %d stats, want 2: %+v", len(list), list)
	}
	if list[0].Name != "pkg/a.go" || list[1].Name != "pkg/b.go" {
		t.Errorf("output names = [%q, %q], want sorted [pkg/a.go, pkg/b.go]",
			list[0].Name, list[1].Name)
	}
}

func TestFileSet_buildMissingFile(t *testing.T) {
	e, _, _ := newTestEnv(t)
	fs := &fileSet{
		name:  "pkg/src",
		files: []string{"pkg/nope.go"},
		out:   "pkg/src.fileset",
	}
	e.nodeType = func(string) string { return "" }
	err := fs.build(e, &buildOpts{})
	if err == nil {
		t.Fatal("want error for missing file, got nil")
	}
}

// TestFileSet_selectAllFilesInRepo verifies that a file_set with
// Select: ["**"] at the root of a single-repo workspace does NOT pull
// in files that live under _/src (which belong to depended-on repos).
// Skipping is handled in listAllFiles, which treats any directory
// literally named "_" the same way it treats ".git".
func TestFileSet_selectAllFilesInRepo(t *testing.T) {
	const repoName = "test.local/self/dockers"
	e, root := newTestRepoEnv(t, repoName)

	// Self-repo files at the workspace root.
	writeFile(t, filepath.Join(root, "a.go"), "a")
	writeFile(t, filepath.Join(root, "sub/b.go"), "b")

	// A dependency repo, pre-checked-out under _/src.
	writeFile(t,
		filepath.Join(root, "_/src/test.local/dep/dockers/dep.go"), "dep",
	)

	fs, err := newFileSet(e, repoName, &FileSet{
		Name:   "all",
		Select: []string{"**"},
	})
	if err != nil {
		t.Fatalf("newFileSet: %v", err)
	}

	for _, f := range fs.files {
		if strings.HasPrefix(f, "test.local/dep/") {
			t.Errorf("file_set unexpectedly includes dep file %q", f)
		}
	}

	want := map[string]bool{
		repoName + "/a.go":     false,
		repoName + "/sub/b.go": false,
	}
	for _, f := range fs.files {
		if _, ok := want[f]; ok {
			want[f] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("expected self file %q missing from %v", name, fs.files)
		}
	}
}
