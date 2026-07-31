package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInterpreterCommandsAndExitCodes(t *testing.T) {
	valid := writeSource(t, `var values []string = args(); print(len(values), values[0]);`)
	invalid := writeSource(t, `var value int = "wrong";`)
	tests := []struct {
		name  string
		args  []string
		stdin string
		code  int
		out   string
		err   string
	}{
		{"check", []string{"check", valid}, "", 0, "", ""},
		{"run", []string{valid, "--", "hello"}, "", 0, "1 hello\n", ""},
		{"default repl", nil, "1 + 2;\nexit\n", 0, "3\n", ""},
		{"--repl mode", []string{"--repl", valid}, "exit\n", 0, "", ""},
		{"source error", []string{"check", invalid}, "", 1, "", "checker:"},
		{"arguments need separator", []string{valid, "hello"}, "", 2, "", "usage: nuru-interpreter"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader(test.stdin)
			code := RunInterpreter(context.Background(), test.args, Streams{Stdout: &stdout, Stderr: &stderr, Stdin: stdin})
			if code != test.code || !strings.Contains(stdout.String(), test.out) || !strings.Contains(stderr.String(), test.err) {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}
}

func TestCompilerCommandsAndExitCodes(t *testing.T) {
	valid := writeSource(t, `print("compiled");`)
	invalid := writeSource(t, `print(unknown);`)
	var stdout, stderr bytes.Buffer
	if code := RunCompiler(context.Background(), []string{"check", valid}, Streams{Stdout: &stdout, Stderr: &stderr}); code != 0 {
		t.Fatalf("check code=%d stderr=%q", code, stderr.String())
	}

	generated := filepath.Join(t.TempDir(), "program.go")
	stderr.Reset()
	if code := RunCompiler(context.Background(), []string{"transpile", "-o", generated, valid}, Streams{Stderr: &stderr}); code != 0 {
		t.Fatalf("transpile code=%d stderr=%q", code, stderr.String())
	}
	if contents, err := os.ReadFile(generated); err != nil || !bytes.Contains(contents, []byte("func main()")) {
		t.Fatalf("generated source = %q, %v", contents, err)
	}

	defaultGenerated, defaultInput, ok := outputArgs([]string{valid}, filepath.Base(valid)+".go", &stderr)
	if !ok || defaultGenerated != "program.nuru.go" || defaultInput != valid {
		t.Fatalf("default transpile output=%q input=%q ok=%v", defaultGenerated, defaultInput, ok)
	}
	defaultBinary, defaultInput, ok := outputArgs([]string{valid}, "nuru.out", &stderr)
	if !ok || defaultBinary != "nuru.out" || defaultInput != valid {
		t.Fatalf("default binary output=%q input=%q ok=%v", defaultBinary, defaultInput, ok)
	}

	binary := filepath.Join(t.TempDir(), "program")
	stderr.Reset()
	if code := RunCompiler(context.Background(), []string{"-o", binary, valid}, Streams{Stderr: &stderr}); code != 0 {
		t.Fatalf("build code=%d stderr=%q", code, stderr.String())
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil || string(output) != "compiled\n" {
		t.Fatalf("binary output=%q err=%v", output, err)
	}

	tests := []struct {
		name string
		args []string
		code int
		text string
	}{
		{"source error", []string{"check", invalid}, 1, "checker:"},
		{"usage", []string{"transpile"}, 2, "usage: nuru-compiler"},
		{"invalid flags", []string{"-x", valid}, 2, "usage: nuru-compiler"},
		{"same path", []string{"transpile", "-o", valid, valid}, 2, "output path must differ"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stderr.Reset()
			code := RunCompiler(context.Background(), test.args, Streams{Stderr: &stderr})
			if code != test.code || !strings.Contains(stderr.String(), test.text) {
				t.Fatalf("code=%d stderr=%q", code, stderr.String())
			}
		})
	}
}

func TestModuleWarning(t *testing.T) {
	file1 := writeSource(t, `export var x int = 1;`)
	file2 := writeSource(t, `import "math";`)
	var stdout, stderr bytes.Buffer

	code := RunInterpreter(context.Background(), []string{"--repl", file1, file2}, Streams{Stdout: &stdout, Stderr: &stderr, Stdin: strings.NewReader("exit\n")})
	if code != 0 || !strings.Contains(stderr.String(), "warning: multiple files/modules have yet to be implemented") {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
}

func writeSource(t *testing.T, source string) string {
	t.Helper()
	filename := filepath.Join(t.TempDir(), "program.nuru")
	if err := os.WriteFile(filename, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	return filename
}
