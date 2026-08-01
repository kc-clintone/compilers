// Package cli implements the shared command-line behavior for the separate
// Nuru interpreter and compiler executables.
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

// RunInterpreter executes one nuru-interpreter invocation and returns its exit
// code without terminating the host process.
func RunInterpreter(ctx context.Context, args []string, streams Streams) int {
	streams = streams.normalized()
	if len(args) > 0 && isHelp(args[0]) {
		interpreterUsage(streams.Stdout)
		return 0
	}

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

// RunCompiler executes one nuru-compiler invocation and returns its exit code
// without terminating the host process.
func RunCompiler(ctx context.Context, args []string, streams Streams) int {
	streams = streams.normalized()
	if len(args) > 0 && isHelp(args[0]) {
		compilerUsage(streams.Stdout)
		return 0
	}

	if len(args) < 1 {
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
		output, input, ok = outputArgs(args[1:], filepath.Base(args[len(args)-1])+".go", streams.Stderr)
	} else {
		output, input, ok = outputArgs(args, "nuru.out", streams.Stderr)
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

func outputArgs(args []string, defaultOutput string, stderr io.Writer) (string, string, bool) {
	var output, input string

	switch {
	case len(args) == 1 && args[0] != "":
		output, input = defaultOutput, args[0]
	case len(args) == 3 && args[0] == "-o" && args[1] != "" && args[2] != "":
		output, input = args[1], args[2]
	default:
		return "", "", false
	}

	absoluteOutput, _ := filepath.Abs(output)
	absoluteInput, _ := filepath.Abs(input)

	if absoluteOutput == absoluteInput {
		fmt.Fprintln(stderr, "output path must differ from input")
		return "", "", false
	}

	return output, input, true
}

func isHelp(arg string) bool {
	return arg == "-h" || arg == "--help"
}

func interpreterUsage(w io.Writer) {
	name := filepath.Base(os.Args[0])

	fmt.Fprintf(w, `%s(1)

NAME
    %s - check and directly execute Nuru programs

SYNOPSIS
    %s
    %s --repl [FILE...]
    %s check FILE
    %s FILE [-- ARG...]
    %s (-h | --help)

DESCRIPTION
    Starts an interactive REPL when invoked without arguments. A source file can
    be checked without running it, executed directly, or loaded before starting
    the REPL.

COMMANDS
    check FILE
        Parse and type-check FILE without executing it.

    FILE [-- ARG...]
        Execute FILE directly. Arguments after -- are passed to the program.

OPTIONS
    --repl [FILE...]
        Start the REPL, optionally loading one or more files first.

    -h, --help
        Print this help page and exit.

    --
        End interpreter options. Remaining arguments are passed to the Nuru
        program and are available through args().
`, name, name, name, name, name, name, name)
}

func compilerUsage(w io.Writer) {
	name := filepath.Base(os.Args[0])

	fmt.Fprintf(w, `%s(1)

NAME
    %s - check, transpile, and build Nuru programs

SYNOPSIS
    %s check FILE
    %s transpile [-o GO-FILE] FILE
    %s [-o BINARY] FILE
    %s (-h | --help)

DESCRIPTION
    Checks Nuru source, translates it to Go, or builds it as a native executable.
    Building is the default operation when no command is given.

COMMANDS
    check FILE
        Parse and type-check FILE without producing output.

    transpile [-o GO-FILE] FILE
        Write generated Go source. The default output is <input-basename>.go.

    FILE
        Build FILE as a native executable. This is the default operation.

OPTIONS
    -o PATH
        Set the generated Go file or executable path. The default executable is
        ./nuru.out.

    -h, --help
        Print this help page and exit.
`, name, name, name, name, name, name)
}
