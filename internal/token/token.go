package token
// Package token defines the token types and values used by the lexer.
//
// Participants will add token definitions, helper functions, and any needed
// constants as part of the workshop exercises.
package token

// TokenType identifies the kind of lexical token produced by the lexer.
type TokenType string

// Token represents a single unit of source input.
type Token struct {
	Type    TokenType
	Literal string
}

// TODO: Define token types for keywords, operators, delimiters, and literals.
// TODO: Add helper functions for token classification and reporting.
