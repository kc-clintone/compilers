// Package cli implements the shared command-line behavior for the workshop's
// separate Nuru interpreter and compiler executables.
package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kc-clintone/compilers/internal/ast"
	"github.com/kc-clintone/compilers/internal/compiler"
	"github.com/kc-clintone/compilers/internal/interpreter"
	"github.com/kc-clintone/compilers/internal/lexer"
	"github.com/kc-clintone/compilers/internal/parser"
	"github.com/kc-clintone/compilers/internal/token"
)

// RunCompiler executes one nuru-compiler invocation and returns its exit code.
func RunCompiler(args []string) int {
	if len(args) == 0 {
		printCompilerHelp()
		return 0
	}
	if isHelp(args[0]) {
		printCompilerHelp()
		return 0
	}

	switch args[0] {
	case "tokens", "lex":
		return runInspection(args[1:], "tokens", printFileTokens)
	case "ast", "parse":
		return runInspection(args[1:], "ast", printFileAST)
	case "transpile":
		input, output, unrecognized := parseOutputArgs(args[1:], func(input string) string {
			return filepath.Base(input) + ".go"
		})
		if input == "" {
			fmt.Println("Error: missing input file for transpile")
			return 1
		}
		warnUnrecognized(unrecognized)
		return transpileFile(input, output)
	default:
		input, output, unrecognized := parseOutputArgs(args, func(string) string { return "nuru.out" })
		if input == "" {
			fmt.Println("Error: missing input file to compile")
			printCompilerHelp()
			return 1
		}
		warnUnrecognized(unrecognized)
		return compileToBinary(input, output)
	}
}

// RunInterpreter executes one nuru-interpreter invocation and returns its exit code.
func RunInterpreter(args []string) int {
	if len(args) == 0 {
		runREPL()
		return 0
	}
	if isHelp(args[0]) {
		printInterpreterHelp()
		return 0
	}

	switch args[0] {
	case "tokens", "lex":
		return runInspection(args[1:], "tokens", printFileTokens)
	case "ast", "parse":
		return runInspection(args[1:], "ast", printFileAST)
	case "repl":
		warnUnrecognized(args[1:])
		runREPL()
		return 0
	default:
		file, unrecognized := firstFile(args)
		warnUnrecognized(unrecognized)
		if file == "" {
			runREPL()
			return 0
		}
		return runFile(file)
	}
}

func isHelp(arg string) bool {
	return arg == "-h" || arg == "--help"
}

func runInspection(args []string, name string, inspect func(string) int) int {
	if len(args) == 0 {
		fmt.Printf("Error: missing filename for %s\n", name)
		return 1
	}
	warnUnrecognized(args[1:])
	return inspect(args[0])
}

func firstFile(args []string) (string, []string) {
	var file string
	var unrecognized []string
	for _, arg := range args {
		if file == "" && !strings.HasPrefix(arg, "-") {
			file = arg
		} else {
			unrecognized = append(unrecognized, arg)
		}
	}
	return file, unrecognized
}

func parseOutputArgs(args []string, defaultOutput func(string) string) (string, string, []string) {
	var input, output string
	var unrecognized []string
	for i := 0; i < len(args); i++ {
		if args[i] == "-o" && i+1 < len(args) {
			output = args[i+1]
			i++
		} else if input == "" && !strings.HasPrefix(args[i], "-") {
			input = args[i]
		} else {
			unrecognized = append(unrecognized, args[i])
		}
	}
	if input != "" && output == "" {
		output = defaultOutput(input)
	}
	return input, output, unrecognized
}

func warnUnrecognized(args []string) {
	for _, arg := range args {
		fmt.Fprintf(os.Stderr, "Warning: unrecognized option/argument '%s' ignored\n", arg)
	}
}

func printFileTokens(filename string) int {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		return 1
	}
	toks, diags := lexer.Lex(filename, data)
	token.PrintTokens(toks)
	if len(diags) > 0 {
		fmt.Println("\nLexer Diagnostics:")
		for _, d := range diags {
			fmt.Println(d)
		}
	}
	return 0
}

func printFileAST(filename string) int {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		return 1
	}
	prog, diags := parser.Parse(filename, data)
	if prog != nil {
		ast.Print(prog)
	}
	if len(diags) > 0 {
		fmt.Println("\nParser Diagnostics:")
		for _, d := range diags {
			fmt.Println(d)
		}
	}
	return 0
}

func transpileFile(filename, outputFile string) int {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		return 1
	}
	prog, diags := parser.Parse(filename, data)
	if len(diags) > 0 {
		for _, d := range diags {
			fmt.Println(d)
		}
		return 1
	}
	c := compiler.New()
	goCode, compDiags := c.Compile(prog)
	if len(compDiags) > 0 {
		for _, d := range compDiags {
			fmt.Println(d)
		}
		return 1
	}
	if err := os.WriteFile(outputFile, []byte(goCode), 0o644); err != nil {
		fmt.Printf("Error writing output file %s: %v\n", outputFile, err)
		return 1
	}
	fmt.Printf("Transpiled Go code written to %s\n", outputFile)
	return 0
}

func compileToBinary(filename, outputBinary string) int {
	tmpDir, err := os.MkdirTemp("", "nuru-build-*")
	if err != nil {
		fmt.Printf("Error creating temp dir: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tmpDir)
	goFile := filepath.Join(tmpDir, "main.go")
	if code := transpileFile(filename, goFile); code != 0 {
		return code
	}
	buildCmd := exec.Command("go", "build", "-o", outputBinary, goFile)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		return 1
	}
	fmt.Printf("Compiled binary output written to %s\n", outputBinary)
	return 0
}

func runFile(filename string) int {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		return 1
	}
	prog, diags := parser.Parse(filename, data)
	if len(diags) > 0 {
		for _, d := range diags {
			fmt.Println(d)
		}
		return 1
	}
	interp := interpreter.New()
	for _, d := range interp.Interpret(prog) {
		fmt.Println(d)
		return 1
	}
	return 0
}

func runREPL() {
	fmt.Println("Nuru REPL (Workshop Edition)")
	fmt.Println("Type 'exit' or Ctrl+D to quit.")
	fmt.Println()
	scanner := bufio.NewScanner(os.Stdin)
	interp := interpreter.New()
	for {
		fmt.Print("nuru> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "exit" {
			break
		}
		if line == "" {
			continue
		}
		prog, diags := parser.Parse("<repl>", []byte(line))
		if len(diags) > 0 {
			for _, d := range diags {
				fmt.Println(d)
			}
			continue
		}
		for _, d := range interp.Interpret(prog) {
			fmt.Println(d)
		}
	}
}

func printCompilerHelp() {
	fmt.Println("Nuru Compiler (Workshop Edition)")
	fmt.Println("\nUsage:")
	fmt.Println("  nuru-compiler [-o binary] <file.nuru>   Compile to a native executable (default: ./nuru.out)")
	fmt.Println("  nuru-compiler tokens <file.nuru>        Inspect Lexer token stream")
	fmt.Println("  nuru-compiler ast <file.nuru>           Inspect Parser AST tree")
	fmt.Println("  nuru-compiler transpile [-o out.go] <f> Transpile to Go (default: ./<input-basename>.go)")
	fmt.Println("  nuru-compiler -h, --help                Show help message")
}

func printInterpreterHelp() {
	fmt.Println("Nuru Interpreter (Workshop Edition)")
	fmt.Println("\nUsage:")
	fmt.Println("  nuru-interpreter <file.nuru>         Run file using interpreter (default)")
	fmt.Println("  nuru-interpreter tokens <file.nuru>  Inspect Lexer token stream")
	fmt.Println("  nuru-interpreter ast <file.nuru>     Inspect Parser AST tree")
	fmt.Println("  nuru-interpreter repl                Start interactive REPL")
	fmt.Println("  nuru-interpreter -h, --help          Show help message")
}
