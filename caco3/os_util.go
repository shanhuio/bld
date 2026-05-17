package caco3

import (
	"os"
)

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
