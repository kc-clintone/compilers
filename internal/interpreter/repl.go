// Package interpreter executes checked Zing syntax trees.
package interpreter

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kc-clintone/compilers/internal/ast"
	"github.com/kc-clintone/compilers/internal/checker"
	"github.com/kc-clintone/compilers/internal/parser"
)

// RunREPL starts an interactive REPL session, pre-loading any initial programs.
func RunREPL(ctx context.Context, initialPrograms []*ast.Program, initialInfos []*checker.Info, options Options) error {
	if options.Stdout == nil {
		options.Stdout = io.Discard
	}
	if options.Stderr == nil {
		options.Stderr = io.Discard
	}
	if options.Stdin == nil {
		options.Stdin = os.Stdin
	}
	if options.Files == nil {
		options.Files = osFileSystem{}
	}
	if ctx == nil {
		ctx = context.Background()
	}

	globals := &environment{values: map[string]value{}}
	r := &runner{
		context: ctx,
		globals: globals,
		current: globals,
		options: options,
		funcs:   map[string]*ast.FuncDecl{},
	}

	accumulatedDecls := preloadPrograms(r, initialPrograms, initialInfos)
	scanner := bufio.NewScanner(options.Stdin)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		fmt.Fprint(options.Stdout, "zing> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			break
		}

		// 1. Parse line (attempting trailing semicolon fallback if needed)
		parseLine := line
		if !strings.HasSuffix(parseLine, ";") && !strings.HasSuffix(parseLine, "}") {
			parseLine = parseLine + ";"
		}

		parsedProg, diagnostics := parser.Parse("repl", []byte(parseLine))
		if len(diagnostics) > 0 && parseLine != line {
			altProg, altDiags := parser.Parse("repl", []byte(line))
			if len(altDiags) == 0 {
				parsedProg = altProg
				diagnostics = nil
			}
		}

		if len(diagnostics) > 0 {
			for _, diag := range diagnostics {
				fmt.Fprintln(options.Stderr, diag.Error())
			}
			continue
		}

		// 2. Augment AST with accumulated declarations and check types
		augmentedProg := &ast.Program{
			Decls: append(append([]ast.Decl{}, accumulatedDecls...), parsedProg.Decls...),
			Stmts: parsedProg.Stmts,
		}
		ast.AssignNodeIDs(augmentedProg)

		info, checkDiags := checker.Check(augmentedProg)
		if len(checkDiags) > 0 {
			for _, diag := range checkDiags {
				fmt.Fprintln(options.Stderr, diag.Error())
			}
			continue
		}

		r.info = info
		r.program = augmentedProg

		// 3. Define new top-level declarations
		for _, decl := range parsedProg.Decls {
			accumulatedDecls = append(accumulatedDecls, decl)
			d := unwrapDeclNode(decl)
			if function, ok := d.(*ast.FuncDecl); ok {
				r.funcs[function.Name] = function
			} else if variable, ok := d.(*ast.VarDecl); ok {
				variableType, _ := info.GlobalType(variable.Name)
				res := r.zero(variableType)
				if variable.Init != nil {
					var err error
					res, err = r.eval(variable.Init)
					if err != nil {
						fmt.Fprintln(options.Stderr, err)
						continue
					}
				}
				r.globals.define(variable.Name, res)
			}
		}

		// 4. Execute statements and output non-void expression results
		for _, stmt := range parsedProg.Stmts {
			exec, err := r.execute(stmt)
			if err != nil {
				fmt.Fprintln(options.Stderr, err)
				break
			}
			if exec.signal != signalNone {
				break
			}
			if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
				val, evalErr := r.eval(exprStmt.Expr)
				if evalErr == nil && val != nil {
					if _, isCall := exprStmt.Expr.(*ast.CallExpr); !isCall {
						fmt.Fprintln(options.Stdout, display(val))
					} else {
						t := info.TypeOf(exprStmt.Expr)
						if t.Kind != ast.TypeVoid && t.Kind != ast.TypeInvalid {
							fmt.Fprintln(options.Stdout, display(val))
						}
					}
				}
			}
		}
	}

	return scanner.Err()
}

func preloadPrograms(r *runner, initialPrograms []*ast.Program, initialInfos []*checker.Info) []ast.Decl {
	var accumulated []ast.Decl
	for i, program := range initialPrograms {
		info := initialInfos[i]
		r.program = program
		r.info = info

		for _, declaration := range program.Decls {
			decl := unwrapDeclNode(declaration)
			accumulated = append(accumulated, declaration)
			if function, ok := decl.(*ast.FuncDecl); ok {
				r.funcs[function.Name] = function
			}
		}

		for _, declaration := range program.Decls {
			decl := unwrapDeclNode(declaration)
			variable, ok := decl.(*ast.VarDecl)
			if !ok {
				continue
			}
			variableType, _ := info.GlobalType(variable.Name)
			res := r.zero(variableType)
			if variable.Init != nil {
				var err error
				res, err = r.eval(variable.Init)
				if err != nil {
					fmt.Fprintln(r.options.Stderr, err)
				}
			}
			r.globals.define(variable.Name, res)
		}

		if _, err := r.executeAll(program.Stmts); err != nil {
			fmt.Fprintln(r.options.Stderr, err)
		}
	}
	return accumulated
}

func unwrapDeclNode(declaration ast.Decl) ast.Decl {
	if exp, ok := declaration.(*ast.ExportStmt); ok {
		if innerDecl, ok := exp.Target.(ast.Decl); ok {
			return innerDecl
		}
	}
	return declaration
}
