package main

import (
	"flag"

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
