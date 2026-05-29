package main

import (
	"fmt"
	"os"

	"shanhu.io/bld/caco3"
	"shanhu.io/std/lexing"
)

const sumsFile = "sums.jsonx"

func cmdSync(args []string) error {
	flags := newFlags()
	config := new(caco3.Config)
	declareBuildFlags(flags, config)
	pull := flags.Bool("pull", false, "pull latest commit")
	save := flags.Bool("save", false, "save latest commit into sums file")
	setRemotes := flags.Bool("set_remotes", false, "sets remote URLs")
	args = parseArgs(flags, args)

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get work dir: %w", err)
	}

	b, err := caco3.NewBuilder(wd, config)
	if err != nil {
		return fmt.Errorf("new builder: %w", err)
	}

	if _, errs := b.ReadWorkspace(); errs != nil {
		lexing.FprintErrs(os.Stderr, errs, wd)
		return fmt.Errorf("read workspace got %d errors", len(errs))
	}
	var sums *caco3.RepoSums
	if !*pull {
		s, err := caco3.ReadRepoSums(sumsFile)
		if err != nil {
			return fmt.Errorf("read build sums: %w", err)
		}
		sums = s
	}

	opts := &caco3.SyncOptions{
		SetRemotes: *setRemotes,
	}

	newSums, err := b.SyncRepos(sums, opts)
	if err != nil {
		return err
	}
	if *save {
		if err := caco3.SaveRepoSums(sumsFile, newSums); err != nil {
			return fmt.Errorf("save build sums: %w", err)
		}
	}
	return nil
}
