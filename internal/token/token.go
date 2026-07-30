// Package token defines the lexical vocabulary of Zing.
package token

import "github.com/kc-clintone/compilers/internal/source"

// Kind identifies one lexical token category.
type Kind string

// Zing token kinds.
const (
	EOF          Kind = "EOF"
	Invalid      Kind = "INVALID"
	Ident        Kind = "IDENT"
	Integer      Kind = "INT"
	String       Kind = "STRING"
	Char         Kind = "CHAR"
	LParen       Kind = "("
	RParen       Kind = ")"
	LBrace       Kind = "{"
	RBrace       Kind = "}"
	LBracket     Kind = "["
	RBracket     Kind = "]"
	Comma        Kind = ","
	Semicolon    Kind = ";"
	Colon        Kind = ":"
	Dot          Kind = "."
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
	Var          Kind = "var"
	Type         Kind = "type"
	Struct       Kind = "struct"
	Func         Kind = "func"
	If           Kind = "if"
	Else         Kind = "else"
	Switch       Kind = "switch"
	Case         Kind = "case"
	Default      Kind = "default"
	For          Kind = "for"
	Return       Kind = "return"
	Break        Kind = "break"
	Continue     Kind = "continue"
	True         Kind = "true"
	False        Kind = "false"
	Map          Kind = "map"
	Make         Kind = "make"
	IntType      Kind = "int"
	CharType     Kind = "char"
	StringType   Kind = "string"
	BoolType     Kind = "bool"
)

// Token contains the original lexeme, its decoded literal value, and span.
type Token struct {
	Kind    Kind
	Lexeme  string
	Literal any
	Span    source.Span
}

// Keywords maps reserved source words to their token kinds.
var Keywords = map[string]Kind{
	"var": Var, "type": Type, "struct": Struct, "func": Func, "if": If, "else": Else, "switch": Switch,
	"case": Case, "default": Default, "for": For, "return": Return, "break": Break, "continue": Continue,
	"true": True, "false": False, "map": Map, "make": Make, "int": IntType, "char": CharType, "string": StringType, "bool": BoolType,
}
