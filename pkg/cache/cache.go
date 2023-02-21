package cache

import (
	"github.com/thesayyn/no/pkg/build"
)

type Cache interface {
	Hit(task build.Task) (bool, error)
	Fetch(task build.Task) (build.Output, error)
	Store(task build.Task) error
	Setup() error
}
