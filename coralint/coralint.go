// Package coralint bundles a collection of lints into a single binary,
// where each lint is exposed as a subcommand.
package coralint

import (
	"shanhu.io/bld/subcmd"
)

// register adds every built-in lint to the command list. Lints are wired
// in here as they are implemented; gofiledag will be added as a subcommand.
func register(_ *subcmd.List) {
	// cmds.Add("gofiledag", "...", gofiledagMain)
}

// Main runs the coralint command-line tool, dispatching to the lint named
// by the first argument. It calls os.Exit on a non-zero result.
func Main() {
	cmds := subcmd.New()
	register(cmds)
	cmds.Main()
}
