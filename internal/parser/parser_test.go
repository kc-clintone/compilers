package parser

import (
	"testing"

	"github.com/kc-clintone/compilers/internal/ast"
)

func TestControlFlowBraceIsNotNamedComposite(t *testing.T) {
	source := []byte(`var value int = 1;
switch value { case 1: print("one"); default: print("other"); }
for value > 0 { value = value - 1; }
`)
	program, diagnostics := Parse("control.nuru", source)

	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}

	if got := len(program.Stmts); got != 2 {
		t.Fatalf("got %d statements, want 2", got)
	}
}

func TestMisplacedDeclarationProducesDiagnostic(t *testing.T) {
	_, diagnostics := Parse("ordering.nuru", []byte("print(1); var later int = 2;"))

	if len(diagnostics) != 1 || diagnostics[0].Message != "declarations must precede top-level statements" {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
}

func TestExpressionPrecedenceAndAssociativity(t *testing.T) {
	program, diagnostics := Parse("precedence.nuru", []byte(`print(1 + 2 * 3 == 7 || false && true);`))

	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}

	call := program.Stmts[0].(*ast.ExprStmt).Expr.(*ast.CallExpr)

	if got, want := ast.DebugExpr(call.Args[0]), "(((1 + (2 * 3)) == 7) || (false && true))"; got != want {
		t.Fatalf("expression = %s, want %s", got, want)
	}

	program, diagnostics = Parse("associativity.nuru", []byte(`print(10 - 3 - 2);`))
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}

	call = program.Stmts[0].(*ast.ExprStmt).Expr.(*ast.CallExpr)
	if got, want := ast.DebugExpr(call.Args[0]), "((10 - 3) - 2)"; got != want {
		t.Fatalf("expression = %s, want %s", got, want)
	}
}

func TestPostfixChainsAndCompositeForms(t *testing.T) {
	source := `
type Item struct { name string; }
var items []Item = []Item{Item{name: "one"},};
var counts map[string]int = map[string]int{"one": 1,};
var empty []int = make([]int, 0);
print(items[0:1][0].name);
`
	program, diagnostics := Parse("postfix.nuru", []byte(source))

	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}

	call := program.Stmts[0].(*ast.ExprStmt).Expr.(*ast.CallExpr)

	if got, want := ast.DebugExpr(call.Args[0]), "items[0:1][0].name"; got != want {
		t.Fatalf("postfix expression = %s, want %s", got, want)
	}
}

func TestNodeIDsAreAssignedDeterministically(t *testing.T) {
	parse := func() *ast.Program {
		program, diagnostics := Parse("ids.nuru", []byte(`var x int = 1; print(x + 2);`))

		if len(diagnostics) != 0 {
			t.Fatalf("unexpected diagnostics: %v", diagnostics)
		}

		return program
	}
	first, second := parse(), parse()

	if first.GetID() == 0 || first.Decls[0].GetID() == 0 || first.Stmts[0].GetID() == 0 {
		t.Fatal("one or more nodes have a zero ID")
	}

	if first.GetID() != second.GetID() || first.Decls[0].GetID() != second.Decls[0].GetID() || first.Stmts[0].GetID() != second.Stmts[0].GetID() {
		t.Fatal("node IDs are not deterministic")
	}

	if first.GetID() == first.Decls[0].GetID() || first.Decls[0].GetID() == first.Stmts[0].GetID() {
		t.Fatal("node IDs are not unique")
	}
}

func TestParserRecoversMultipleDiagnostics(t *testing.T) {
	_, diagnostics := Parse("recovery.nuru", []byte("print(,); if true { print(1) } print(,);"))

	if len(diagnostics) < 2 {
		t.Fatalf("got %d diagnostics, want at least 2: %v", len(diagnostics), diagnostics)
	}

	for _, diagnostic := range diagnostics {
		if diagnostic.Phase != "parser" || diagnostic.Span.Filename != "recovery.nuru" {
			t.Fatalf("malformed diagnostic: %v", diagnostic)
		}
	}
}

func TestParseExportAndImport(t *testing.T) {
	source := `import "math";
export var x int = 10;
export func add(a int, b int) int { export var res int = a + b; return res; }
`
	program, diagnostics := Parse("module.nuru", []byte(source))

	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}

	if len(program.Decls) != 3 {
		t.Fatalf("got %d decls, want 3", len(program.Decls))
	}

	if _, ok := program.Decls[0].(*ast.ImportStmt); !ok {
		t.Fatalf("decl[0] is %T, want *ast.ImportStmt", program.Decls[0])
	}

	if _, ok := program.Decls[1].(*ast.ExportStmt); !ok {
		t.Fatalf("decl[1] is %T, want *ast.ExportStmt", program.Decls[1])
	}
}

func FuzzParse(f *testing.F) {
	for _, seed := range []string{"", "var x int = 1;", "if true { print(1); }", "\xff\x00"} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, source string) {
		Parse("fuzz.nuru", []byte(source))
	})
}
