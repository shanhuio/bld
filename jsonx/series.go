package jsonx

import (
	"shanhu.io/bld/lexing"
)

type typeName struct {
	tok  *lexing.Token
	name string
}

type typedEntry struct {
	typ   *typeName
	value value
	semi  *lexing.Token
}

type series struct {
	entries []*typedEntry
}
