// Package gopkgdag builds the package-level dependency graph of a Go
// module: one node per loaded package, one edge per import between two
// loaded packages. Imports outside the loaded set (standard library and
// external modules) are omitted, so the result is the module's internal
// package DAG.
//
// It is exposed as the coralint "gopkgdag" subcommand and emits the graph
// as shanhu.io/std/graph.Graph JSON.
package gopkgdag
