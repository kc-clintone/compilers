package main

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

func printHelp() {
	fmt.Println("Zing Compiler & Interpreter (Workshop Edition)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  zing tokens <file.zing>            Inspect Lexer token stream")
	fmt.Println("  zing ast <file.zing>               Inspect Parser AST tree")
	fmt.Println("  zing run <file.zing>               Run Zing program with interpreter")
	fmt.Println("  zing compile <file.zing>           Transpile Zing program to Go and execute")
	fmt.Println("  zing transpile [-o out.go] <file>   Transpile Zing program to Go source file")
	fmt.Println("  zing repl                          Start interactive REPL")
	fmt.Println("  zing -h, --help                    Show this help message")
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		runREPL()
		return
	}

	cmd := args[0]
	if cmd == "-h" || cmd == "--help" {
		printHelp()
		return
	}

	switch cmd {
	case "tokens", "lex":
		if len(args) < 2 {
			fmt.Println("Error: missing filename for tokens")
			os.Exit(1)
		}
		printFileTokens(args[1])
	case "ast", "parse":
		if len(args) < 2 {
			fmt.Println("Error: missing filename for ast")
			os.Exit(1)
		}
		printFileAST(args[1])
	case "run":
		if len(args) < 2 {
			fmt.Println("Error: missing filename for run")
			os.Exit(1)
		}
		runFile(args[1])
	case "compile":
		if len(args) < 2 {
			fmt.Println("Error: missing filename for compile")
			os.Exit(1)
		}
		compileAndRun(args[1])
	case "transpile":
		outputFile := ""
		inputFile := ""
		for i := 1; i < len(args); i++ {
			if args[i] == "-o" && i+1 < len(args) {
				outputFile = args[i+1]
				i++
			} else {
				inputFile = args[i]
			}
		}
		if inputFile == "" {
			fmt.Println("Error: missing input file for transpile")
			os.Exit(1)
		}
		transpileFile(inputFile, outputFile)
	case "repl":
		runREPL()
	default:
		if strings.HasPrefix(cmd, "-") {
			fmt.Printf("Unknown flag: %s\n", cmd)
			printHelp()
			os.Exit(1)
		}
		runFile(cmd)
	}
}

func printFileTokens(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		os.Exit(1)
	}

	toks, diags := lexer.Lex(filename, data)
	token.PrintTokens(toks)

	if len(diags) > 0 {
		fmt.Println("\nLexer Diagnostics:")
		for _, d := range diags {
			fmt.Println(d)
		}
	}
}

func printFileAST(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		os.Exit(1)
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
}

func runFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		os.Exit(1)
	}

	prog, diags := parser.Parse(filename, data)
	if len(diags) > 0 {
		for _, d := range diags {
			fmt.Println(d)
		}
		os.Exit(1)
	}

	interp := interpreter.New()
	evalDiags := interp.Interpret(prog)
	if len(evalDiags) > 0 {
		for _, d := range evalDiags {
			fmt.Println(d)
		}
		os.Exit(1)
	}
}

func transpileFile(filename string, outputFile string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		os.Exit(1)
	}

	prog, diags := parser.Parse(filename, data)
	if len(diags) > 0 {
		for _, d := range diags {
			fmt.Println(d)
		}
		os.Exit(1)
	}

	c := compiler.New()
	goCode, compDiags := c.Compile(prog)
	if len(compDiags) > 0 {
		for _, d := range compDiags {
			fmt.Println(d)
		}
		os.Exit(1)
	}

	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(goCode), 0644)
		if err != nil {
			fmt.Printf("Error writing output file %s: %v\n", outputFile, err)
			os.Exit(1)
		}
		fmt.Printf("Transpiled Go code written to %s\n", outputFile)
	} else {
		fmt.Print(goCode)
	}
}

func compileAndRun(filename string) {
	tmpDir, err := os.MkdirTemp("", "zing-build-*")
	if err != nil {
		fmt.Printf("Error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	goFile := filepath.Join(tmpDir, "main.go")
	transpileFile(filename, goFile)

	exeFile := filepath.Join(tmpDir, "program")
	buildCmd := exec.Command("go", "build", "-o", exeFile, goFile)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		os.Exit(1)
	}

	runCmd := exec.Command(exeFile)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	runCmd.Stdin = os.Stdin
	if err := runCmd.Run(); err != nil {
		fmt.Printf("Execution error: %v\n", err)
		os.Exit(1)
	}
}

func runREPL() {
	fmt.Println("Zing REPL (Workshop Version)")
	fmt.Println("Type 'exit' or Ctrl+D to quit.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	interp := interpreter.New()

	for {
		fmt.Print("zing> ")
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

		evalDiags := interp.Interpret(prog)
		if len(evalDiags) > 0 {
			for _, d := range evalDiags {
				fmt.Println(d)
			}
		}
	}
}
