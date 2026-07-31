/*
===============================================================================
QUEST STAGE 1: THE LEXICAL CONDUIT (Lexer)
===============================================================================
Overview:
  Transform raw source bytes into lexical tokens. In this stage, you will implement
  character classification helpers and keyword matching for variables and functions.

Tasks in this file:
  - TASK [LEX-01]: Implement character classification helpers:
                     - isDigit(c byte) bool
                     - isIdentStart(c byte) bool
  - TASK [LEX-02]: Implement keyword lookup in identifier() for reserved words
                   ('var', 'func', 'if', 'else').

Commands:
  - Run tests:  go test ./internal/lexer
  - Inspect:    ./zing tokens examples/01-tokens.zing
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

// Lexer converts Zing source bytes into a sequence of Tokens.
type Lexer struct {
	filename       string
	src            []byte
	start, current int
	line, column   int
	startPos       source.Position
	tokens         []token.Token
	diags          []diagnostic.Diagnostic
}

// Lex tokenizes input source code and appends an EOF token.
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

func (l *Lexer) pos() source.Position {
	return source.Position{Offset: l.current, Line: l.line, Column: l.column}
}
func (l *Lexer) atEnd() bool { return l.current >= len(l.src) }
func (l *Lexer) peek() byte {
	if l.atEnd() {
		return 0
	}
	return l.src[l.current]
}
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
func (l *Lexer) match(want byte) bool {
	if l.atEnd() || l.peek() != want {
		return false
	}
	l.advance()
	return true
}
func (l *Lexer) span() source.Span {
	return source.Span{Filename: l.filename, Start: l.startPos, End: l.pos()}
}
func (l *Lexer) add(k token.Kind, lit any) {
	l.tokens = append(l.tokens, token.Token{Kind: k, Lexeme: string(l.src[l.start:l.current]), Literal: lit, Span: l.span()})
}
func (l *Lexer) error(msg string) {
	l.diags = append(l.diags, diagnostic.Diagnostic{Span: l.span(), Phase: "lexer", Message: msg})
}

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

func (l *Lexer) identifier() {
	for isIdentPart(l.peek()) {
		l.advance()
	}
	raw := string(l.src[l.start:l.current])

	// TASK [LEX-02]: Implement keyword matching for reserved words ('var', 'func', 'if', 'else', 'true', 'false')!
	// Look up 'raw' in token.Keywords. If present, add the corresponding token kind; otherwise add token.Ident.
	// See HINT [LEX-02-HINT] at the bottom of this file for details.
	if k, ok := token.Keywords[raw]; ok {
		switch k {
		case token.True:
			l.add(k, true)
		case token.False:
			l.add(k, false)
		default:
			l.add(k, nil)
		}
	} else {
		l.add(token.Ident, raw)
	}
}

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

// TASK [LEX-01]: Implement character classification helpers:
//   - isDigit returns true if character 'c' is between '0' and '9'.
//   - isIdentStart returns true if 'c' can start an identifier ('_', 'a'-'z', 'A'-'Z').
// See HINT [LEX-01-HINT] at the bottom of this file for details.

func isDigit(c byte) bool {
	// TODO: implement isDigit helper
	return false
}

func isIdentStart(c byte) bool {
	// TODO: implement isIdentStart helper
	return false
}

func isIdentPart(c byte) bool {
	return isIdentStart(c) || isDigit(c)
}

/*
===============================================================================
QUEST HINTS & SOLUTIONS
===============================================================================
HINT [LEX-01-HINT]:
  Implement isDigit and isIdentStart:
    func isDigit(c byte) bool {
        return c >= '0' && c <= '9'
    }

    func isIdentStart(c byte) bool {
        return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
    }

HINT [LEX-02-HINT]:
  Check token.Keywords for 'raw':
    if k, ok := token.Keywords[raw]; ok {
        switch k {
        case token.True:
            l.add(k, true)
        case token.False:
            l.add(k, false)
        default:
            l.add(k, nil)
        }
    } else {
        l.add(token.Ident, raw)
    }
===============================================================================
*/
