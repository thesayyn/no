package build

import (
	"hash"
)

type Input interface {
	Hash() (hash.Hash, error)
}

type Output interface {
	Path() string
	Hash() (hash.Hash, error)
}

type Task interface {
	Inputs() ([]Input, error)
	Build() error
	Output() (Output, error)
}
