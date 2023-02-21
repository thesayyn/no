package build

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/thesayyn/no/pkg/node"
)

type ImageTask struct {
	options      Options
	package_json string
	_layers      []Output
	_output      Output
}

// build implements Task
func (it *ImageTask) Build() error {

	rmt, err := remote.Get(*it.options.image)
	if err != nil {
		return err
	}

	var img v1.Image

	if rmt.MediaType.IsIndex() {
		idx, err := rmt.ImageIndex()
		if err != nil {
			return err
		}

		m, err := idx.IndexManifest()
		if err != nil {
			return err
		}
		for _, manifest := range m.Manifests {

			if manifest.Platform.OS == it.options.platform.OS && manifest.Platform.Architecture == it.options.platform.Architecture {
				iimg, err := idx.Image(manifest.Digest)

				if err != nil {
					return err
				}

				img = iimg

				break
			}
		}
	} else {
		iimg, err := rmt.Image()

		if err != nil {
			return err
		}

		img = iimg
	}

	addendums := []mutate.Addendum{}

	for _, _layer := range it._layers {
		layer, err := tarball.LayerFromFile(_layer.Path())
		if err != nil {
			return err
		}
		addendums = append(addendums, mutate.Addendum{
			Layer: layer,
			History: v1.History{
				Author:    "no",
				Created:   v1.Time{Time: *it.options.mtime},
				CreatedBy: "no build",
			},
			MediaType: types.DockerLayer,
		})
	}

	img, err = mutate.Append(img, addendums...)
	if err != nil {
		return err
	}

	pkg, err := node.ParsePackageJson(it.package_json)

	if err != nil {
		return err
	}

	cfg, err := img.ConfigFile()

	if err != nil {
		return err
	}

	cfg = cfg.DeepCopy()

	workingDir := *it.options.workdir

	if !filepath.IsAbs(workingDir) {
		workingDir = fmt.Sprintf("/%s", workingDir)
	}

	bin := path.Join(*it.options.root, pkg.Bin)

	if !filepath.IsAbs(bin) {
		bin = fmt.Sprintf("/%s", bin)
	}

	cfg.Config.WorkingDir = workingDir
	cfg.Config.Cmd = []string{bin}
	cfg.Author = "github.com/thesayyn/no"

	img, err = mutate.ConfigFile(img, cfg)

	if err != nil {
		return err
	}

	output := filepath.Join(os.TempDir(), "layout")

	p, err := layout.Write(output, empty.Index)
	if err != nil {
		return err
	}
	p.AppendImage(img)

	digest, err := img.Digest()
	if err != nil {
		return err
	}

	it._output = NewDiskTreeOutput(output, &digest)

	return err
}

// inputs implements Task
func (it ImageTask) Inputs() ([]Input, error) {
	pkg_json, err := NewDiskInput(it.package_json)
	if err != nil {
		return nil, err
	}
	inputs := []Input{
		it.options.only([]string{"workdir", "root", "image", "mtime", "platform"}),
		pkg_json,
	}
	for _, layer := range it._layers {
		inputs = append(inputs, NewInputFromOutput(layer))
	}
	return inputs, nil
}

// outputs implements Task
func (it ImageTask) Output() (Output, error) {
	return it._output, nil
}

func NewImageTask(options Options, layers []Output, package_json string) Task {
	return &ImageTask{options: options, _layers: layers, package_json: package_json}
}
