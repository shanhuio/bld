package caco3

import (
	"reflect"
	"testing"
)

func TestMakeStrSet(t *testing.T) {
	for _, c := range []struct {
		name string
		in   []string
		want map[string]bool
	}{
		{"empty", nil, map[string]bool{}},
		{"unique", []string{"a", "b", "c"}, map[string]bool{"a": true, "b": true, "c": true}},
		{"duplicates", []string{"a", "a", "b"}, map[string]bool{"a": true, "b": true}},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := makeStrSet(c.in)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("makeStrSet(%v) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

func TestSortedStrList(t *testing.T) {
	for _, c := range []struct {
		name string
		in   map[string]bool
		want []string
	}{
		{"empty", map[string]bool{}, nil},
		{"single", map[string]bool{"a": true}, []string{"a"}},
		{"sorted", map[string]bool{"c": true, "a": true, "b": true}, []string{"a", "b", "c"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := sortedStrList(c.in)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("sortedStrList(%v) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

func TestStrSet_roundTrip(t *testing.T) {
	in := []string{"banana", "apple", "cherry", "apple"}
	got := sortedStrList(makeStrSet(in))
	want := []string{"apple", "banana", "cherry"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("roundtrip: got %v, want %v", got, want)
	}
}
