package caco3

import (
	"shanhu.io/std/lexing"
)

const (
	nodeSrc  = "src"
	nodeRule = "rule"
	nodeOut  = "out"
	nodeSub  = "sub"
)

type buildNode struct {
	name string
	typ  string
	deps []string
	pos  *lexing.Pos

	ruleType string
	rule     buildRule
	ruleMeta *buildRuleMeta

	sub *subBuilds
}
