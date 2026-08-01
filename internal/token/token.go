/*
===============================================================================
QUEST STAGE 1: THE LEXICAL CONDUIT (Token Vocabulary)
===============================================================================
Overview:
  Token kinds are the shared vocabulary between the lexer and parser. The lexer
  emits them; the parser uses them to decide which grammar rule to apply.

  This part of the lexer QUEST adds the reserved spellings that introduce
  variable and function declarations.

Tasks in this file:
  - TASK [LEX-03]: Define and register the var and func keyword kinds.

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

	// TASK [LEX-03]: Define Var and Func kinds using their Nuru source spellings.
	// Then register both constants in Keywords below.
	// See HINT [LEX-03-HINT] at the bottom of this file for details.
)

// Token contains a scanned token's kind, literal representation, source text, and span.
type Token struct {
	Kind    Kind
	Lexeme  string
	Literal any
	Span    source.Span
}

// String formats a token as one row of the token-inspection table.
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

// Keywords maps reserved source spellings to the kinds emitted by identifier.
var Keywords = map[string]Kind{
	"if":       If,
	"else":     Else,
	"for":      For,
	"break":    Break,
	"continue": Continue,
	"true":     True,
	"false":    False,
	// TASK [LEX-03]: Register the "var" and "func" spellings here.
}

/*
===============================================================================
QUEST HINTS
===============================================================================
HINT [LEX-03-HINT]:
  1. Add two Kind constants beside the existing control-keyword constants.
  2. Give each constant the same lowercase spelling used in Nuru source.
  3. Add one Keywords entry from each source spelling to its new constant.
  4. Keep the constant names and map values linked rather than constructing
     token.Kind values at each parser call site.
===============================================================================
*/
