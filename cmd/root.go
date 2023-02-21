package cmd

import (
	"os"

	cranecmd "github.com/google/go-containerregistry/cmd/crane/cmd"
	"github.com/google/go-containerregistry/pkg/logs"
	"github.com/spf13/cobra"
)

var Root = New()

func New() *cobra.Command {
	var verbose bool
	root := &cobra.Command{
		Use:               "no",
		Short:             "Rapidly iterate with NodeJS and Containers.",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				logs.Warn.SetOutput(os.Stderr)
				logs.Debug.SetOutput(os.Stderr)
			}
			logs.Progress.SetOutput(os.Stderr)
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logs")

	root.AddCommand(cranecmd.NewCmdAuthLogin("no"))
	root.AddCommand(NewBuild())
	root.AddCommand(NewRun())

	return root
}
