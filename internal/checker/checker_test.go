package checker

import (
	"strings"
	"testing"

	"github.com/kc-clintone/compilers/internal/ast"
	"github.com/kc-clintone/compilers/internal/diagnostic"
	"github.com/kc-clintone/compilers/internal/parser"
)

func TestZeroArgumentBuiltin(t *testing.T) {
	program, diagnostics := parser.Parse("args.zing", []byte("var values []string = args(); print(len(values));"))

	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics: %v", diagnostics)
	}

	_, diagnostics = Check(program)
	if len(diagnostics) != 0 {
		t.Fatalf("check diagnostics: %v", diagnostics)
	}
}

func TestValidPrograms(t *testing.T) {
	tests := map[string]string{
		"forward recursion": `
func odd(n int) bool { if n == 0 { return false; } return even(n - 1); }
func even(n int) bool { if n == 0 { return true; } return odd(n - 1); }
print(even(4));`,
		"shadow and composites": `
type Pair struct { left int; right int; }
var pair Pair = Pair{right: 2, left: 1};
var values []int = []int{1, 2};
var counts map[string]int = map[string]int{"x": 1};
func value(x int) int { { var x int = 2; print(x); } return x; }
print(value(pair.left), values[0:2][1], counts["missing"]);`,
		"return paths": `
func choose(x int) int { if x > 0 { return 1; } else { return 2; } }
func select(x int) int { switch x { case 1: return 1; default: return 0; } }
print(choose(1), select(1));`,
		"all builtins": `
var xs []string = args();
var ys []string = append(xs, "x");
var m map[string]int = make(map[string]int);
var c char = char(65);
print(len(ys), int(c), int("2"), string(c), string(true), string(3), len(m));`,
	}

	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			program := mustParse(t, source)
			if _, diagnostics := Check(program); len(diagnostics) != 0 {
				t.Fatalf("unexpected diagnostics: %v", diagnostics)
			}
		})
	}
}

func TestInvalidPrograms(t *testing.T) {
	tests := map[string]struct {
		source, message string
	}{
		"unknown value":         {`print(missing);`, "unknown variable missing"},
		"unknown type":          {`var value Missing;`, "unknown type Missing"},
		"duplicate global":      {`var x int; var x int;`, "duplicate global x"},
		"duplicate value kinds": {`var x int; func x() { return; }`, "duplicate value name x"},
		"reserved main":         {`var main int;`, "reserved name main"},
		"reserved builtin":      {`var print int;`, "reserved built-in name print"},
		"wrong arity":           {`print(len());`, "len expects 1 arguments"},
		"wrong operands":        {`print(true + false);`, "operator + does not accept bool"},
		"invalid target":        {`1 = 2;`, "expression is not assignable"},
		"missing field":         {`type P struct { x int; y int; } var p P = P{x: 1};`, "missing field y"},
		"duplicate field":       {`type P struct { x int; } var p P = P{x: 1, x: 2};`, "duplicate field x"},
		"bad return":            {`func value() int { return; }`, "return value required"},
		"missing return":        {`func value(x bool) int { if x { return 1; } }`, "does not return on every path"},
		"illegal break":         {`break;`, "break is only valid inside a loop"},
		"illegal continue":      {`continue;`, "continue is only valid inside a loop"},
		"duplicate case":        {`switch 1 { case 1: print(1); case 1: print(2); }`, "duplicate switch case"},
		"invalid map key":       {`var values map[[]int]int;`, "invalid map key type"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			program := mustParse(t, test.source)
			info, diagnostics := Check(program)
			if info != nil {
				t.Fatal("checker returned Info for an invalid program")
			}
			if len(diagnostics) == 0 || !strings.Contains(diagnosticsText(diagnostics), test.message) {
				t.Fatalf("diagnostics %v do not contain %q", diagnostics, test.message)
			}
		})
	}
}

func TestInfoQueriesUseNodeIDs(t *testing.T) {
	program := mustParse(t, `
type Point struct { x int; }
var point Point = Point{x: 1};
var value int = 2;
print(value);`)
	info, diagnostics := Check(program)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	call := program.Stmts[0].(*ast.ExprStmt).Expr.(*ast.CallExpr)
	identifier := call.Args[0].(*ast.IdentExpr)
	if got := info.TypeOf(identifier); !got.Equal(Int) {
		t.Fatalf("identifier type = %v, want int", got)
	}
	if symbol, ok := info.SymbolOf(identifier); !ok || symbol.Name != "value" || symbol.Kind != SymbolVariable {
		t.Fatalf("resolved symbol = %#v, %v", symbol, ok)
	}
	if builtin, ok := info.BuiltinOf(call); !ok || builtin != BuiltinPrint {
		t.Fatalf("resolved builtin = %q, %v", builtin, ok)
	}
	point := program.Decls[0].(*ast.StructDecl)
	pointVar := program.Decls[1].(*ast.VarDecl)
	if structure, ok := info.StructOf(pointVar.Type); !ok || structure.Decl != point {
		t.Fatalf("resolved struct = %#v, %v", structure, ok)
	}
}

func mustParse(t *testing.T, source string) *ast.Program {
	t.Helper()
	program, diagnostics := parser.Parse("test.zing", []byte(source))
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics: %v", diagnostics)
	}
	return program
}

func diagnosticsText(diagnostics []diagnostic.Diagnostic) string {
	var messages []string
	for _, item := range diagnostics {
		messages = append(messages, item.Message)
	}
	return strings.Join(messages, "\n")
}
