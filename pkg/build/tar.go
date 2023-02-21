package build

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func BuildTar(tw *tar.Writer, root string, prefix string, mtime time.Time, glob Glob) error {
	prev := ""
	for _, dir := range strings.Split(prefix, "/") {
		dir = filepath.Join(prev, dir)
		if err := tw.WriteHeader(&tar.Header{
			Name:     dir,
			Typeflag: tar.TypeDir,
			Mode:     0555,
			ModTime:  mtime,
		}); err != nil {
			return fmt.Errorf("writing dir %q: %w", dir, err)
		}
	}
	return filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {

		if !glob.Matches(path) {
			return nil
		}

		newPath := filepath.Join(prefix, path)

		if info.IsDir() {
			if err := tw.WriteHeader(&tar.Header{
				Name:     newPath,
				Typeflag: tar.TypeDir,
				Mode:     0555,
				ModTime:  mtime,
			}); err != nil {
				return fmt.Errorf("writing dir %q: %w", newPath, err)
			}
		} else {

			if err := tw.WriteHeader(&tar.Header{
				Name:     newPath,
				Size:     info.Size(),
				Typeflag: tar.TypeReg,
				Mode:     0555,
			}); err != nil {
				return fmt.Errorf("writing file %q: %w", path, err)
			}

			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("opening file(%q): %w", path, err)
			}
			defer file.Close()
			if _, err := io.Copy(tw, file); err != nil {
				return fmt.Errorf("copying file (%q): %w", path, err)
			}
		}

		return nil
	})
}
