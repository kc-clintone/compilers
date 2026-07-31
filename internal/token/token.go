package token

import "github.com/kc-clintone/compilers/internal/source"

// Kind identifies a category of lexical token.
type Kind string

const (
	EOF          Kind = "EOF"
	Invalid      Kind = "INVALID"
	Ident        Kind = "IDENT"
	Integer      Kind = "INT"
	String       Kind = "STRING"
	LParen       Kind = "("
	RParen       Kind = ")"
	LBrace       Kind = "{"
	RBrace       Kind = "}"
	Comma        Kind = ","
	Semicolon    Kind = ";"
	Assign       Kind = "="
	Plus         Kind = "+"
	Minus        Kind = "-"
	Star         Kind = "*"
	Slash        Kind = "/"
	Percent      Kind = "%"
	Bang         Kind = "!"
	Equal        Kind = "=="
	NotEqual     Kind = "!="
	Less         Kind = "<"
	LessEqual    Kind = "<="
	Greater      Kind = ">"
	GreaterEqual Kind = ">="
	And          Kind = "&&"
	Or           Kind = "||"

	// Control keywords
	If       Kind = "if"
	Else     Kind = "else"
	For      Kind = "for"
	Break    Kind = "break"
	Continue Kind = "continue"
	True     Kind = "true"
	False    Kind = "false"

	// TODO: Lexer - Define Var ("var") and Func ("func") token kinds here for Stage 1!
	// Example:
	// Var  Kind = "var"
	// Func Kind = "func"
)

// Token contains a scanned token's kind, literal representation, source text, and span.
type Token struct {
	Kind    Kind
	Lexeme  string
	Literal any
	Span    source.Span
}

// Keywords maps reserved keywords to token kinds.
// TODO: Lexer - Add "var" and "func" keywords to this map for Stage 1!
var Keywords = map[string]Kind{
	"if":       If,
	"else":     Else,
	"for":      For,
	"break":    Break,
	"continue": Continue,
	"true":     True,
	"false":    False,
	// "var":  Var,
	// "func": Func,
}
