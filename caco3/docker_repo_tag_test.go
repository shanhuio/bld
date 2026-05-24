package caco3

import "testing"

func TestParseRepoTag(t *testing.T) {
	for _, c := range []struct {
		in          string
		wantRepo    string
		wantTag     string
	}{
		{"alpine", "alpine", "latest"},
		{"alpine:3.23", "alpine", "3.23"},
		{"test.local/proj/foo", "test.local/proj/foo", "latest"},
		{"test.local/proj/foo:v1.0", "test.local/proj/foo", "v1.0"},
		{"", "", "latest"},
	} {
		t.Run(c.in, func(t *testing.T) {
			gotRepo, gotTag := parseRepoTag(c.in)
			if gotRepo != c.wantRepo || gotTag != c.wantTag {
				t.Errorf("parseRepoTag(%q) = (%q, %q), want (%q, %q)",
					c.in, gotRepo, gotTag, c.wantRepo, c.wantTag)
			}
		})
	}
}

func TestRepoTag(t *testing.T) {
	for _, c := range []struct {
		repo, tag, want string
	}{
		{"alpine", "latest", "alpine:latest"},
		{"alpine", "", "alpine"},
		{"test.local/proj/foo", "v1", "test.local/proj/foo:v1"},
		{"", "", ""},
		{"", "x", ":x"},
	} {
		got := repoTag(c.repo, c.tag)
		if got != c.want {
			t.Errorf("repoTag(%q, %q) = %q, want %q",
				c.repo, c.tag, got, c.want)
		}
	}
}

func TestNameToRepoTag(t *testing.T) {
	for _, c := range []struct {
		name    string
		want    string
		wantErr bool
	}{
		// Basic 4-part, third component "dockers".
		{
			name: "test.local/proj/dockers/base",
			want: "test.local/proj/base",
		},
		// Third component may end in "-dockers".
		{
			name: "test.local/proj/web-dockers/app",
			want: "test.local/proj/app",
		},
		// shanhu.io domain is remapped to cr.shanhu.io.
		{
			name: "shanhu.io/x/dockers/base",
			want: "cr.shanhu.io/x/base",
		},
		{
			name: "shanhu.io/lib/web-dockers/app",
			want: "cr.shanhu.io/lib/app",
		},

		// Errors.
		{name: "", wantErr: true},
		{name: "a", wantErr: true},
		{name: "a/b/c", wantErr: true},
		{name: "a/b/c/d/e", wantErr: true},
		{name: "a/b/other/d", wantErr: true},   // third component not dockers
		{name: "a/b/dockers2/d", wantErr: true}, // third component looks similar but doesn't match
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := nameToRepoTag(c.name)
			if c.wantErr {
				if err == nil {
					t.Errorf("nameToRepoTag(%q) = %q, want error", c.name, got)
				}
				return
			}
			if err != nil {
				t.Errorf("nameToRepoTag(%q) returned %v", c.name, err)
				return
			}
			if got != c.want {
				t.Errorf("nameToRepoTag(%q) = %q, want %q", c.name, got, c.want)
			}
		})
	}
}
