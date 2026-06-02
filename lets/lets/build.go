package main

import (
	"fmt"
	"os"
	"path/filepath"

	"shanhu.io/bld/lets"
	"shanhu.io/std/lexing"
)

func cmdBuild(args []string) error {
	flags := newFlags()
	config := new(lets.Config)
	declareBuildFlags(flags, config)
	args = parseArgs(flags, args)

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get work dir: %w", err)
	}
	if config.Root != "" {
		root, err := filepath.Abs(config.Root)
		if err != nil {
			return fmt.Errorf("get abs root dir: %w", err)
		}
		config.Root = root
	}

	b, err := lets.NewBuilder(wd, config)
	if err != nil {
		return fmt.Errorf("new builder: %w", err)
	}

	if _, errs := b.ReadWorkspace(); errs != nil {
		lexing.FprintErrs(os.Stderr, errs, wd)
		return fmt.Errorf("read workspace got %d errors", len(errs))
	}

	if errs := b.Build(args); errs != nil {
		lexing.FprintErrs(os.Stderr, errs, wd)
		return fmt.Errorf("build got %d errors", len(errs))
	}

	return nil
}
