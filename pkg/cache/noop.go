package cache

import (
	"fmt"

	"github.com/thesayyn/no/pkg/build"
)

type NoopCache struct{}

// Setup implements Cache
func (dc *NoopCache) Setup() error {
	return nil
}

// Store implements Cache
func (dc *NoopCache) Store(task build.Task) error {
	return nil
}

// Fetch implements Cache
func (dc *NoopCache) Fetch(task build.Task) (build.Output, error) {
	return nil, fmt.Errorf("not implemented")
}

// Hit implements Cache
func (dc *NoopCache) Hit(task build.Task) (bool, error) {
	return false, nil
}

func NewNoopCache() Cache {
	return &NoopCache{}
}
