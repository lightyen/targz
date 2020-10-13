package targz

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

func Extract(src io.Reader, destination string) error {
	gr, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	defer gr.Close()
	t := tar.NewReader(gr)

	for {
		hdr, err := t.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		to := filepath.Join(destination, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(to, hdr.FileInfo().Mode()); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeChar, tar.TypeBlock, tar.TypeFifo, tar.TypeGNUSparse:
			if err = writeNewFile(to, t, hdr.FileInfo().Mode()); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err = writeNewSymbolicLink(to, hdr.Linkname); err != nil {
				return err
			}
		case tar.TypeLink:
			if err = writeNewHardLink(to, filepath.Join(to, hdr.Linkname)); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeNewFile(filename string, in io.Reader, mode os.FileMode) error {
	dst, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer dst.Close()
	if err = dst.Chmod(mode); err != nil {
		return err
	}
	if _, err = io.Copy(dst, in); err != nil {
		return err
	}
	return nil
}

func writeNewSymbolicLink(name string, target string) error {
	if err := os.Remove(name); err != nil && err != os.ErrNotExist {
		return err
	}
	if err := os.Symlink(target, name); err != nil {
		return err
	}
	return nil
}

func writeNewHardLink(name string, target string) error {
	if err := os.Remove(name); err != nil && err != os.ErrNotExist {
		return err
	}
	if err := os.Link(target, name); err != nil {
		return err
	}
	return nil
}
