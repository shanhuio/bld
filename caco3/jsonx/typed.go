package jsonx

import (
	"shanhu.io/bld/caco3/lexing"
)

// TypeMaker is a function that makes an interface based on the given type.
type TypeMaker func(t string) any

// Typed is an item in a typed list.
type Typed struct {
	Type string
	V    any
	Pos  *lexing.Pos
}
