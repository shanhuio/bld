package main

import (
	"fmt"
	"os"

	"shanhu.io/bld/gofiledag"
)

const usage = `usage: gofiledag [flags] [packages...]

Reports file-DAG rule violations to stdout, exiting non-zero on failure.

flags:
  -tags=...           comma-separated build tags
  -goos=...           target GOOS (default: current go env)
  -goarch=...         target GOARCH (default: current go env)
  -report_output=...  if set, write the file DAG report for each package here
  -graph_output=...   if set, write the single package's file DAG as JSON here

packages default to "./...".
`

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "-h", "--help", "help":
			fmt.Print(usage)
			os.Exit(0)
		}
	}
	os.Exit(gofiledag.Main(args))
}
