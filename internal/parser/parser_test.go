package parser

import (
	"testing"

	"github.com/kc-clintone/compilers/internal/ast"
)

func TestParserBasic(t *testing.T) {
	input := `print(1 + 2 * 3)`
	prog, diags := Parse("test.zing", []byte(input))
	if len(diags) > 0 {
		t.Fatalf("unexpected parser diagnostics: %v", diags)
	}

	if len(prog.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Stmts))
	}

	printStmt, ok := prog.Stmts[0].(*ast.PrintStmt)
	if !ok {
		t.Fatalf("expected PrintStmt, got %T", prog.Stmts[0])
	}

	if len(printStmt.Args) != 1 {
		t.Fatalf("expected 1 print argument, got %d", len(printStmt.Args))
	}
}
