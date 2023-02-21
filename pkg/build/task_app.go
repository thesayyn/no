package build

import (
	"archive/tar"
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
)

type AppTask struct {
	options Options
	path    string
	_output Output
	_inputs []Input
}

// build implements Task
func (app *AppTask) Build() error {
	p := filepath.Join(os.TempDir(), "app.tar")
	hash := sha256.New()

	file, err := os.Create(p)
	if err != nil {
		return err
	}
	defer file.Close()

	mw := io.MultiWriter(hash, file)
	tw := tar.NewWriter(mw)
	defer tw.Close()

	exclude := "**/node_modules/**"
	glob := NewGlob("**", &exclude)

	if err := BuildTar(tw, app.path, "", *app.options.mtime, glob); err != nil {
		return err
	}

	app._output = DiskOutput{_path: p, _hash: hash}

	return nil
}

// inputs implements Task
func (app *AppTask) Inputs() ([]Input, error) {
	if app._inputs == nil {
		exclude := "**/node_modules/**"
		glob := NewGlob("**", &exclude)
		tree := NewDiskTreeInput(app.path, glob)
		app._inputs = []Input{
			app.options.only([]string{"mtime"}),
			tree,
		}
	}

	return app._inputs, nil
}

// outputs implements Task
func (app AppTask) Output() (Output, error) {
	return app._output, nil
}

func NewAppTask(options Options, path string) Task {
	return &AppTask{
		options: options,
		path:    path,
	}
}
