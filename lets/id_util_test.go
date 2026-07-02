package lets

import "testing"

func TestShortDigest(t *testing.T) {
	for _, c := range []struct {
		in, want string
	}{
		{"", ""},
		{"abc", "abc"},
		{"abcdefg", "abcdefg"},          // exactly 7
		{"abcdefgh", "abcdefg"},         // longer than 7, truncated
		{"abcdef0123456789", "abcdef0"}, // truncated at 7
		{"ABCDEFGHIJ", "ABCDEFG"},       // uppercase
		{"0123456789", "0123456"},       // digits
		{"aB3xY9Z2", "aB3xY9Z"},         // mixed
		{"abc-def", "abc"},              // stops at '-'
		{"abcdef!ghij", "abcdef"},       // stops at '!' before 7
		{"!abc", ""},                    // leading non-alnum
		{"ab§cd", "ab"},                 // non-ASCII rune stops scan
		{"sha256:abcdef0123", "sha256"}, // stops at ':'
	} {
		if got := shortDigest(c.in); got != c.want {
			t.Errorf("shortDigest(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
