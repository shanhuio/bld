package gofiledag

import (
	"reflect"
	"sort"
	"testing"
)

// graphOf builds a FileGraph for tests from an adjacency-list shorthand.
// "files" is the full file list; edges maps from -> [tos...].
func graphOf(files []string, edges map[string][]string) *FileGraph {
	sort.Strings(files)
	g := &FileGraph{
		Files: files,
		Edges: make(map[string]map[string]*Edge),
	}
	for from, tos := range edges {
		m := make(map[string]*Edge)
		for _, to := range tos {
			m[to] = &Edge{Symbol: from + "->" + to}
		}
		g.Edges[from] = m
	}
	return g
}

func TestFindCycles_acyclic(t *testing.T) {
	g := graphOf(
		[]string{"a", "b", "c"},
		map[string][]string{
			"a": {"b"},
			"b": {"c"},
		},
	)
	if got := findCycles(g); len(got) != 0 {
		t.Errorf("findCycles on DAG: got %v, want []", got)
	}
}

func TestFindCycles_simple(t *testing.T) {
	g := graphOf(
		[]string{"a", "b"},
		map[string][]string{
			"a": {"b"},
			"b": {"a"},
		},
	)
	want := [][]string{{"a", "b"}}
	if got := findCycles(g); !reflect.DeepEqual(got, want) {
		t.Errorf("findCycles: got %v, want %v", got, want)
	}
}

func TestFindCycles_multiple(t *testing.T) {
	g := graphOf(
		[]string{"a", "b", "c", "x", "y"},
		map[string][]string{
			"a": {"b"},
			"b": {"c"},
			"c": {"a"},
			"x": {"y"},
			"y": {"x"},
		},
	)
	got := findCycles(g)
	want := [][]string{{"a", "b", "c"}, {"x", "y"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("findCycles: got %v, want %v", got, want)
	}
}

func TestPickSmallestCycle(t *testing.T) {
	sccs := [][]string{{"a", "b", "c"}, {"x", "y"}, {"m", "n"}}
	got := pickSmallestCycle(sccs)
	want := []string{"m", "n"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("pickSmallestCycle: got %v, want %v", got, want)
	}
}

func TestWalkCycle_two(t *testing.T) {
	g := graphOf(
		[]string{"a", "b"},
		map[string][]string{
			"a": {"b"},
			"b": {"a"},
		},
	)
	got := walkCycle(g, []string{"a", "b"})
	want := []string{"a", "b", "a"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("walkCycle: got %v, want %v", got, want)
	}
}

func TestWalkCycle_three(t *testing.T) {
	g := graphOf(
		[]string{"a", "b", "c"},
		map[string][]string{
			"a": {"b"},
			"b": {"c"},
			"c": {"a"},
		},
	)
	got := walkCycle(g, []string{"a", "b", "c"})
	want := []string{"a", "b", "c", "a"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("walkCycle: got %v, want %v", got, want)
	}
}

func TestFindFirstCycle_none(t *testing.T) {
	g := graphOf([]string{"a"}, nil)
	if got := findFirstCycle(g); got != nil {
		t.Errorf("findFirstCycle: got %v, want nil", got)
	}
}

func TestFindFirstCycle_picksSmallest(t *testing.T) {
	g := graphOf(
		[]string{"a", "b", "c", "x", "y"},
		map[string][]string{
			"a": {"b"},
			"b": {"c"},
			"c": {"a"},
			"x": {"y"},
			"y": {"x"},
		},
	)
	got := findFirstCycle(g)
	want := []string{"x", "y", "x"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("findFirstCycle: got %v, want %v", got, want)
	}
}
