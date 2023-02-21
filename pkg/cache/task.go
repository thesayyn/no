package cache

import (
	"fmt"
	"reflect"

	"github.com/fatih/color"
	"github.com/thesayyn/no/pkg/build"
)

func RunIfNotCached(cache Cache, task build.Task) (build.Output, error) {

	hit, err := cache.Hit(task)
	if err != nil {
		return nil, err
	}

	if hit {
		output, err := cache.Fetch(task)

		if err == nil {
			fmt.Println(color.HiGreenString("> cache hit for %s", reflect.TypeOf(task).Elem().Name()))
		}

		return output, err
	}

	fmt.Println(color.YellowString("> cache miss for %s", reflect.TypeOf(task).Elem().Name()))

	if err = task.Build(); err != nil {
		return nil, err
	}

	if err := cache.Store(task); err != nil {
		return nil, err
	}

	return task.Output()
}
