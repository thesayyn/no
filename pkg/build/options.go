package build

import (
	"crypto/sha512"
	"hash"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type Options struct {
	mtime    *time.Time
	root     *string
	workdir  *string
	image    *name.Reference
	platform *v1.Platform
}

// hash implements Artifact
func (opt Options) Hash() (hash.Hash, error) {

	hasher := sha512.New()

	if opt.root != nil {
		if _, err := hasher.Write([]byte(*opt.root)); err != nil {
			return nil, err
		}
	}

	if opt.workdir != nil {
		if _, err := hasher.Write([]byte(*opt.workdir)); err != nil {
			return nil, err
		}
	}

	if opt.mtime != nil {
		if _, err := hasher.Write([]byte(opt.mtime.String())); err != nil {
			return nil, err
		}
	}

	if opt.image != nil {
		if _, err := hasher.Write([]byte((*opt.image).String())); err != nil {
			return nil, err
		}
	}

	if opt.platform != nil {
		if _, err := hasher.Write([]byte(opt.platform.String())); err != nil {
			return nil, err
		}
	}

	return hasher, nil
}

func (opt Options) only(keys []string) Options {
	only := Options{}
	for _, key := range keys {
		switch key {
		case "mtime":
			only.mtime = opt.mtime
		case "root":
			only.root = opt.root
		case "workdir":
			only.workdir = opt.workdir
		}

	}
	return only
}

func NewOptions(mtime time.Time, root string, workdir string, image name.Reference, platform v1.Platform) Options {
	return Options{
		mtime:    &mtime,
		root:     &root,
		workdir:  &workdir,
		image:    &image,
		platform: &platform,
	}
}
