package targz

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Compress(destination string, assets ...string) error {
	dest, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer dest.Close()

	gz := gzip.NewWriter(dest)
	defer gz.Close()

	dst := tar.NewWriter(gz)
	defer dst.Close()

	if len(assets) == 0 {
		return nil
	}

	skip := true
	info, err := os.Stat(assets[0])
	if err != nil {
		return err
	}
	if !info.IsDir() {
		skip = false
	}

	if len(assets) > 1 {
		skip = false
	}

	root := ""
	walk := func(path string, info os.FileInfo, err error) error {
		if path == destination {
			return nil
		}
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}

		if skip {
			rel, err := filepath.Rel(assets[0], path)
			if err != nil {
				return err
			}
			if rel == "." {
				return nil // skip
			}
			hdr.Name = filepath.ToSlash(rel)
		} else {
			hdr.Name = filepath.ToSlash(filepath.Join(filepath.Base(root), strings.TrimPrefix(path, root)))
		}

		if err := dst.WriteHeader(hdr); err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		if _, err = io.Copy(dst, file); err != nil {
			return err
		}
		return nil
	}

	for _, a := range assets {
		info, err := os.Stat(a)
		if err != nil {
			return err
		}
		root = a
		if info.IsDir() {
			if err := filepath.Walk(a, walk); err != nil {
				return err
			}
		} else if err := walk(a, info, err); err != nil {
			return err
		}
	}
	return nil
}
