package lets

import (
	"os"
	"path/filepath"
	"strings"
)

// pathUnder reports whether physical lives inside (or is equal to) base
// and, if so, returns the segment beneath base with no leading separator.
// Both arguments must be clean absolute filesystem paths.
func pathUnder(base, physical string) (string, bool) {
	if base == "" {
		return "", false
	}
	if physical == base {
		return "", true
	}
	prefix := base + string(filepath.Separator)
	if !strings.HasPrefix(physical, prefix) {
		return "", false
	}
	return strings.TrimPrefix(physical, prefix), true
}

func isRegularFile(filePath string) (bool, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return stat.Mode().IsRegular(), nil
}

func isDir(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return stat.IsDir(), nil
}
