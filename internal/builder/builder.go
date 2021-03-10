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
	tempLoaderOutFileName    = "loader.exe"
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

	if err := b.moveCompiledLoader(workspacePath, outputFilePath); err != nil {
		return err
	}

	if err := b.deleteWorkspace(workspacePath); err != nil {
		return err
	}

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

	loaderGoModPath := filepath.Join(workspace, "go.mod")
	loaderPath := filepath.Join(workspace, loaderMainFileName)

	log.Printf("[+] Generating and saving go.mod file for the loader %s", loaderGoModPath)

	if err := os.WriteFile(loaderGoModPath, b.codeGenerator.GenerateLoaderGoModFile(), 0666); err != nil {
		return err
	}

	log.Printf("[+] Generating and saving loader code to %s", loaderPath)

	if err := os.WriteFile(loaderPath, b.codeGenerator.GenerateDefaultCode(), 0666); err != nil {
		return err
	}

	log.Println("[+] Running go mod tidy command")

	modCmd := exec.Command("go", "mod", "tidy")
	modCmd.Dir = workspace

	modCmdCombinedOut, err := modCmd.CombinedOutput()
	if err != nil {
		log.Printf("[ERROR] Go mod tidy output:\n\n%s\n", modCmdCombinedOut)
		return err
	}

	log.Printf("[+] Executing Go compiler to compile the loader with shellcode")

	// TODO: Strip debugging symbols
	buildCmd := exec.Command("go", "build", "-a", "-o", tempLoaderOutFileName)
	buildCmd.Dir = workspace

	buildCmdCombinedOut, err := buildCmd.CombinedOutput()
	if err != nil {
		log.Printf("[ERROR] Go build command output:\n\n%s\n", buildCmdCombinedOut)
		return err
	}

	log.Printf("[+] Go build command output:\n\n%s\n", buildCmdCombinedOut)
	log.Printf("[+] Loader has been compiled. Output file: %s", tempLoaderOutFileName)

	return nil
}

func (b *Builder) moveCompiledLoader(workspace string, outputFilePath string) error {
	compiledLoaderPath := filepath.Join(workspace, tempLoaderOutFileName)
	log.Printf("[+] Moving compiled loader from %s to %s", compiledLoaderPath, outputFilePath)

	compiledLoaderData, err := os.ReadFile(compiledLoaderPath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputFilePath, compiledLoaderData, 0666); err != nil {
		return err
	}

	return nil
}
