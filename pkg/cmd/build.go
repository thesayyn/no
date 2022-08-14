package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/tarball"

	"github.com/spf13/cobra"
	"github.com/thesayyn/no/pkg/build"
	"github.com/thesayyn/no/pkg/options"
)

func NewRun() *cobra.Command {

	buildopts := &options.BuildOptions{}

	var output string

	cmd := &cobra.Command{
		Use:   "build path...",
		Short: "Build and publish container images from the given sub-paths.",
		RunE: func(cmd *cobra.Command, args []string) error {

			if err := buildopts.Load(); err != nil {
				return err
			}

			if len(args) == 0 {
				args = []string{"."}
			}

			builder := build.Builder{
				BuildOpts: *buildopts,
			}

			for _, v := range args {
				img, err := builder.Build(v)
				if err != nil {
					return err
				}

				if output != "" {
					p, err := layout.FromPath(output)
					if err != nil {
						p, err = layout.Write(output, empty.Index)
						if err != nil {
							return err
						}
					}

					p.AppendImage(img)
				} else {
					ref, err := name.ParseReference(fmt.Sprintf("no/%s:latest", filepath.Clean(v)))
					if err != nil {
						return err
					}
					if err = tarball.WriteToFile(fmt.Sprintf("%s.tar", v), ref, img); err != nil {
						return err
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "")

	return cmd
}
