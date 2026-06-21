package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"shanhu.io/bld/lets"
)

func newFlags() *flag.FlagSet {
	return flag.NewFlagSet("lets", flag.ExitOnError)
}

func parseArgs(set *flag.FlagSet, args []string) []string {
	set.Parse(args)
	return set.Args()
}

func declareBuildFlags(flags *flag.FlagSet, c *lets.Config) {
	flags.StringVar(&c.Root, "root", "", "root directory")
	flags.BoolVar(&c.AlwaysRebuild, "rebuild", false, "always rebuild")
	flags.BoolVar(
		&c.UseDockerBuildCache, "docker_build_cache", true,
		"use docker build cache or not",
	)
}

// resolveWorkDir picks the directory to run the builder from. When a root is
// given explicitly via -root, that root (made absolute) is used as the work
// dir, so targets resolve relative to the root regardless of where the
// command was launched. Otherwise the launcher's current directory is used.
func resolveWorkDir(c *lets.Config) (string, error) {
	if c.Root != "" {
		root, err := filepath.Abs(c.Root)
		if err != nil {
			return "", fmt.Errorf("get abs root dir: %w", err)
		}
		c.Root = root
		return root, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get work dir: %w", err)
	}
	return wd, nil
}
