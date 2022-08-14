package build

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/thesayyn/no/pkg/node"
	"github.com/thesayyn/no/pkg/options"
)

const AppRoot = "/var/run/no"

type Builder struct {
	BuildOpts options.BuildOptions
}

func (b *Builder) Build(root string) (v1.Image, error) {

	ref, err := name.ParseReference(b.BuildOpts.BaseImage)
	if err != nil {
		return nil, fmt.Errorf("parsing reference %q: %w", b.BuildOpts.BaseImage, err)
	}

	rmt, err := remote.Get(ref)
	if err != nil {
		return nil, err
	}

	var img v1.Image

	if rmt.MediaType.IsIndex() {
		idx, err := rmt.ImageIndex()
		if err != nil {
			return nil, err
		}

		m, err := idx.IndexManifest()
		if err != nil {
			return nil, err
		}
		for _, manifest := range m.Manifests {
			if manifest.Platform.OS == "linux" && manifest.Platform.Architecture == "arm64" {
				iimg, err := idx.Image(manifest.Digest)

				if err != nil {
					return nil, err
				}

				img = iimg

				break
			}
		}
	} else {
		iimg, err := rmt.Image()

		if err != nil {
			return nil, err
		}

		img = iimg
	}

	buffer, err := tarLayer(root, b.BuildOpts.CreationTime)

	if err != nil {
		return nil, err
	}

	layer, err := tarball.LayerFromOpener(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBuffer(buffer.Bytes())), nil
	}, tarball.WithCompressedCaching)

	if err != nil {
		return nil, err
	}

	img, err = mutate.Append(img, mutate.Addendum{
		Layer: layer,
		History: v1.History{
			Author:    "no",
			Created:   b.BuildOpts.CreationTime,
			CreatedBy: "no build " + ref.String(),
		},
		MediaType: types.DockerLayer,
	})
	if err != nil {
		return nil, err
	}

	pkg, err := node.GetPackageJson(root)

	if err != nil {
		return nil, err
	}

	cfg, err := img.ConfigFile()

	if err != nil {
		return nil, err
	}

	cfg = cfg.DeepCopy()

	workDir := filepath.Join(AppRoot, root)
	cfg.Config.WorkingDir = workDir

	cfg.Config.Cmd = []string{pkg.Bin}
	cfg.Author = "github.com/thesayyn/no"

	img, err = mutate.ConfigFile(img, cfg)

	if err != nil {
		return nil, err
	}

	return img, nil
}

func tarLayer(root string, creationTime v1.Time) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	dirs := []string{
		"/var",
		"/var/run",
		AppRoot,
	}

	for _, dir := range dirs {
		if err := tw.WriteHeader(&tar.Header{
			Name:     dir,
			Typeflag: tar.TypeDir,
			// Use a fixed Mode, so that this isn't sensitive to the directory and umask
			// under which it was created. Additionally, windows can only set 0222,
			// 0444, or 0666, none of which are executable.
			Mode:    0555,
			ModTime: creationTime.Time,
		}); err != nil {
			return nil, fmt.Errorf("writing dir %q: %w", dir, err)
		}
	}

	if err := CopyContents(tw, root, AppRoot, creationTime); err != nil {
		return nil, err
	}

	return buf, nil
}

func CopyContents(tw *tar.Writer, root string, chroot string, creationTime v1.Time) error {
	return filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {

		newPath := filepath.Join(chroot, path)

		if info.IsDir() {
			if err := tw.WriteHeader(&tar.Header{
				Name:     newPath,
				Typeflag: tar.TypeDir,
				Mode:     0555,
				ModTime:  creationTime.Time,
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
