package lexer

import (
	"strconv"
	"strings"
	"testing"

	"github.com/kc-clintone/compilers/internal/token"
)

func TestLexTokensAndKeywords(t *testing.T) {
	source := `(){}[],;:.=+-*/%! == != < <= > >= && ||
var type struct func if else switch case default for return break continue
true false map make int char string bool name 123 "a\n" '\t'`
	want := []token.Kind{
		token.LParen, token.RParen, token.LBrace, token.RBrace, token.LBracket, token.RBracket,
		token.Comma, token.Semicolon, token.Colon, token.Dot, token.Assign, token.Plus, token.Minus,
		token.Star, token.Slash, token.Percent, token.Bang, token.Equal, token.NotEqual, token.Less,
		token.LessEqual, token.Greater, token.GreaterEqual, token.And, token.Or, token.Var, token.Type,
		token.Struct, token.Func, token.If, token.Else, token.Switch, token.Case, token.Default,
		token.For, token.Return, token.Break, token.Continue, token.True, token.False, token.Map,
		token.Make, token.IntType, token.CharType, token.StringType, token.BoolType, token.Ident,
		token.Integer, token.String, token.Char, token.EOF,
	}

	tokens, diagnostics := Lex("tokens.zing", []byte(source))
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if len(tokens) != len(want) {
		t.Fatalf("got %d tokens, want %d", len(tokens), len(want))
	}
	for i, kind := range want {
		if tokens[i].Kind != kind {
			t.Fatalf("token %d: got %s, want %s", i, tokens[i].Kind, kind)
		}
	}
	if tokens[len(tokens)-4].Literal != 123 || tokens[len(tokens)-3].Literal != "a\n" || tokens[len(tokens)-2].Literal != byte('\t') {
		t.Fatalf("decoded literals are wrong: %#v", tokens[len(tokens)-4:])
	}
}

func TestLexCommentsEscapesAndPositions(t *testing.T) {
	source := "// one\n\tvar name string = \"\\r\\t\\\\\\\"\\'\"; /* two\nlines */\nname"
	tokens, diagnostics := Lex("positions.zing", []byte(source))
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if tokens[0].Span.Start.Line != 2 || tokens[0].Span.Start.Column != 2 {
		t.Fatalf("var starts at %v", tokens[0].Span.Start)
	}
	last := tokens[len(tokens)-2]
	if last.Kind != token.Ident || last.Span.Start.Line != 4 || last.Span.Start.Column != 1 {
		t.Fatalf("last identifier has span %v", last.Span)
	}
}

func TestLexDiagnosticsAndRecovery(t *testing.T) {
	overflow := strings.Repeat("9", strconv.IntSize/3+10)
	tests := []struct {
		name, source, message string
	}{
		{"invalid byte", "\xff var x int;", "invalid byte 0xff"},
		{"single and", "&", "expected '&' after '&'"},
		{"single or", "|", "expected '|' after '|'"},
		{"overflow", overflow, "integer literal overflows int"},
		{"invalid escape", `"bad\q" var x int;`, "invalid escape sequence"},
		{"unterminated string", `"bad`, "unterminated literal"},
		{"newline string", "\"bad\nvar x int;", "unterminated literal"},
		{"empty char", `''`, "character literal must contain exactly one byte"},
		{"long char", `'ab'`, "character literal must contain exactly one byte"},
		{"unterminated comment", `/* bad`, "unterminated block comment"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokens, diagnostics := Lex("bad.zing", []byte(test.source))
			if len(diagnostics) == 0 || diagnostics[0].Message != test.message {
				t.Fatalf("diagnostics = %v, want %q", diagnostics, test.message)
			}
			if tokens[len(tokens)-1].Kind != token.EOF {
				t.Fatalf("last token is %s, want EOF", tokens[len(tokens)-1].Kind)
			}
		})
	}
}

func FuzzLex(f *testing.F) {
	for _, seed := range [][]byte{nil, []byte("var x int = 1;"), []byte("\xff\x00\n"), []byte(`"\\q`)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, source []byte) {
		tokens, _ := Lex("fuzz.zing", source)
		if len(tokens) == 0 || tokens[len(tokens)-1].Kind != token.EOF {
			t.Fatal("lexer did not emit EOF")
		}
	})
}
