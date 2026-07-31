package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kc-clintone/compilers/internal/ast"
	"github.com/kc-clintone/compilers/internal/interpreter"
	"github.com/kc-clintone/compilers/internal/lexer"
	"github.com/kc-clintone/compilers/internal/parser"
	"github.com/kc-clintone/compilers/internal/token"
)

func printHelp() {
	fmt.Println("Nuru Interpreter (Workshop Edition)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  nuru-interpreter <file.nuru>         Run file using interpreter (default)")
	fmt.Println("  nuru-interpreter tokens <file.nuru>  Inspect Lexer token stream")
	fmt.Println("  nuru-interpreter ast <file.nuru>     Inspect Parser AST tree")
	fmt.Println("  nuru-interpreter repl                Start interactive REPL")
	fmt.Println("  nuru-interpreter -h, --help          Show help message")
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		runREPL()
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
	case "repl":
		warnUnrecognized(args[1:])
		runREPL()
	default:
		// Default: interpret first file argument, warn on unrecognized options/arguments
		var file string
		var unrec []string
		for _, arg := range args {
			if file == "" && !strings.HasPrefix(arg, "-") {
				file = arg
			} else {
				unrec = append(unrec, arg)
			}
		}
		warnUnrecognized(unrec)
		if file == "" {
			runREPL()
		} else {
			runFile(file)
		}
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

		evalDiags := interp.Interpret(prog)
		if len(evalDiags) > 0 {
			for _, d := range evalDiags {
				fmt.Println(d)
			}
		}
	}
}
