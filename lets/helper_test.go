package lets

import "strings"

// multiLine joins the given lines with newlines and appends a trailing
// newline. Used in tests so multi-line string literals don't have to
// break the surrounding code's indentation.
func multiLine(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}
