package cmd

import (
	"path/filepath"

	"github.com/davecgh/go-spew/spew"

	"github.com/spf13/cobra"
	"github.com/thesayyn/no/pkg/build"
	"github.com/thesayyn/no/pkg/cache"
	"github.com/thesayyn/no/pkg/node"
)

func NewRun() *cobra.Command {

	var output string

	cmd := &cobra.Command{
		Use:   "build path...",
		Short: "Build and publish container images from the given sub-paths.",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) == 0 {
				args = []string{"."}
			}

			dc := cache.NewDiskCache()

			if err := dc.Setup(); err != nil {
				return err
			}

			for _, v := range args {
				options := build.NewOptions(DEFAULT_MTIME, v, v, DEFAULT_BASE, *DEFAULT_PLATFORM)
				lock, err := node.FindPackageLock(v)
				if err != nil {
					return err
				}
				package_json := filepath.Join(v, "package.json")

				appTask := build.NewAppTask(options, v)
				nodeModulesTask := build.NewNodeModulesTask(options, v, *lock)

				appOutput, err := cache.RunIfNotCached(dc, appTask)
				if err != nil {
					return err
				}
				nodeModulesOutput, err := cache.RunIfNotCached(dc, nodeModulesTask)
				if err != nil {
					return err
				}

				imageTask := build.NewImageTask(options, []build.Output{appOutput, nodeModulesOutput}, package_json)
				imageOutput, err := cache.RunIfNotCached(dc, imageTask)
				if err != nil {
					return err
				}

				spew.Dump(appOutput, nodeModulesOutput, imageOutput)

				// if output != "" {
				// 	p, err := layout.FromPath(output)
				// 	if err != nil {
				// 		p, err = layout.Write(output, empty.Index)
				// 		if err != nil {
				// 			return err
				// 		}
				// 	}

				// 	p.AppendImage(img)
				// } else {
				// 	ref, err := name.ParseReference(fmt.Sprintf("no/%s:latest", filepath.Clean(v)))
				// 	if err != nil {
				// 		return err
				// 	}
				// 	if err = tarball.WriteToFile(fmt.Sprintf("%s.tar", v), ref, img); err != nil {
				// 		return err
				// 	}
				// }
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "")

	return cmd
}
