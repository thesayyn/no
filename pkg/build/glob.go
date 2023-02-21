package build

import (
	"github.com/bmatcuk/doublestar/v4"
)

type Glob struct {
	include string
	exclude *string
}

func NewGlob(include string, exclude *string) Glob {
	return Glob{
		include: include,
		exclude: exclude,
	}
}

func (glob *Glob) Matches(path string) bool {
	included, err := doublestar.Match(glob.include, path)
	if err != nil {
		panic(err)
	}

	excluded := false
	if glob.exclude != nil {
		excluded, err = doublestar.Match(*glob.exclude, path)

		if err != nil {
			panic(err)
		}
	}

	return included && !excluded
}
