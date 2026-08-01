/*
===============================================================================
QUEST STAGE 1: THE LEXICAL CONDUIT (Lexer) [COMPLETED]
===============================================================================
Overview:
  The lexer is the first compiler stage. It walks source bytes from left to right
  and groups them into tokens that the parser can understand. Each token records
  its category, original spelling, optional literal value, and source span.

  This QUEST teaches the lexer to recognize numbers and names, distinguish names
  from reserved words, and preserve boolean literal values for later stages.

Tasks in this file:
  - [COMPLETED] TASK [LEX-01]: Classify digits and identifier-start characters.
  - [COMPLETED] TASK [LEX-02]: Emit identifiers, reserved keywords, and boolean literals.

Commands:
  - Run tests:  go test ./internal/lexer
  - Inspect:    ./nuru-interpreter tokens examples/01-tokens.nuru
  - Skip stage: ./savepoint.sh 1
  - Reset stage: ./savepoint.sh 0

===============================================================================
*/

package lexer

import (
	"fmt"
	"strconv"

	"github.com/kc-clintone/compilers/internal/diagnostic"
	"github.com/kc-clintone/compilers/internal/source"
	"github.com/kc-clintone/compilers/internal/token"
)

// Lexer tracks the source slice, current token boundaries, source position,
// emitted tokens, and diagnostics while scanning a Nuru file.
type Lexer struct {
	filename       string
	src            []byte
	start, current int
	line, column   int
	startPos       source.Position
	tokens         []token.Token
	diags          []diagnostic.Diagnostic
}

// Lex tokenizes input source code, appends an EOF token, and returns any lexer
// diagnostics alongside the token stream.
func Lex(filename string, src []byte) ([]token.Token, []diagnostic.Diagnostic) {
	l := &Lexer{filename: filename, src: src, line: 1, column: 1}

	for !l.atEnd() {
		l.start = l.current
		l.startPos = l.pos()
		l.scanToken()
	}

	p := l.pos()
	l.tokens = append(l.tokens, token.Token{Kind: token.EOF, Span: source.Span{Filename: filename, Start: p, End: p}})
	return l.tokens, l.diags
}

// scanToken consumes one token, comment, or whitespace sequence beginning at
// the current byte. Lex sets the token start and starting position beforehand.
func (l *Lexer) scanToken() {
	c := l.advance()

	switch c {
	case ' ', '\r', '\t', '\n':
		return
	case '(':
		l.add(token.LParen, nil)
	case ')':
		l.add(token.RParen, nil)
	case '{':
		l.add(token.LBrace, nil)
	case '}':
		l.add(token.RBrace, nil)
	case ',':
		l.add(token.Comma, nil)
	case ';':
		l.add(token.Semicolon, nil)
	case '+':
		l.add(token.Plus, nil)
	case '-':
		l.add(token.Minus, nil)
	case '*':
		l.add(token.Star, nil)
	case '%':
		l.add(token.Percent, nil)
	case '=':
		if l.match('=') {
			l.add(token.Equal, nil)
		} else {
			l.add(token.Assign, nil)
		}
	case '!':
		if l.match('=') {
			l.add(token.NotEqual, nil)
		} else {
			l.add(token.Bang, nil)
		}
	case '<':
		if l.match('=') {
			l.add(token.LessEqual, nil)
		} else {
			l.add(token.Less, nil)
		}
	case '>':
		if l.match('=') {
			l.add(token.GreaterEqual, nil)
		} else {
			l.add(token.Greater, nil)
		}
	case '&':
		if l.match('&') {
			l.add(token.And, nil)
		} else {
			l.error("expected '&' after '&'")
		}
	case '|':
		if l.match('|') {
			l.add(token.Or, nil)
		} else {
			l.error("expected '|' after '|'")
		}
	case '/':
		if l.match('/') {
			for !l.atEnd() && l.peek() != '\n' {
				l.advance()
			}
		} else {
			l.add(token.Slash, nil)
		}
	case '"':
		l.stringLiteral()
	default:
		if isDigit(c) {
			l.number()
		} else if isIdentStart(c) {
			l.identifier()
		} else {
			l.error(fmt.Sprintf("invalid character 0x%02x", c))
		}
	}
}

// pos returns the current half-open source position. It points immediately
// after the most recently consumed byte.
func (l *Lexer) pos() source.Position {
	return source.Position{Offset: l.current, Line: l.line, Column: l.column}
}

// atEnd reports whether every source byte has been consumed.
func (l *Lexer) atEnd() bool { return l.current >= len(l.src) }

// peek returns the next byte without consuming it, or zero at end of input.
func (l *Lexer) peek() byte {
	if l.atEnd() {
		return 0
	}
	return l.src[l.current]
}

// advance consumes and returns the next byte while updating the byte offset,
// line, and one-based column. Callers must check atEnd before advancing.
func (l *Lexer) advance() byte {
	c := l.src[l.current]
	l.current++
	if c == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return c
}

// match consumes want when it is the next byte and reports whether it matched.
// A failed match leaves the lexer position unchanged.
func (l *Lexer) match(want byte) bool {
	if l.atEnd() || l.peek() != want {
		return false
	}
	l.advance()
	return true
}

// span returns the half-open source range for the token currently being scanned.
func (l *Lexer) span() source.Span {
	return source.Span{Filename: l.filename, Start: l.startPos, End: l.pos()}
}

// add emits a token spanning the current token boundaries. The lexeme is copied
// from the source; lit stores a decoded value such as int, string, or bool.
func (l *Lexer) add(k token.Kind, lit any) {
	l.tokens = append(l.tokens, token.Token{Kind: k, Lexeme: string(l.src[l.start:l.current]), Literal: lit, Span: l.span()})
}

// error records a lexer diagnostic over the token currently being scanned.
func (l *Lexer) error(msg string) {
	l.diags = append(l.diags, diagnostic.Diagnostic{Span: l.span(), Phase: "lexer", Message: msg})
}

// number consumes the remaining decimal digits, converts the complete lexeme
// to an int, and emits token.Integer or an overflow diagnostic.
func (l *Lexer) number() {
	for isDigit(l.peek()) {
		l.advance()
	}
	raw := string(l.src[l.start:l.current])
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		l.error("integer literal overflow")
		return
	}
	l.add(token.Integer, int(n))
}

// identifier consumes the remainder of a name. It emits the mapped keyword
// kind when the spelling is reserved, preserving bool values for true and
// false; otherwise it emits token.Ident with the name as its literal value.
func (l *Lexer) identifier() {
	for isIdentPart(l.peek()) {
		l.advance()
	}
	raw := string(l.src[l.start:l.current])

	if kind, ok := token.Keywords[raw]; ok {
		switch kind {
		case token.True:
			l.add(kind, true)
		case token.False:
			l.add(kind, false)
		default:
			l.add(kind, nil)
		}
		return
	}
	l.add(token.Ident, raw)
}

// stringLiteral consumes through the closing quote, decodes the supported
// escape sequences, and emits token.String or a diagnostic for invalid input.
func (l *Lexer) stringLiteral() {
	var out []byte
	for !l.atEnd() && l.peek() != '"' {
		c := l.advance()
		if c == '\n' {
			l.error("unterminated string literal")
			return
		}
		if c == '\\' {
			if l.atEnd() {
				break
			}
			e := l.advance()
			if e == 'n' {
				out = append(out, '\n')
			} else if e == 't' {
				out = append(out, '\t')
			} else if e == '"' || e == '\\' {
				out = append(out, e)
			} else {
				l.error("invalid escape sequence")
				return
			}
		} else {
			out = append(out, c)
		}
	}
	if l.atEnd() {
		l.error("unterminated string literal")
		return
	}
	l.advance() // consume closing quote
	l.add(token.String, string(out))
}

// isDigit reports whether c is an ASCII decimal digit.
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// isIdentStart reports whether c can begin a Nuru identifier: an ASCII letter
// or underscore. Digits are accepted only after the first character.
func isIdentStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isIdentPart reports whether c may follow the first identifier character.
func isIdentPart(c byte) bool {
	return isIdentStart(c) || isDigit(c)
}

/*
===============================================================================
QUEST HINTS
===============================================================================
HINT [LEX-01-HINT]:
  1. For isDigit, check that c is at least '0' and at most '9'.
  2. For isIdentStart, accept underscore first.
  3. Also accept c when it falls in either ASCII letter range.
  4. Return false for every other byte.

HINT [LEX-02-HINT]:
  1. Look up raw in token.Keywords.
  2. If it is not present, emit token.Ident and preserve raw as the literal.
  3. If it maps to token.True or token.False, emit the mapped kind with the
     corresponding Go bool literal.
  4. Emit every other keyword with a nil literal.
===============================================================================
*/
