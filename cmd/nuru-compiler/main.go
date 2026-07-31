package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kc-clintone/compilers/internal/ast"
	"github.com/kc-clintone/compilers/internal/compiler"
	"github.com/kc-clintone/compilers/internal/lexer"
	"github.com/kc-clintone/compilers/internal/parser"
	"github.com/kc-clintone/compilers/internal/token"
)

func printHelp() {
	fmt.Println("Nuru Compiler (Workshop Edition)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  nuru-compiler [-o binary] <file.nuru>   Compile Nuru program to native executable (default)")
	fmt.Println("  nuru-compiler tokens <file.nuru>         Inspect Lexer token stream")
	fmt.Println("  nuru-compiler ast <file.nuru>            Inspect Parser AST tree")
	fmt.Println("  nuru-compiler transpile [-o out.go] <f> Transpile Nuru program to Go source")
	fmt.Println("  nuru-compiler -h, --help                 Show help message")
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
		return
	}

	if args[0] == "-h" || args[0] == "--help" {
		printHelp()
		return
	}

	cmd := args[0]
	switch cmd {
	case "tokens", "lex":
		if len(args) < 2 {
			fmt.Println("Error: missing filename for tokens")
			os.Exit(1)
		}
		warnUnrecognized(args[2:])
		printFileTokens(args[1])

	case "ast", "parse":
		if len(args) < 2 {
			fmt.Println("Error: missing filename for ast")
			os.Exit(1)
		}
		warnUnrecognized(args[2:])
		printFileAST(args[1])

	case "transpile":
		outputFile := ""
		inputFile := ""
		var unrec []string
		for i := 1; i < len(args); i++ {
			if args[i] == "-o" && i+1 < len(args) {
				outputFile = args[i+1]
				i++
			} else if inputFile == "" && !strings.HasPrefix(args[i], "-") {
				inputFile = args[i]
			} else {
				unrec = append(unrec, args[i])
			}
		}
		if inputFile == "" {
			fmt.Println("Error: missing input file for transpile")
			os.Exit(1)
		}
		warnUnrecognized(unrec)
		transpileFile(inputFile, outputFile)

	default:
		// Default: build binary for the first file in arguments list
		outputBinary := ""
		inputFile := ""
		var unrec []string
		for i := 0; i < len(args); i++ {
			if args[i] == "-o" && i+1 < len(args) {
				outputBinary = args[i+1]
				i++
			} else if inputFile == "" && !strings.HasPrefix(args[i], "-") {
				inputFile = args[i]
			} else {
				unrec = append(unrec, args[i])
			}
		}
		if inputFile == "" {
			fmt.Println("Error: missing input file to compile")
			printHelp()
			os.Exit(1)
		}
		warnUnrecognized(unrec)

		if outputBinary == "" {
			base := filepath.Base(inputFile)
			outputBinary = strings.TrimSuffix(base, filepath.Ext(base))
		}
		compileToBinary(inputFile, outputBinary)
	}
}

func warnUnrecognized(args []string) {
	for _, arg := range args {
		fmt.Fprintf(os.Stderr, "Warning: unrecognized option/argument '%s' ignored\n", arg)
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

func compileToBinary(filename string, outputBinary string) {
	tmpDir, err := os.MkdirTemp("", "nuru-build-*")
	if err != nil {
		fmt.Printf("Error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	goFile := filepath.Join(tmpDir, "main.go")
	transpileFile(filename, goFile)

	buildCmd := exec.Command("go", "build", "-o", outputBinary, goFile)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Compiled binary output written to %s\n", outputBinary)
}
