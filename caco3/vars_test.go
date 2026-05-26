package caco3

import (
	"reflect"
	"testing"
)

func TestMakeDockerVars(t *testing.T) {
	lookup := func(env map[string]string) func(string) (string, bool) {
		return func(k string) (string, bool) {
			v, ok := env[k]
			return v, ok
		}
	}

	for _, c := range []struct {
		name    string
		envs    []string
		environ map[string]string
		want    map[string]string
	}{
		{
			name: "literal assignments",
			envs: []string{"A=1", "B=two"},
			want: map[string]string{"A": "1", "B": "two"},
		},
		{
			name:    "lookup hit",
			envs:    []string{"HOME"},
			environ: map[string]string{"HOME": "/root"},
			want:    map[string]string{"HOME": "/root"},
		},
		{
			name:    "lookup miss is dropped",
			envs:    []string{"MISSING"},
			environ: map[string]string{},
			want:    map[string]string{},
		},
		{
			name:    "literal overrides lookup",
			envs:    []string{"X=fromArg"},
			environ: map[string]string{"X": "fromEnv"},
			want:    map[string]string{"X": "fromArg"},
		},
		{
			name:    "mixed",
			envs:    []string{"A=1", "B", "C"},
			environ: map[string]string{"B": "bb"},
			want:    map[string]string{"A": "1", "B": "bb"},
		},
		{
			name: "empty value after =",
			envs: []string{"A="},
			want: map[string]string{"A": ""},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := makeDockerVars(c.envs, lookup(c.environ))
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestMakeDockerVars_nilLookupUsesOSEnv(t *testing.T) {
	t.Setenv("CACO3_TEST_VAR", "hello")
	got := makeDockerVars([]string{"CACO3_TEST_VAR"}, nil)
	want := map[string]string{"CACO3_TEST_VAR": "hello"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
