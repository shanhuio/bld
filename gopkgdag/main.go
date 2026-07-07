package gopkgdag

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
)

// Main runs the gopkgdag tool over args, writing the module's package
// dependency graph as shanhu.io/std/graph.Graph JSON to the -output file
// (or stdout when -output is empty). It returns a process exit code: 0 on
// success, non-zero on a load or output failure.
func Main(args []string) int {
	fs := flag.NewFlagSet("gopkgdag", flag.ExitOnError)
	tags := fs.String("tags", "", "comma-separated build tags")
	goos := fs.String("goos", runtime.GOOS, "target GOOS")
	goarch := fs.String("goarch", runtime.GOARCH, "target GOARCH")
	output := fs.String("output", "", "write JSON graph here (default: stdout)")
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

	pkgs, err := Load(cfg, patterns)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load:", err)
		return 1
	}

	g, err := BuildGraph(pkgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "build graph:", err)
		return 1
	}

	bs, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "marshal:", err)
		return 1
	}
	bs = append(bs, '\n')

	if *output == "" {
		if _, err := os.Stdout.Write(bs); err != nil {
			fmt.Fprintln(os.Stderr, "write:", err)
			return 1
		}
		return 0
	}
	if err := os.WriteFile(*output, bs, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 1
	}
	return 0
}
