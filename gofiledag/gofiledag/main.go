package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"shanhu.io/bld/gofiledag"
)

const usage = `usage: gofiledag <command> [flags] [packages...]

commands:
  check     report rule violations and exit non-zero on failure
  graph     print the file DAG for each package (or violations on failure)

flags:
  -tags=...     comma-separated build tags
  -goos=...     target GOOS (default: current go env)
  -goarch=...   target GOARCH (default: current go env)

packages default to "./...".
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	switch cmd {
	case "check":
		os.Exit(run(args, false))
	case "graph":
		os.Exit(run(args, true))
	case "-h", "--help", "help":
		fmt.Print(usage)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", cmd, usage)
		os.Exit(2)
	}
}

func run(args []string, graphMode bool) int {
	fs := flag.NewFlagSet("gofiledag", flag.ExitOnError)
	tags := fs.String("tags", "", "comma-separated build tags")
	goos := fs.String("goos", runtime.GOOS, "target GOOS")
	goarch := fs.String("goarch", runtime.GOARCH, "target GOARCH")
	fs.Parse(args)

	patterns := fs.Args()
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	cfg := &gofiledag.LoadConfig{
		GOOS:   *goos,
		GOARCH: *goarch,
	}
	if *tags != "" {
		cfg.Tags = []string{*tags}
	}

	passes, err := gofiledag.LoadPasses(cfg, patterns)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load:", err)
		return 1
	}

	cwd, _ := os.Getwd()

	results := gofiledag.AnalyzePasses(passes)

	var fails int
	if graphMode {
		fails = gofiledag.PrintGraphResults(os.Stdout, results, cwd)
	} else {
		fails = gofiledag.PrintCheckResults(os.Stdout, results, cwd)
	}
	if fails > 0 {
		return 1
	}
	return 0
}
