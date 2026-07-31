// Package lexer converts Nuru source bytes into tokens.
package lexer

import (
	"fmt"
	"strconv"

	"github.com/kc-clintone/compilers/internal/diagnostic"
	"github.com/kc-clintone/compilers/internal/source"
	"github.com/kc-clintone/compilers/internal/token"
)

// Lexer converts one Nuru source file into a token stream.
type Lexer struct {
	filename       string
	src            []byte
	start, current int
	line, column   int
	startPos       source.Position
	tokens         []token.Token
	diags          []diagnostic.Diagnostic
}

// Lex tokenizes src and always appends an EOF token, even when diagnostics are
// produced.
func Lex(filename string, src []byte) ([]token.Token, []diagnostic.Diagnostic) {
	s := &Lexer{filename: filename, src: src, line: 1, column: 1}

	for !s.atEnd() {
		s.start = s.current
		s.startPos = s.pos()
		s.scanToken()
	}

	p := s.pos()

	s.tokens = append(s.tokens, token.Token{Kind: token.EOF, Span: source.Span{Filename: filename, Start: p, End: p}})
	return s.tokens, s.diags
}

func (s *Lexer) pos() source.Position {
	return source.Position{Offset: s.current, Line: s.line, Column: s.column}
}
func (s *Lexer) atEnd() bool { return s.current >= len(s.src) }
func (s *Lexer) peek() byte {
	if s.atEnd() {
		return 0
	}

	return s.src[s.current]
}
func (s *Lexer) peekNext() byte {
	if s.current+1 >= len(s.src) {
		return 0
	}

	return s.src[s.current+1]
}
func (s *Lexer) advance() byte {
	c := s.src[s.current]

	s.current++
	if c == '\n' {
		s.line++
		s.column = 1
	} else {
		s.column++
	}

	return c
}
func (s *Lexer) match(want byte) bool {
	if s.atEnd() || s.peek() != want {
		return false
	}

	s.advance()
	return true
}
func (s *Lexer) span() source.Span {
	return source.Span{Filename: s.filename, Start: s.startPos, End: s.pos()}
}
func (s *Lexer) add(k token.Kind, lit any) {
	s.tokens = append(s.tokens, token.Token{Kind: k, Lexeme: string(s.src[s.start:s.current]), Literal: lit, Span: s.span()})
}
func (s *Lexer) error(msg string) {
	s.diags = append(s.diags, diagnostic.Diagnostic{Span: s.span(), Phase: "lexer", Message: msg})
}

func (s *Lexer) scanToken() {
	c := s.advance()

	switch c {
	case ' ', '\r', '\t', '\n':
		return
	case '(':
		s.add(token.LParen, nil)
	case ')':
		s.add(token.RParen, nil)
	case '{':
		s.add(token.LBrace, nil)
	case '}':
		s.add(token.RBrace, nil)
	case '[':
		s.add(token.LBracket, nil)
	case ']':
		s.add(token.RBracket, nil)
	case ',':
		s.add(token.Comma, nil)
	case ';':
		s.add(token.Semicolon, nil)
	case ':':
		s.add(token.Colon, nil)
	case '.':
		s.add(token.Dot, nil)
	case '+':
		s.add(token.Plus, nil)
	case '-':
		s.add(token.Minus, nil)
	case '*':
		s.add(token.Star, nil)
	case '%':
		s.add(token.Percent, nil)
	case '=':
		if s.match('=') {
			s.add(token.Equal, nil)
		} else {
			s.add(token.Assign, nil)
		}
	case '!':
		if s.match('=') {
			s.add(token.NotEqual, nil)
		} else {
			s.add(token.Bang, nil)
		}
	case '<':
		if s.match('=') {
			s.add(token.LessEqual, nil)
		} else {
			s.add(token.Less, nil)
		}
	case '>':
		if s.match('=') {
			s.add(token.GreaterEqual, nil)
		} else {
			s.add(token.Greater, nil)
		}
	case '&':
		if s.match('&') {
			s.add(token.And, nil)
		} else {
			s.error("expected '&' after '&'")
		}
	case '|':
		if s.match('|') {
			s.add(token.Or, nil)
		} else {
			s.error("expected '|' after '|'")
		}
	case '/':
		if s.match('/') {
			for !s.atEnd() && s.peek() != '\n' {
				s.advance()
			}
		} else if s.match('*') {
			s.blockComment()
		} else {
			s.add(token.Slash, nil)
		}
	case '"':
		s.quoted('"', token.String)
	case '\'':
		s.quoted('\'', token.Char)
	default:
		if isDigit(c) {
			s.number()
		} else if isIdentStart(c) {
			s.identifier()
		} else {
			s.error(fmt.Sprintf("invalid byte 0x%02x", c))
		}
	}
}

func (s *Lexer) blockComment() {
	for !s.atEnd() {
		if s.peek() == '*' && s.peekNext() == '/' {
			s.advance()
			s.advance()
			return
		}

		s.advance()
	}

	s.error("unterminated block comment")
}
func (s *Lexer) number() {
	for isDigit(s.peek()) {
		s.advance()
	}

	raw := string(s.src[s.start:s.current])
	n, err := strconv.ParseInt(raw, 10, 0)
	if err != nil {
		s.error("integer literal overflows int")
		return
	}

	s.add(token.Integer, int(n))
}
func (s *Lexer) identifier() {
	for isIdentPart(s.peek()) {
		s.advance()
	}

	raw := string(s.src[s.start:s.current])

	if k, ok := token.Keywords[raw]; ok {
		switch k {
		case token.True:
			s.add(k, true)
		case token.False:
			s.add(k, false)
		default:
			s.add(k, nil)
		}
	} else {
		s.add(token.Ident, raw)
	}
}
func (s *Lexer) quoted(quote byte, k token.Kind) {
	var out []byte

	for !s.atEnd() && s.peek() != quote {
		c := s.advance()

		if c == '\n' {
			s.error("unterminated literal")
			return
		}

		if c == '\\' {
			if s.atEnd() {
				break
			}

			e := s.advance()
			m := map[byte]byte{'n': '\n', 'r': '\r', 't': '\t', '\\': '\\', '"': '"', '\'': '\''}
			v, ok := m[e]

			if !ok {
				s.error("invalid escape sequence")
				s.recoverQuoted(quote)
				return
			}

			out = append(out, v)
		} else {
			out = append(out, c)
		}
	}

	if s.atEnd() {
		s.error("unterminated literal")
		return
	}

	s.advance()
	if k == token.Char {
		if len(out) != 1 {
			s.error("character literal must contain exactly one byte")
			return
		}

		s.add(k, out[0])
	} else {
		s.add(k, string(out))
	}
}
func (s *Lexer) recoverQuoted(quote byte) {
	for !s.atEnd() && s.peek() != '\n' {
		if s.advance() == quote {
			return
		}
	}
}
func isDigit(c byte) bool      { return c >= '0' && c <= '9' }
func isIdentStart(c byte) bool { return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }
func isIdentPart(c byte) bool  { return isIdentStart(c) || isDigit(c) }
