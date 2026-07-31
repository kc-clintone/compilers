package interpreter

import (
	"testing"

	"github.com/kc-clintone/compilers/internal/parser"
)

func TestInterpreterBasic(t *testing.T) {
	input := `if (1 < 2) { print(42) }`
	prog, diags := parser.Parse("test.zing", []byte(input))
	if len(diags) > 0 {
		t.Fatalf("unexpected parser diagnostics: %v", diags)
	}

	interp := New()
	evalDiags := interp.Interpret(prog)
	if len(evalDiags) > 0 {
		t.Fatalf("unexpected interpreter diagnostics: %v", evalDiags)
	}
}
