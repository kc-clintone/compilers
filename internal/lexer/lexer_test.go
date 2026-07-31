package lexer

import (
	"testing"

	"github.com/kc-clintone/compilers/internal/token"
)

func TestLexerBasic(t *testing.T) {
	input := `1 + 2 * (3 - 4) == 5 && true`
	toks, diags := Lex("test.zing", []byte(input))
	if len(diags) > 0 {
		t.Fatalf("unexpected lexer diagnostics: %v", diags)
	}

	expectedKinds := []token.Kind{
		token.Integer, token.Plus, token.Integer, token.Star,
		token.LParen, token.Integer, token.Minus, token.Integer, token.RParen,
		token.Equal, token.Integer, token.And, token.True, token.EOF,
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
