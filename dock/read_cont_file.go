package dock

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
)

// ReadContFile reads out a single file from a container.
func ReadContFile(c *Cont, f string) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := c.CopyOutTar(f, buf); err != nil {
		return nil, fmt.Errorf("read from container: %w", err)
	}

	var content []byte
	got := false
	r := tar.NewReader(buf)
	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar: %w", err)
		}

		if got {
			return nil, errors.New("not one single file")
		}
		if h.Typeflag != tar.TypeReg {
			return nil, errors.New("not a regular file")
		}

		bs, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("read file content: %w", err)
		}
		content = bs
		got = true
	}

	if !got {
		return nil, ErrFileNotFound
	}

	return content, nil
}
