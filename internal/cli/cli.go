// Package cli implements the shared command-line behavior for the separate
// Zing interpreter and compiler executables.
package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kc-clintone/compilers/internal/ast"
	"github.com/kc-clintone/compilers/internal/checker"
	"github.com/kc-clintone/compilers/internal/compiler"
	"github.com/kc-clintone/compilers/internal/interpreter"
	"github.com/kc-clintone/compilers/internal/parser"
)

// Streams contains the process streams used by a CLI invocation.
type Streams struct {
	Stdout io.Writer
	Stderr io.Writer
}

func (streams Streams) normalized() Streams {
	if streams.Stdout == nil {
		streams.Stdout = io.Discard
	}
	if streams.Stderr == nil {
		streams.Stderr = io.Discard
	}
	return streams
}

// RunInterpreter executes one zing-interpreter invocation and returns its exit
// code without terminating the host process.
func RunInterpreter(ctx context.Context, args []string, streams Streams) int {
	streams = streams.normalized()
	if len(args) < 2 {
		interpreterUsage(streams.Stderr)
		return 2
	}
	switch args[0] {
	case "check":
		if len(args) != 2 {
			interpreterUsage(streams.Stderr)
			return 2
		}
		_, _, ok := frontEnd(args[1], streams.Stderr)
		if !ok {
			return 1
		}
		return 0
	case "run":
		programArgs := []string{}
		if len(args) > 2 {
			if args[2] != "--" {
				interpreterUsage(streams.Stderr)
				return 2
			}
			programArgs = args[3:]
		}
		program, info, ok := frontEnd(args[1], streams.Stderr)
		if !ok {
			return 1
		}
		if err := interpreter.Run(ctx, program, info, interpreter.Options{Args: programArgs, Stdout: streams.Stdout}); err != nil {
			fmt.Fprintln(streams.Stderr, err)
			return 1
		}
		return 0
	default:
		interpreterUsage(streams.Stderr)
		return 2
	}
}

// RunCompiler executes one zing-compiler invocation and returns its exit code
// without terminating the host process.
func RunCompiler(ctx context.Context, args []string, streams Streams) int {
	streams = streams.normalized()
	if len(args) < 2 {
		compilerUsage(streams.Stderr)
		return 2
	}
	if args[0] == "check" {
		if len(args) != 2 {
			compilerUsage(streams.Stderr)
			return 2
		}
		_, _, ok := frontEnd(args[1], streams.Stderr)
		if !ok {
			return 1
		}
		return 0
	}
	if args[0] != "transpile" && args[0] != "build" {
		compilerUsage(streams.Stderr)
		return 2
	}
	output, input, ok := outputArgs(args[1:], streams.Stderr)
	if !ok {
		compilerUsage(streams.Stderr)
		return 2
	}
	program, info, valid := frontEnd(input, streams.Stderr)
	if !valid {
		return 1
	}
	source, err := compiler.Generate(program, info)
	if err != nil {
		fmt.Fprintln(streams.Stderr, err)
		return 1
	}
	if args[0] == "transpile" {
		err = os.WriteFile(output, source, 0o644)
	} else {
		err = compiler.Build(ctx, source, output)
	}
	if err != nil {
		fmt.Fprintln(streams.Stderr, err)
		return 1
	}
	return 0
}

func frontEnd(filename string, stderr io.Writer) (*ast.Program, *checker.Info, bool) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return nil, nil, false
	}
	program, diagnostics := parser.Parse(filename, contents)
	if len(diagnostics) > 0 {
		for _, item := range diagnostics {
			fmt.Fprintln(stderr, item.Error())
		}
		return nil, nil, false
	}
	info, diagnostics := checker.Check(program)
	if len(diagnostics) > 0 {
		for _, item := range diagnostics {
			fmt.Fprintln(stderr, item.Error())
		}
		return nil, nil, false
	}
	return program, info, true
}

func outputArgs(args []string, stderr io.Writer) (string, string, bool) {
	if len(args) != 3 || args[0] != "-o" || args[1] == "" || args[2] == "" {
		return "", "", false
	}
	output, _ := filepath.Abs(args[1])
	input, _ := filepath.Abs(args[2])
	if output == input {
		fmt.Fprintln(stderr, "output path must differ from input")
		return "", "", false
	}
	return args[1], args[2], true
}

func interpreterUsage(stderr io.Writer) {
	fmt.Fprintln(stderr, "usage: zing-interpreter check <file> | zing-interpreter run <file> [-- program-args...]")
}

func compilerUsage(stderr io.Writer) {
	fmt.Fprintln(stderr, "usage: zing-compiler check <file> | zing-compiler transpile -o <go-file> <file> | zing-compiler build -o <binary> <file>")
}
