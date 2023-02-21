package build

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type DiskInput struct {
	_hash hash.Hash
}

func (di DiskInput) Hash() (hash.Hash, error) {
	return di._hash, nil
}

func NewDiskInput(path string) (*DiskInput, error) {
	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}
	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return nil, err
	}

	return &DiskInput{_hash: h}, nil
}

type InputFromOutput struct {
	_output Output
}

// Hash implements Input
func (ifo InputFromOutput) Hash() (hash.Hash, error) {
	return ifo._output.Hash()
}

func NewInputFromOutput(output Output) Input {
	return InputFromOutput{_output: output}
}

type DiskOutput struct {
	_path string
	_hash hash.Hash
}

// Hash implements Output
func (do DiskOutput) Hash() (hash.Hash, error) {
	return do._hash, nil
}

// Path implements Output
func (do DiskOutput) Path() string {
	return do._path
}

func NewDiskOutput(path string) (Output, error) {
	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}
	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return nil, err
	}

	return DiskOutput{_hash: h, _path: path}, nil
}

type DiskTreeOutput struct {
	_path string
	_hash hash.Hash
}

// hash implements Output
func (lo DiskTreeOutput) Hash() (hash.Hash, error) {
	if lo._hash != nil {
		return lo._hash, nil
	}

	return nil, fmt.Errorf("not implemented")
}

// path implements Output
func (lo DiskTreeOutput) Path() string {
	return lo._path
}

func NewDiskTreeOutput(path string, h *v1.Hash) Output {
	var hash hash.Hash
	if h != nil {
		hash := sha256.New()
		hash.Write([]byte(h.String()))
	}
	return DiskTreeOutput{_hash: hash, _path: path}
}
