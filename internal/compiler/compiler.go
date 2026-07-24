package compiler
// Package compiler translates the AST into simple pseudo-assembly.
//
// The compiler is intentionally left as a placeholder for the workshop.
package compiler

import "github.com/kc-clintone/compiler-workshop/internal/ast"

// Compiler translates AST nodes into instructions.
type Compiler struct{}

// New creates a new compiler.
func New() *Compiler {
	panic("not implemented")
}

// Compile transforms a program into pseudo-assembly instructions.
func (c *Compiler) Compile(program *ast.Program) []string {
	panic("not implemented")
}

// TODO: Implement instruction emission for literals, expressions, and statements.
// TODO: Define pseudo-assembly formats for workshop participants to follow.
