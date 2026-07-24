package ast
// Package ast defines the abstract syntax tree nodes for the tiny language.
//
// Participants will add node types and helper methods as they implement the
// parser and the rest of the compiler pipeline.
package ast

// Node is the common interface for all AST nodes.
type Node interface {
	Accept(visitor Visitor) interface{}
}

// Statement is a node that can appear in a program body.
type Statement interface {
	Node
	statementNode()
}

// Expression is a node that evaluates to a value.
type Expression interface {
	Node
	expressionNode()
}

// Program represents a complete source program.
type Program struct {
	Statements []Statement
}

// Accept implements the visitor pattern for the program node.
func (p *Program) Accept(visitor Visitor) interface{} {
	panic("not implemented")
}

// Visitor defines traversal behavior for AST nodes.
type Visitor interface {
	VisitProgram(*Program) interface{}
}

// TODO: Add concrete AST node types for literals, identifiers, expressions, and statements.
// TODO: Implement visitor methods and helper interfaces for traversal.
