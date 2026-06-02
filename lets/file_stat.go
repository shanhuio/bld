package lets

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type fileStat struct {
	Name         string
	Type         string
	Size         int64
	ModTimestamp int64
	Mode         uint32
	Symlink      string `json:",omitempty"`
}

const (
	fileTypeSrc = "s"
	fileTypeOut = "o"
)

func newOutFileStat(env *env, p string) (*fileStat, error) {
	return newFileStat(env, p, fileTypeOut)
}

func newSrcFileStat(env *env, p string) (*fileStat, error) {
	return newFileStat(env, p, fileTypeSrc)
}

var errNotFound = errors.New("not found")

func newFileStat(env *env, p, t string) (*fileStat, error) {
	var f string
	if t == fileTypeOut {
		f = env.out(p)
	} else {
		f = env.src(p)
	}

	info, err := os.Lstat(f) // Does not follow symlink.
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s:%s %w", t, p, errNotFound)
		}
		return nil, err
	}

	var symLink string
	mod := info.Mode()
	if mod&fs.ModeSymlink != 0 { // A sym link.
		dest, err := os.Readlink(f)
		if err != nil {
			return nil, fmt.Errorf("read sym link: %w", err)
		}
		symLink = dest
	}

	return &fileStat{
		Name:         p,
		Type:         t,
		Size:         info.Size(),
		ModTimestamp: info.ModTime().UnixNano(),
		Mode:         uint32(info.Mode()),
		Symlink:      symLink,
	}, nil
}

func sameFileStat(env *env, stat *fileStat) (bool, error) {
	cur, err := newFileStat(env, stat.Name, stat.Type)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("check current: %w", err)
	}

	same := cur.Size == stat.Size
	same = same && cur.ModTimestamp == stat.ModTimestamp
	same = same && cur.Mode == stat.Mode
	same = same && cur.Symlink == stat.Symlink

	return same, nil
}
