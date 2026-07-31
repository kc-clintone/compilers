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
	Stdin  io.Reader
}

func (streams Streams) normalized() Streams {
	if streams.Stdout == nil {
		streams.Stdout = io.Discard
	}
	if streams.Stderr == nil {
		streams.Stderr = io.Discard
	}
	if streams.Stdin == nil {
		streams.Stdin = os.Stdin
	}
	return streams
}

// RunInterpreter executes one zing-interpreter invocation and returns its exit
// code without terminating the host process.
func RunInterpreter(ctx context.Context, args []string, streams Streams) int {
	streams = streams.normalized()
	if len(args) == 0 {
		if err := interpreter.RunREPL(ctx, nil, nil, interpreter.Options{Stdout: streams.Stdout, Stderr: streams.Stderr, Stdin: streams.Stdin}); err != nil {
			fmt.Fprintln(streams.Stderr, err)
			return 1
		}
		return 0
	}
	if args[0] == "--repl" {
		files := args[1:]
		var initialPrograms []*ast.Program
		var initialInfos []*checker.Info
		warned := false
		if len(files) > 1 {
			fmt.Fprintln(streams.Stderr, "warning: multiple files/modules have yet to be implemented")
			warned = true
		}
		for _, file := range files {
			prog, info, ok := frontEnd(file, streams.Stderr)
			if !ok {
				return 1
			}
			if !warned && hasModuleFeatures(prog) {
				fmt.Fprintln(streams.Stderr, "warning: multiple files/modules have yet to be implemented")
				warned = true
			}
			initialPrograms = append(initialPrograms, prog)
			initialInfos = append(initialInfos, info)
		}
		if err := interpreter.RunREPL(ctx, initialPrograms, initialInfos, interpreter.Options{Stdout: streams.Stdout, Stderr: streams.Stderr, Stdin: streams.Stdin}); err != nil {
			fmt.Fprintln(streams.Stderr, err)
			return 1
		}
		return 0
	}
	if args[0] == "check" {
		if len(args) != 2 {
			interpreterUsage(streams.Stderr)
			return 2
		}
		prog, _, ok := frontEnd(args[1], streams.Stderr)
		if !ok {
			return 1
		}
		if hasModuleFeatures(prog) {
			fmt.Fprintln(streams.Stderr, "warning: multiple files/modules have yet to be implemented")
		}
		return 0
	}

	filename := args[0]
	programArgs := []string{}
	if len(args) > 1 {
		if args[1] != "--" {
			interpreterUsage(streams.Stderr)
			return 2
		}
		programArgs = args[2:]
	}
	program, info, ok := frontEnd(filename, streams.Stderr)
	if !ok {
		return 1
	}
	if hasModuleFeatures(program) {
		fmt.Fprintln(streams.Stderr, "warning: multiple files/modules have yet to be implemented")
	}
	if err := interpreter.Run(ctx, program, info, interpreter.Options{Args: programArgs, Stdout: streams.Stdout, Stderr: streams.Stderr, Stdin: streams.Stdin}); err != nil {
		fmt.Fprintln(streams.Stderr, err)
		return 1
	}
	return 0
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
		prog, _, ok := frontEnd(args[1], streams.Stderr)
		if !ok {
			return 1
		}
		if hasModuleFeatures(prog) {
			fmt.Fprintln(streams.Stderr, "warning: multiple files/modules have yet to be implemented")
		}
		return 0
	}
	isTranspile := args[0] == "transpile"
	var output, input string
	var ok bool
	if isTranspile {
		output, input, ok = outputArgs(args[1:], streams.Stderr)
	} else {
		output, input, ok = outputArgs(args, streams.Stderr)
	}
	if !ok {
		compilerUsage(streams.Stderr)
		return 2
	}
	program, info, valid := frontEnd(input, streams.Stderr)
	if !valid {
		return 1
	}
	if hasModuleFeatures(program) {
		fmt.Fprintln(streams.Stderr, "warning: multiple files/modules have yet to be implemented")
	}
	source, err := compiler.Generate(program, info)
	if err != nil {
		fmt.Fprintln(streams.Stderr, err)
		return 1
	}
	if isTranspile {
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

func hasModuleFeatures(program *ast.Program) bool {
	if program == nil {
		return false
	}
	for _, d := range program.Decls {
		if _, ok := d.(*ast.ExportStmt); ok {
			return true
		}
		if _, ok := d.(*ast.ImportStmt); ok {
			return true
		}
	}
	for _, s := range program.Stmts {
		if _, ok := s.(*ast.ExportStmt); ok {
			return true
		}
		if _, ok := s.(*ast.ImportStmt); ok {
			return true
		}
	}
	return false
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
	fmt.Fprintln(stderr, "usage: zing-interpreter [--repl [files...]] | zing-interpreter check <file> | zing-interpreter <file> [-- program-args...]")
}

func compilerUsage(stderr io.Writer) {
	fmt.Fprintln(stderr, "usage: zing-compiler check <file> | zing-compiler transpile -o <go-file> <file> | zing-compiler -o <binary> <file>")
}
