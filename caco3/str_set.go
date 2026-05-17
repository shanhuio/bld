package caco3

import (
	"sort"
)

func makeStrSet(list []string) map[string]bool {
	m := make(map[string]bool)
	for _, s := range list {
		m[s] = true
	}
	return m
}

func sortedStrList(m map[string]bool) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
