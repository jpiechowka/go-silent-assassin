package cmd

import (
	"github.com/spf13/cobra"
)

// TODO
const msg = "TODO"

var rootCmd = &cobra.Command{
	Use:     "go-silent-assassin",
	Version: "0.1.0",
	Short:   msg,
	Long:    msg,
}

// Execute executes the root command.
func Execute() error {
	rootCmd.AddCommand(buildCommand)
	return rootCmd.Execute()
}
