package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/google/go-containerregistry/pkg/logs"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
	"github.com/thesayyn/no/pkg/build"
	"github.com/thesayyn/no/pkg/cache"
	"github.com/thesayyn/no/pkg/node"
	"github.com/thesayyn/no/pkg/registry"
)

func NewBuild() *cobra.Command {
	var runUnder string
	var workingDir string
	var noCache bool

	cmd := &cobra.Command{
		Use:   "run PATH",
		Short: "A variant of `docker run` that containerizes PATH first.",
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			var cach cache.Cache

			if noCache {
				cach = cache.NewNoopCache()
			} else {
				cach = cache.NewDiskCache()
			}

			if err := cach.Setup(); err != nil {
				return err
			}

			paths := args
			dashes := cmd.Flags().ArgsLenAtDash()
			if dashes != -1 {
				paths = args[:dashes]
			}
			if len(paths) == 0 {
				log.Fatalf("no run: no paths listed")
			}

			dockerArgs := []string{}
			dashes = unparsedDashes()
			if dashes != -1 {
				dockerArgs = os.Args[dashes:]
			}

			p := args[0]

			currentWorkingDir := p

			if workingDir != "" {
				currentWorkingDir = workingDir
			}
			options := build.NewOptions(DEFAULT_MTIME, p, currentWorkingDir, DEFAULT_BASE, *DEFAULT_PLATFORM)
			lock, err := node.FindPackageLock(p)
			if err != nil {
				return err
			}
			package_json := filepath.Join(p, "package.json")

			appTask := build.NewAppTask(options, p)
			nodeModulesTask := build.NewNodeModulesTask(options, p, *lock)

			appOutput, err := cache.RunIfNotCached(cach, appTask)
			if err != nil {
				return err
			}
			nodeModulesOutput, err := cache.RunIfNotCached(cach, nodeModulesTask)
			if err != nil {
				return err
			}

			imageTask := build.NewImageTask(options, []build.Output{appOutput, nodeModulesOutput}, package_json)
			imageOutput, err := cache.RunIfNotCached(cach, imageTask)
			if err != nil {
				return err
			}

			port, err := registry.Serve()
			if err != nil {
				return err
			}

			nm, err := name.NewTag(fmt.Sprintf("localhost:%d/%s:latest", *port, filepath.Clean(p)), name.Insecure)
			if err != nil {
				return err
			}

			index, err := layout.ImageIndexFromPath(imageOutput.Path())
			if err != nil {
				return err
			}
			logs.Progress = log.New(io.Discard, "", log.LstdFlags)
			err = remote.WriteIndex(nm, index)
			if err != nil {
				return err
			}

			nm, err = name.NewTag(fmt.Sprintf("host.containers.internal:%d/%s:latest", *port, filepath.Clean(p)), name.Insecure)
			if err != nil {
				return err
			}

			runArgs := []string{"run", "--rm"}
			if runUnder == "podman" {
				runArgs = append(runArgs, "--tls-verify=false")
			}
			runArgs = append(runArgs, nm.String())
			runArgs = append(runArgs, dockerArgs...)
			fmt.Println(color.BlueString("$ %s %s\n", runUnder, strings.Join(runArgs, " ")))

			dockerCmd := exec.CommandContext(cmd.Context(), runUnder, runArgs...)
			dockerCmd.Env = os.Environ()
			dockerCmd.Stderr = os.Stderr
			dockerCmd.Stdout = os.Stdout
			dockerCmd.Stdin = os.Stdin
			if err := dockerCmd.Run(); err != nil {
				log.Fatalf("error executing \"%v run\": %v", runUnder, err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&runUnder, "run-under", "r", "podman", "")
	cmd.Flags().BoolVarP(&noCache, "no-cache", "", false, "")
	cmd.Flags().StringVarP(&workingDir, "working-dir", "", "", "")

	return cmd
}

func unparsedDashes() int {
	for i, s := range os.Args {
		if s == "--" {
			return i
		}
	}
	return -1
}
