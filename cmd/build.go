package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"

	cp "github.com/otiai10/copy"
	"github.com/spf13/cobra"
	"github.com/thesayyn/no/pkg/build"
	"github.com/thesayyn/no/pkg/cache"
	"github.com/thesayyn/no/pkg/node"
)

func saveTarball(p string, out string, tag string) error {
	l, err := layout.ImageIndexFromPath(p)
	if err != nil {
		return err
	}
	ref, err := name.ParseReference(tag)
	if err != nil {
		return err
	}

	mn, err := l.IndexManifest()
	if err != nil {
		return err
	}
	img, err := l.Image(mn.Manifests[0].Digest)
	if err != nil {
		return err
	}
	if err = tarball.WriteToFile(p, ref, img); err != nil {
		return err
	}
	return nil
}

func NewRun() *cobra.Command {

	var workingDir string
	var format string

	cmd := &cobra.Command{
		Use:   "build path output",
		Args:  cobra.ExactValidArgs(2),
		Short: "Build and publish container images from the given sub-path.",
		RunE: func(cmd *cobra.Command, args []string) error {

			project := args[0]
			output := args[1]

			currentWorkingDir := project
			if workingDir != "" {
				currentWorkingDir = workingDir
			}

			dc := cache.NewDiskCache()
			if err := dc.Setup(); err != nil {
				return err
			}

			options := build.NewOptions(DEFAULT_MTIME, project, currentWorkingDir, DEFAULT_BASE, *DEFAULT_PLATFORM)
			lock, err := node.FindPackageLock(project)
			if err != nil {
				return err
			}
			package_json := filepath.Join(project, "package.json")

			appTask := build.NewAppTask(options, project)
			nodeModulesTask := build.NewNodeModulesTask(options, project, *lock)

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

			if format == "tarball" {
				saveTarball(imageOutput.Path(), output, fmt.Sprintf("no/%s:latest", filepath.Clean(project)))
			} else if format == "oci" {
				cp.Copy(imageOutput.Path(), output)
			} else {
				ref, err := name.ParseReference(output)
				if err != nil {
					return err
				}
				l, err := layout.ImageIndexFromPath(imageOutput.Path())
				if err != nil {
					return err
				}
				mn, err := l.IndexManifest()
				if err != nil {
					return err
				}
				img, err := l.Image(mn.Manifests[0].Digest)
				if err != nil {
					return err
				}
				err = remote.Write(ref, img, remote.WithAuthFromKeychain(authn.DefaultKeychain))
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "", "", "")
	cmd.Flags().StringVarP(&workingDir, "working-dir", "", "", "")
	return cmd
}
