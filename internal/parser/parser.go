package parser
// Package parser builds an AST from a stream of tokens.
//
// The parser is intentionally left as a placeholder so participants can
// implement expression parsing and statement parsing during the workshop.
package parser

import (
	"github.com/kc-clintone/compiler-workshop/internal/ast"
	"github.com/kc-clintone/compiler-workshop/internal/token"
)

// Parser converts tokens into an abstract syntax tree.
type Parser struct {
	currToken token.Token
	nextToken token.Token
}

// New creates a parser for the supplied token stream.
func New(tokens []token.Token) *Parser {
	panic("not implemented")
}

// ParseProgram parses the complete program into an AST.
func (p *Parser) ParseProgram() *ast.Program {
	panic("not implemented")
}

// TODO: Implement parsing for expressions, statements, and precedence levels.
// TODO: Add error handling for invalid syntax.
