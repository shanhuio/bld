package filedag

import "sort"

// findCycles returns all SCCs of size >= 2 (real cycles) via Tarjan's
// algorithm. Each SCC is returned with its files in sorted order, and the
// outer slice is sorted by the alphabetically smallest member.
func findCycles(g *FileGraph) [][]string {
	idx := 0
	indices := make(map[string]int)
	lowlinks := make(map[string]int)
	onStack := make(map[string]bool)
	var stack []string
	var sccs [][]string

	var sc func(v string)
	sc = func(v string) {
		indices[v] = idx
		lowlinks[v] = idx
		idx++
		stack = append(stack, v)
		onStack[v] = true

		for _, w := range g.successors(v) {
			if _, seen := indices[w]; !seen {
				sc(w)
				if lowlinks[w] < lowlinks[v] {
					lowlinks[v] = lowlinks[w]
				}
			} else if onStack[w] {
				if indices[w] < lowlinks[v] {
					lowlinks[v] = indices[w]
				}
			}
		}

		if lowlinks[v] == indices[v] {
			var scc []string
			for {
				n := len(stack) - 1
				w := stack[n]
				stack = stack[:n]
				onStack[w] = false
				scc = append(scc, w)
				if w == v {
					break
				}
			}
			if len(scc) >= 2 {
				sort.Strings(scc)
				sccs = append(sccs, scc)
			}
		}
	}

	for _, f := range g.Files {
		if _, seen := indices[f]; !seen {
			sc(f)
		}
	}
	sort.Slice(sccs, func(i, j int) bool { return sccs[i][0] < sccs[j][0] })
	return sccs
}

// pickSmallestCycle returns the smallest SCC by node count; ties broken by
// the alphabetically smallest first member.
func pickSmallestCycle(sccs [][]string) []string {
	if len(sccs) == 0 {
		return nil
	}
	best := sccs[0]
	for _, s := range sccs[1:] {
		if len(s) < len(best) || (len(s) == len(best) && s[0] < best[0]) {
			best = s
		}
	}
	return best
}

// walkCycle returns a closed cycle through the SCC as an ordered chain of
// file names: chain[0] == chain[len-1]. Uses BFS from the alphabetically
// smallest node to find the shortest cycle for clarity.
func walkCycle(g *FileGraph, scc []string) []string {
	if len(scc) == 0 {
		return nil
	}
	member := make(map[string]bool)
	for _, f := range scc {
		member[f] = true
	}
	start := scc[0]

	type node struct {
		name string
		prev *node
	}
	visited := make(map[string]bool)
	var queue []*node
	for _, w := range g.successors(start) {
		if !member[w] {
			continue
		}
		queue = append(queue, &node{name: w, prev: &node{name: start}})
		visited[w] = true
	}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		if n.name == start {
			var chain []string
			for c := n; c != nil; c = c.prev {
				chain = append([]string{c.name}, chain...)
			}
			return chain
		}
		for _, w := range g.successors(n.name) {
			if !member[w] {
				continue
			}
			if visited[w] && w != start {
				continue
			}
			queue = append(queue, &node{name: w, prev: n})
			if w != start {
				visited[w] = true
			}
		}
	}
	// Should not happen for a real SCC of size >= 2.
	closed := append([]string{}, scc...)
	closed = append(closed, scc[0])
	return closed
}

// findFirstCycle returns one closed cycle (smallest by node count) for the
// graph, or nil if the graph is acyclic.
func findFirstCycle(g *FileGraph) []string {
	sccs := findCycles(g)
	if len(sccs) == 0 {
		return nil
	}
	return walkCycle(g, pickSmallestCycle(sccs))
}
