package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	fmt.Println("Verifying workshop repository...")

	if err := verifyGoVersion(); err != nil {
		fmt.Printf("[FAIL] %v\n", err)
		os.Exit(1)
	}

	requiredDirs := []string{"cmd", "internal", "examples", "docs", "exercises", "tests", "solutions"}
	for _, dir := range requiredDirs {
		if err := verifyDir(dir); err != nil {
			fmt.Printf("[FAIL] %v\n", err)
			os.Exit(1)
		}
	}

	if err := verifyDir("examples"); err != nil {
		fmt.Printf("[FAIL] %v\n", err)
		os.Exit(1)
	}

	if err := verifyDir("docs"); err != nil {
		fmt.Printf("[FAIL] %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[OK] Repository structure looks ready for the workshop.")
}

func verifyGoVersion() error {
	cmd := exec.Command("go", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unable to run go version: %w\n%s", err, output)
	}

	fmt.Printf("[OK] %s", string(output))
	return nil
}

func verifyDir(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("unable to resolve path %q: %w", path, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("missing required directory %q", path)
	}
	if !info.IsDir() {
		return fmt.Errorf("path %q is not a directory", path)
	}

	return nil
}
