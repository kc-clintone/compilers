package parser

import (
	"testing"

	"github.com/kc-clintone/compilers/internal/ast"
)

func TestParserScaffold(t *testing.T) {
	prog, diags := Parse("test.nuru", []byte(`"a" + "b" * "c"`))
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if prog == nil || len(prog.Stmts) != 1 {
		t.Fatalf("expected one parsed statement, got %#v", prog)
	}
	exprStmt, ok := prog.Stmts[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected expression statement, got %T", prog.Stmts[0])
	}
	add, ok := exprStmt.Expr.(*ast.BinaryExpr)
	if !ok || add.Op != "+" {
		t.Fatalf("expected additive root, got %#v", exprStmt.Expr)
	}
	if multiply, ok := add.Right.(*ast.BinaryExpr); !ok || multiply.Op != "*" {
		t.Fatalf("expected multiplication to bind tighter, got %#v", add.Right)
	}
}
