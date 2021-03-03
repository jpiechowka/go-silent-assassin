package main

import (
	"log"

	"github.com/jpiechowka/go-silent-assassin/loader"
)

func main() {
	// if err := cmd.Execute(); err != nil {
	// 	os.Exit(1)
	// }
	//
	// os.Exit(0)

	// TODO: delete after testing
	l, err := loader.NewLoader()
	if err != nil {
		log.Fatal(err)
	}

	err = l.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
