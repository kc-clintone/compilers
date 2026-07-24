package lexer
// Package lexer converts source text into a stream of tokens.
//
// The lexer is intentionally left as a placeholder so workshop participants can
// implement scanning logic and tokenization during the exercises.
package lexer

import "github.com/kc-clintone/compiler-workshop/internal/token"

// Lexer scans source input and produces tokens.
type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
}

// New creates a new lexer for the supplied input.
func New(input string) *Lexer {
	panic("not implemented")
}

// NextToken reads and returns the next token from the input.
func (l *Lexer) NextToken() token.Token {
	panic("not implemented")
}

// TODO: Implement character scanning and token generation.
// TODO: Support identifiers, integers, operators, and delimiters.
