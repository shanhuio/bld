package gofiledag

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

// Main runs the gofiledag tool over args. It always reports rule violations
// to stdout. When the -graph_output flag is non-empty, the file DAG for each
// package is also written to that file. It returns a process exit code: 0 on
// success, non-zero on a load/output failure or when violations are found.
func Main(args []string) int {
	fs := flag.NewFlagSet("gofiledag", flag.ExitOnError)
	tags := fs.String("tags", "", "comma-separated build tags")
	goos := fs.String("goos", runtime.GOOS, "target GOOS")
	goarch := fs.String("goarch", runtime.GOARCH, "target GOARCH")
	graphOutput := fs.String(
		"graph_output", "", "if set, write the file graph to this file",
	)
	fs.Parse(args)

	patterns := fs.Args()
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	cfg := &LoadConfig{
		GOOS:   *goos,
		GOARCH: *goarch,
	}
	if *tags != "" {
		cfg.Tags = []string{*tags}
	}

	passes, err := LoadPasses(cfg, patterns)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load:", err)
		return 1
	}

	cwd, _ := os.Getwd()

	results := AnalyzePasses(passes)

	if *graphOutput != "" {
		if err := writeGraphFile(*graphOutput, results, cwd); err != nil {
			fmt.Fprintln(os.Stderr, "graph output:", err)
			return 1
		}
	}

	fails := PrintCheckResults(os.Stdout, results, cwd)
	if fails > 0 {
		return 1
	}
	return 0
}

func writeGraphFile(file string, results []*Result, cwd string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	PrintGraphResults(f, results, cwd)
	return f.Close()
}
