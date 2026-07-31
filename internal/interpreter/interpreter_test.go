package interpreter

import (
	"testing"

	"github.com/kc-clintone/compilers/internal/ast"
)

func TestInterpreterScaffold(t *testing.T) {
	prog := &ast.Program{}
	interp := New()
	diags := interp.Interpret(prog)
	if len(diags) > 0 {
		t.Fatalf("unexpected interpreter diagnostics: %v", diags)
	}
}
