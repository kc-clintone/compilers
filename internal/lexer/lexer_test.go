package lexer

import (
	"testing"

	"github.com/kc-clintone/compilers/internal/token"
)

func TestLexerSymbols(t *testing.T) {
	input := `+ - * / ( ) { } = == != < > <= >= && ||`
	toks, diags := Lex("test.nuru", []byte(input))
	if len(diags) > 0 {
		t.Fatalf("unexpected lexer diagnostics: %v", diags)
	}

	expectedKinds := []token.Kind{
		token.Plus, token.Minus, token.Star, token.Slash,
		token.LParen, token.RParen, token.LBrace, token.RBrace,
		token.Assign, token.Equal, token.NotEqual,
		token.Less, token.Greater, token.LessEqual, token.GreaterEqual,
		token.And, token.Or, token.EOF,
	}

	if len(toks) != len(expectedKinds) {
		t.Fatalf("token count mismatch: got %d, want %d", len(toks), len(expectedKinds))
	}

	for i, want := range expectedKinds {
		if toks[i].Kind != want {
			t.Errorf("token %d mismatch: got %s, want %s", i, toks[i].Kind, want)
		}
	}
}
