package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/jpiechowka/go-silent-assassin/internal/builder"
)

var (
	buildCommand = &cobra.Command{
		Use:   "build",
		Short: "Builds the executable",
		Long:  `TODO`,
		Run: func(cmd *cobra.Command, args []string) {
			buildCmd()
		},
	}
)

func init() {
	// TODO: Add flags
}

func buildCmd() {
	b := builder.NewBuilder()
	err := b.BuildExecutable("calc.exe", "compiled-loader.exe") // TODO: Make configurable
	if err != nil {
		log.Printf("[ERROR] Error: %s", err)
		return
	}
}
