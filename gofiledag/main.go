package gofiledag

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
)

// Main runs the gofiledag tool over args. It always reports rule violations
// to stdout. When the -report_output flag is non-empty, a human-readable file
// DAG for each package is written to that file; when -graph_output is
// non-empty, the file DAG of the single package under analysis is written
// there as JSON in the shanhu.io/std/graph.Graph format (it is an error to
// combine -graph_output with more than one package). It returns a process
// exit code: 0 on success, non-zero on a load/output failure or when
// violations are found.
func Main(args []string) int {
	fs := flag.NewFlagSet("gofiledag", flag.ExitOnError)
	tags := fs.String("tags", "", "comma-separated build tags")
	goos := fs.String("goos", runtime.GOOS, "target GOOS")
	goarch := fs.String("goarch", runtime.GOARCH, "target GOARCH")
	reportOutput := fs.String(
		"report_output", "", "if set, write the file graph report to this file",
	)
	graphOutput := fs.String(
		"graph_output", "",
		"if set, write the file graph as JSON (std/graph.Graph) to this file",
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

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "get work dir:", err)
		return 1
	}

	results := AnalyzePasses(passes)

	if *reportOutput != "" {
		if err := writeReportFile(*reportOutput, results, cwd); err != nil {
			fmt.Fprintln(os.Stderr, "report output:", err)
			return 1
		}
	}

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

func writeReportFile(file string, results []*Result, cwd string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	PrintReportResults(f, results, cwd)
	return f.Close()
}

// writeGraphFile writes the single package's file DAG to file as indented
// JSON in the shanhu.io/std/graph.Graph format. It fails if results span
// more than one package.
func writeGraphFile(file string, results []*Result, cwd string) error {
	g, err := buildGraph(results, cwd)
	if err != nil {
		return err
	}
	bs, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	bs = append(bs, '\n')
	return os.WriteFile(file, bs, 0644)
}
