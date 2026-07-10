package gofiledag

import (
	"fmt"
	"go/token"
	"path/filepath"
)

// relPath returns p relative to cwd, or p unchanged when cwd is empty or the
// relative path cannot be computed.
func relPath(p, cwd string) string {
	if cwd == "" {
		return p
	}
	r, err := filepath.Rel(cwd, p)
	if err != nil {
		return p
	}
	return r
}

// relPos formats a token position as "file:line:column" with the filename
// made relative to cwd.
func relPos(p token.Position, cwd string) string {
	return fmt.Sprintf("%s:%d:%d", relPath(p.Filename, cwd), p.Line, p.Column)
}
