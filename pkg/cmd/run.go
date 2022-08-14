package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/spf13/cobra"
	"github.com/thesayyn/no/pkg/build"
	"github.com/thesayyn/no/pkg/options"
)

func NewBuild() *cobra.Command {

	buildopts := &options.BuildOptions{}

	var runUnder string

	cmd := &cobra.Command{
		Use:   "run PATH",
		Short: "A variant of `docker run` that containerizes PATH first.",
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			if err := buildopts.Load(); err != nil {
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

			builder := build.Builder{
				BuildOpts: *buildopts,
			}

			for _, v := range paths {
				log.Printf("Building %q", v)

				img, err := builder.Build(v)
				if err != nil {
					return err
				}

				nm, err := name.NewTag(fmt.Sprintf("no.local/no/%s:latest", filepath.Clean(v)))
				if err != nil {
					return err
				}

				log.Printf("Loading to daemon %q", v)

				_, err = daemon.Write(nm, img)

				if err != nil {
					return err
				}

				argv := []string{"run", "--rm", nm.String()}

				argv = append(argv, dockerArgs...)

				log.Printf("$ %s %s", runUnder, strings.Join(argv, " "))

				dockerCmd := exec.CommandContext(cmd.Context(), runUnder, argv...)
				dockerCmd.Env = os.Environ()
				dockerCmd.Stderr = os.Stderr
				dockerCmd.Stdout = os.Stdout
				dockerCmd.Stdin = os.Stdin
				if err := dockerCmd.Run(); err != nil {
					log.Fatalf("error executing \"%v run\": %v", runUnder, err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&runUnder, "run-under", "r", "podman", "")

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
