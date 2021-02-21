package loader

import "golang.org/x/sys/windows"

type Loader struct {
	kernel32Dll *windows.LazyDLL
}

func NewLoader() *Loader {
	return &Loader{
		kernel32Dll: windows.NewLazyDLL("kernel32.dll"),
	}
}
