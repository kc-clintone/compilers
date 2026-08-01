package compiler

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kc-clintone/compilers/internal/ast"
	"github.com/kc-clintone/compilers/internal/interpreter"
	"github.com/kc-clintone/compilers/internal/parser"
)

func TestCompilerScaffold(t *testing.T) {
	prog := &ast.Program{}
	c := New()
	goCode, diags := c.Compile(prog)
	if len(diags) > 0 {
		t.Fatalf("unexpected compiler diagnostics: %v", diags)
	}

	if !strings.Contains(goCode, "package main") {
		t.Errorf("expected package main in output: %s", goCode)
	}
}

func TestRecursiveFibonacciMatchesInterpreterAndGeneratedGo(t *testing.T) {
	source := []byte(`
func fib(n) {
    if (n <= 1) { return n }
    return fib(n - 1) + fib(n - 2)
}
print(fib(10))
`)
	prog, parseDiags := parser.Parse("fib.nuru", source)
	if len(parseDiags) != 0 {
		t.Fatalf("unexpected parser diagnostics: %v", parseDiags)
	}

	interpreted := captureOutput(t, func() {
		if diags := interpreter.New().Interpret(prog); len(diags) != 0 {
			t.Fatalf("unexpected interpreter diagnostics: %v", diags)
		}
	})

	goCode, compileDiags := New().Compile(prog)
	if len(compileDiags) != 0 {
		t.Fatalf("unexpected compiler diagnostics: %v", compileDiags)
	}
	tempDir := t.TempDir()
	mainFile := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(mainFile, []byte(goCode), 0o600); err != nil {
		t.Fatalf("write generated Go: %v", err)
	}
	cmd := exec.Command("go", "run", mainFile)
	cmd.Env = append(os.Environ(), "GOCACHE="+filepath.Join(tempDir, "go-cache"))
	compiled, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated Go failed: %v\n%s\nGenerated source:\n%s", err, compiled, goCode)
	}

	if interpreted != string(compiled) || interpreted != "55\n" {
		t.Fatalf("backend mismatch: interpreter %q, generated Go %q", interpreted, compiled)
	}
}

func captureOutput(t *testing.T, run func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	original := os.Stdout
	os.Stdout = writer
	defer func() { os.Stdout = original }()
	run()
	os.Stdout = original
	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return string(output)
}
