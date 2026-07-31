/*
===============================================================================
QUEST STAGE 1: THE LEXICAL CONDUIT (Token Vocabulary)
===============================================================================
Overview:
  Define the lexical vocabulary tokens for variable and function declarations.

Tasks in this file:
  - TASK [LEX-03]: Define Var ("var") and Func ("func") token kinds and add them
                   to the Keywords map.

Commands:
  - Run tests:  go test ./internal/token ./internal/lexer
  - Skip stage: ./savepoint.sh 1
  - Reset stage: ./savepoint.sh 0
===============================================================================
*/

package token

import (
	"fmt"

	"github.com/kc-clintone/compilers/internal/source"
)

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

	// TASK [LEX-03]: Define Var ("var") and Func ("func") token kinds here for Stage 1!
	// See HINT [LEX-03-HINT] at the bottom of this file for details.
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

func (t Token) String() string {
	if t.Literal != nil {
		return fmt.Sprintf("%-12s %-10s %-12v %s", t.Span.Start, t.Kind, t.Literal, t.Lexeme)
	}
	return fmt.Sprintf("%-12s %-10s %-12s %s", t.Span.Start, t.Kind, "", t.Lexeme)
}

// PrintTokens displays a clean tabular output of scanned tokens.
func PrintTokens(tokens []Token) {
	fmt.Printf("%-12s %-10s %-12s %s\n", "POSITION", "KIND", "LITERAL", "LEXEME")
	fmt.Println("---------------------------------------------------------")
	for _, tok := range tokens {
		fmt.Println(tok)
	}
}

// Keywords maps reserved keywords to token kinds.
// TASK [LEX-03]: Add "var": Var and "func": Func to the Keywords map below for Stage 1!
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

/*
===============================================================================
QUEST HINTS & SOLUTIONS
===============================================================================
HINT [LEX-03-HINT]:
  Define the Var and Func token constants:
    Var  Kind = "var"
    Func Kind = "func"

  Then add them to the Keywords map:
    "var":  Var,
    "func": Func,
===============================================================================
*/
