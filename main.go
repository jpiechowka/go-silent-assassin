package main

import (
	"os"

	"github.com/jpiechowka/go-silent-assassin/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
