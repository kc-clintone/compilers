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
		name string
		args []string
		code int
		out  string
		err  string
	}{
		{"check", []string{"check", valid}, 0, "", ""},
		{"run", []string{"run", valid, "--", "hello"}, 0, "1 hello\n", ""},
		{"source error", []string{"check", invalid}, 1, "", "checker:"},
		{"missing command", nil, 2, "", "usage: zing-interpreter"},
		{"unknown command", []string{"build", valid}, 2, "", "usage: zing-interpreter"},
		{"arguments need separator", []string{"run", valid, "hello"}, 2, "", "usage: zing-interpreter"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := RunInterpreter(context.Background(), test.args, Streams{Stdout: &stdout, Stderr: &stderr})
			if code != test.code || stdout.String() != test.out || !strings.Contains(stderr.String(), test.err) {
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

	binary := filepath.Join(t.TempDir(), "program")
	stderr.Reset()
	if code := RunCompiler(context.Background(), []string{"build", "-o", binary, valid}, Streams{Stderr: &stderr}); code != 0 {
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
		{"usage", []string{"transpile", valid}, 2, "usage: zing-compiler"},
		{"unknown", []string{"run", valid}, 2, "usage: zing-compiler"},
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

func writeSource(t *testing.T, source string) string {
	t.Helper()
	filename := filepath.Join(t.TempDir(), "program.zing")
	if err := os.WriteFile(filename, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	return filename
}
