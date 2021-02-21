package cmd

import "github.com/spf13/cobra"

const (
	defaultIsVerbose = false
)

var (
	isVerbose bool

	buildCommand = &cobra.Command{
		Use:   "build",
		Short: "Builds the executable",
		Long:  `TODO`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return buildCmd()
		},
	}
)

func init() {
	buildCommand.Flags().BoolVarP(&isVerbose, "verbose", "v", defaultIsVerbose, "enables verbose logging")
}

func buildCmd() error {
	return nil
}
