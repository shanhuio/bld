package caco3

// shortDigest returns the first 7 bytes of an ID.
// The prefix must be letters or digits.
func shortDigest(s string) string {
	n := len(s)
	for i, r := range s {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		n = i
		break
	}
	if n > 7 {
		n = 7
	}
	return s[:n]
}
