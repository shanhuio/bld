package tarutil

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"path"
)

func copyZipFile(w io.Writer, f *zip.File) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, rc); err != nil {
		rc.Close()
		return err
	}
	return rc.Close()
}

// TarZipFile puts all files from a zip file into a tar stream.
func TarZipFile(tw *tar.Writer, p string, dir string) error {
	z, err := zip.OpenReader(p)
	if err != nil {
		return fmt.Errorf("open zip file: %w", err)
	}

	for _, f := range z.File {
		stat := f.FileInfo()
		tarStat, err := tar.FileInfoHeader(stat, "")
		if err != nil {
			return fmt.Errorf("tar stat for: %q: %w", f.Name, err)
		}
		name := f.Name
		if dir != "" {
			name = path.Join(dir, name)
		}
		tarStat.Name = name
		if err := tw.WriteHeader(tarStat); err != nil {
			return fmt.Errorf("write header: %q: %w", f.Name, err)
		}
		if err := copyZipFile(tw, f); err != nil {
			return fmt.Errorf("copy zip file: %q: %w", f.Name, err)
		}
	}

	return nil
}
