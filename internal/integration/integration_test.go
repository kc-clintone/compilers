package integration

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/kc-clintone/compilers/internal/checker"
	"github.com/kc-clintone/compilers/internal/compiler"
	"github.com/kc-clintone/compilers/internal/interpreter"
	"github.com/kc-clintone/compilers/internal/parser"
)

func TestInterpreterAndCompiledProgramsAgree(t *testing.T) {
	root := filepath.Join("..", "..", "examples")

	for _, file := range []string{"01-basics.nuru", "02-control-flow.nuru", "03-collections.nuru", "04-functions.nuru", "05-files.nuru", "source-analyzer.nuru"} {
		t.Run(file, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join(root, file))
			if err != nil {
				t.Fatal(err)
			}

			program, diagnostics := parser.Parse(file, source)

			if len(diagnostics) > 0 {
				t.Fatalf("parse: %v", diagnostics)
			}

			info, diagnostics := checker.Check(program)

			if len(diagnostics) > 0 {
				t.Fatalf("check: %v", diagnostics)
			}

			var interpreted bytes.Buffer
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

			defer cancel()
			var programArgs []string
			var interpretedFile, compiledFile string

			switch file {
			case "05-files.nuru":
				interpretedFile = filepath.Join(t.TempDir(), "interpreted.txt")
				compiledFile = filepath.Join(t.TempDir(), "compiled.txt")
				programArgs = []string{filepath.Join(root, "fixtures", "sample.txt"), interpretedFile}
			case "source-analyzer.nuru":
				programArgs = []string{filepath.Join(root, "fixtures", "analyzer-input.nuru")}
			}

			if err := interpreter.Run(ctx, program, info, interpreter.Options{Args: programArgs, Stdout: &interpreted}); err != nil {
				t.Fatal(err)
			}

			goSource, err := compiler.Generate(program, info)
			if err != nil {
				t.Fatal(err)
			}

			binary := filepath.Join(t.TempDir(), "program")

			if err := compiler.Build(ctx, goSource, binary); err != nil {
				t.Fatalf("build: %v\n%s", err, goSource)
			}

			compiledArgs := programArgs

			if file == "05-files.nuru" {
				compiledArgs = []string{programArgs[0], compiledFile}
			}

			output, err := exec.CommandContext(ctx, binary, compiledArgs...).CombinedOutput()
			if err != nil {
				t.Fatalf("run binary: %v: %s", err, output)
			}

			if !bytes.Equal(output, interpreted.Bytes()) {
				t.Fatalf("output differs\ninterpreter: %q\ncompiled: %q", interpreted.String(), output)
			}

			expected, err := os.ReadFile(filepath.Join(root, "expected", file[:len(file)-len(filepath.Ext(file))]+".stdout"))
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(output, expected) {
				t.Fatalf("unexpected output\nwant: %q\ngot:  %q", expected, output)
			}

			if file == "05-files.nuru" {
				want, _ := os.ReadFile(programArgs[0])
				gotInterpreter, err := os.ReadFile(interpretedFile)
				if err != nil || !bytes.Equal(gotInterpreter, want) {
					t.Fatalf("interpreter file differs: %v", err)
				}

				gotCompiled, err := os.ReadFile(compiledFile)
				if err != nil || !bytes.Equal(gotCompiled, want) {
					t.Fatalf("compiled file differs: %v", err)
				}
			}
		})
	}
}

func TestRuntimeFailuresAgree(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.txt")
	tests := map[string]string{
		"explicit fail":        `fail("stopped");`,
		"file failure":         `print(readFile(` + strconv.Quote(missing) + `));`,
		"bounds failure":       `var values []int = []int{1}; print(values[2]);`,
		"slice failure":        `var value string = "a"; print(value[1:0]);`,
		"integer conversion":   `print(int("invalid"));`,
		"character conversion": `print(char(256));`,
		"division by zero":     `print(1 / 0);`,
		"nil map assignment":   `var values map[string]int; values["x"] = 1;`,
	}

	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			program, diagnostics := parser.Parse("runtime.nuru", []byte(source))
			if len(diagnostics) != 0 {
				t.Fatalf("parse: %v", diagnostics)
			}
			info, diagnostics := checker.Check(program)
			if len(diagnostics) != 0 {
				t.Fatalf("check: %v", diagnostics)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			var interpreted bytes.Buffer
			interpretedErr := interpreter.Run(ctx, program, info, interpreter.Options{Stdout: &interpreted})
			if interpretedErr == nil {
				t.Fatal("interpreter unexpectedly succeeded")
			}

			goSource, err := compiler.Generate(program, info)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(t.TempDir(), "program")
			if err := compiler.Build(ctx, goSource, binary); err != nil {
				t.Fatalf("build: %v\n%s", err, goSource)
			}
			compiledOutput, compiledErr := exec.CommandContext(ctx, binary).CombinedOutput()
			if compiledErr == nil {
				t.Fatalf("compiled program unexpectedly succeeded: %q", compiledOutput)
			}
			want := interpreted.String() + interpretedErr.Error() + "\n"
			if string(compiledOutput) != want {
				t.Fatalf("observable failure differs\ninterpreter: %q\ncompiled:    %q", want, compiledOutput)
			}
		})
	}
}

func TestSourceAnalyzerFallbackIsCurrent(t *testing.T) {
	root := filepath.Join("..", "..", "examples")
	source, err := os.ReadFile(filepath.Join(root, "source-analyzer.nuru"))
	if err != nil {
		t.Fatal(err)
	}
	program, diagnostics := parser.Parse("examples/source-analyzer.nuru", source)
	if len(diagnostics) != 0 {
		t.Fatalf("parse: %v", diagnostics)
	}
	info, diagnostics := checker.Check(program)
	if len(diagnostics) != 0 {
		t.Fatalf("check: %v", diagnostics)
	}
	generated, err := compiler.Generate(program, info)
	if err != nil {
		t.Fatal(err)
	}
	fallback, err := os.ReadFile(filepath.Join(root, "generated", "source-analyzer.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(generated, fallback) {
		t.Fatal("checked-in source analyzer fallback is stale; regenerate it with nuru-compiler")
	}
}
