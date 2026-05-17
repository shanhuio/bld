package main

import (
	"shanhu.io/bld/caco3/subcmd"
)

func cmd() *subcmd.List {
	c := subcmd.New()
	c.Add("build", "build rules", cmdBuild)
	c.Add("sync", "sync source repos", cmdSync)
	return c
}

func main() { cmd().Main() }
