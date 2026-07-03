// Command coralint bundles a collection of lints into a single binary,
// where each lint is exposed as a subcommand.
package main

import (
	"os"

	"shanhu.io/bld/gofiledag"
	"shanhu.io/bld/subcmd"
)

// register adds every built-in lint to the command list.
func register(cmds *subcmd.List) {
	cmds.Add("gofiledag", "checks file-level DAG rules", gofiledagMain)
}

// gofiledagMain adapts gofiledag.Main, which returns a process exit code,
// to a subcmd entry. It exits with that code on failure and returns nil on
// success so dispatch continues normally.
func gofiledagMain(args []string) error {
	if code := gofiledag.Main(args); code != 0 {
		os.Exit(code)
	}
	return nil
}

func main() {
	cmds := subcmd.New()
	register(cmds)
	cmds.Main()
}
