package interpreter
// Package interpreter evaluates programs represented by the AST.
//
// The interpreter is intentionally left as a placeholder for the workshop.
package interpreter

import (
	"github.com/kc-clintone/compiler-workshop/internal/ast"
	"github.com/kc-clintone/compiler-workshop/internal/object"
)

// Interpreter evaluates AST nodes and produces runtime objects.
type Interpreter struct{}

// New creates a new interpreter.
func New() *Interpreter {
	panic("not implemented")
}

// Interpret evaluates a program.
func (i *Interpreter) Interpret(program *ast.Program) object.Object {
	panic("not implemented")
}

// TODO: Implement evaluation of literals, expressions, and statements.
// TODO: Add support for environment and runtime values.
