package builder

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jpiechowka/go-silent-assassin/internal/codegen"
)

const (
	pe2shcGitHubRepoUrl      = "https://github.com/hasherezade/pe_to_shellcode"
	pe2shcGitHubDownloadUrl  = pe2shcGitHubRepoUrl + "/releases/latest/download/pe2shc.exe"
	pe2shcExeFileName        = "pe2shc.exe"
	tempPe2shcInputFileName  = "in.exe"
	tempPe2shcOutputFileName = "out.exe"
	loaderMainFileName       = "main.go"
)

type Builder struct {
	codeGenerator *codegen.CodeGenerator
}

func NewBuilder() *Builder {
	return &Builder{
		codeGenerator: codegen.NewCodeGenerator(),
	}
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

	if err := b.convertPEToShellcode(inputFilePath, workspacePath); err != nil {
		return err
	}

	if err := b.compileLoader(workspacePath); err != nil {
		return err
	}

	// if err := b.deleteWorkspace(workspacePath); err != nil {
	// 	return err
	// }

	return nil
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

	log.Printf("[+] Downloading latest Hasherezade's pe_to_shellcode from GitHub (%s)", pe2shcGitHubRepoUrl)
	resp, err := http.Get(pe2shcGitHubDownloadUrl)
	if err != nil {
		return err
	}

	log.Printf("[+] Saving downloaded executable to %s", pe2shcExePath)
	_, err = io.Copy(pe2shcExeFile, resp.Body)

	if err := pe2shcExeFile.Close(); err != nil {
		return err
	}

	if err := resp.Body.Close(); err != nil {
		return err
	}

	return nil
}

func (b *Builder) convertPEToShellcode(inputFilePath string, workspace string) error {
	log.Printf("[+] Moving input file %s to temp directory as %s", inputFilePath, tempPe2shcInputFileName)

	input, err := os.ReadFile(inputFilePath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(workspace, tempPe2shcInputFileName), input, 0666); err != nil {
		return err
	}

	log.Printf("[+] Executing %s to convert PE file to shellcode so that it can be injected", pe2shcExeFileName)

	cmd := exec.Command(fmt.Sprintf("./%s", pe2shcExeFileName), tempPe2shcInputFileName, tempPe2shcOutputFileName)
	cmd.Dir = workspace

	cmdCombinedOut, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[ERROR] pe2shc command output:\n\n%s\n", cmdCombinedOut)
		return err
	}

	log.Printf("[+] pe2shc command output:\n\n%s\n", cmdCombinedOut)

	return nil
}

func (b *Builder) compileLoader(workspace string) error {
	log.Printf("[+] Preparing loader")

	loaderPath := filepath.Join(workspace, loaderMainFileName)

	log.Printf("[+] Generating and saving loader code to %s", loaderPath)

	if err := os.WriteFile(loaderPath, b.codeGenerator.GenerateDefaultCode(), 0666); err != nil {
		return err
	}

	log.Printf("[+] Executing Go compiler to compile the loader with shellcode")

	// TODO: Strip debugging symbols
	cmd := exec.Command("go", "build")
	cmd.Dir = workspace

	cmdCombinedOut, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[ERROR] Go build command output:\n\n%s\n", cmdCombinedOut)
		return err
	}

	log.Printf("[+] Go build command output:\n\n%s\n", cmdCombinedOut)

	return nil
}
