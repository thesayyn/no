package cache

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"

	"github.com/thesayyn/no/pkg/build"
)

type DiskCache struct {
	output_base string
}

// Setup implements Cache
func (dc *DiskCache) Setup() error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("error getting cache dir: %s", err)
	}
	outputBase := filepath.Join(cacheDir, "_no", "output_base")
	err = os.MkdirAll(outputBase, 0777)
	if err != nil {
		return fmt.Errorf("error creating cache dir: %s", err)
	}
	dc.output_base = outputBase
	return nil
}

// Store implements Cache
func (dc *DiskCache) Store(task build.Task) error {
	h, err := dc.calculateHash(task)
	if err != nil {
		return err
	}

	dest := filepath.Join(dc.output_base, fmt.Sprintf("%x", h.Sum(nil)))

	output, err := task.Output()
	if err != nil {
		return err
	}

	info, err := os.Stat(output.Path())
	if err != nil {
		return err
	}

	if info.Mode().IsDir() {

		return filepath.Walk(output.Path(), func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(output.Path(), path)
			if err != nil {
				return err
			}
			if rel == "." {
				return nil
			}
			full := filepath.Join(dest, rel)
			if info.Mode().IsDir() {
				if err := os.MkdirAll(full, 0777); err != nil {
					return err
				}
				return nil
			} else if info.Mode().IsRegular() {
				return os.Link(path, full)
			}
			return fmt.Errorf("not in a good state: %s", path)
		})
	} else if info.Mode().IsRegular() {
		return os.Link(output.Path(), dest)
	}

	return nil
}

// Fetch implements Cache
func (dc *DiskCache) Fetch(task build.Task) (build.Output, error) {
	h, err := dc.calculateHash(task)
	if err != nil {
		return nil, err
	}

	dest := filepath.Join(dc.output_base, fmt.Sprintf("%x", h.Sum(nil)))

	info, err := os.Stat(dest)

	if err != nil {
		return nil, err
	}

	if info.Mode().IsDir() {
		return build.NewDiskTreeOutput(dest, nil), nil
	} else if info.Mode().IsRegular() {
		return build.NewDiskOutput(dest)
	}

	return nil, fmt.Errorf("not in a good state")
}

// Hit implements Cache
func (dc *DiskCache) Hit(task build.Task) (bool, error) {
	h, err := dc.calculateHash(task)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(filepath.Join(dc.output_base, fmt.Sprintf("%x", h.Sum(nil))))

	if err != nil {
		return false, nil
	}

	return true, nil
}

func (dc *DiskCache) calculateHash(task build.Task) (hash.Hash, error) {
	hash := sha256.New()
	hash.Write([]byte(reflect.TypeOf(task).Elem().Name()))

	inputs, err := task.Inputs()
	if err != nil {
		return nil, err
	}
	for _, input := range inputs {
		h, err := input.Hash()
		if err != nil {
			return nil, err
		}
		hash.Write(h.Sum(nil))
	}

	return hash, nil
}

func NewDiskCache() Cache {
	return &DiskCache{}
}
