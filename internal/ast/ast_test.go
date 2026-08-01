package ast

import (
	"reflect"
	"testing"

	"github.com/kc-clintone/compilers/internal/source"
)

func TestBaseExposesIdentityAndSpan(t *testing.T) {
	span := source.Span{
		Filename: "program.nuru",
		Start:    source.Position{Offset: 4, Line: 2, Column: 3},
		End:      source.Position{Offset: 8, Line: 2, Column: 7},
	}
	base := Base{NodeID: 42, Span: span}

	if base.GetID() != 42 {
		t.Fatalf("GetID() = %d, want 42", base.GetID())
	}

	if base.GetSpan() != span {
		t.Fatalf("GetSpan() = %#v, want %#v", base.GetSpan(), span)
	}
}

func TestAssignNodeIDsUsesDeterministicPreorder(t *testing.T) {
	integerType := &TypeRef{Kind: TypeInt}
	sliceType := &TypeRef{Kind: TypeSlice, Elem: integerType}
	left := &LiteralExpr{Value: 1, Type: TypeInt}
	right := &IdentExpr{Name: "other"}
	initializer := &BinaryExpr{Left: left, Op: "+", Right: right}
	declaration := &VarDecl{Name: "values", Type: sliceType, Init: initializer}
	argument := &LiteralExpr{Value: true, Type: TypeBool}
	call := &CallExpr{Callee: "print", Args: []Expr{argument}}
	statement := &ExprStmt{Expr: call}
	program := &Program{Decls: []Decl{declaration}, Stmts: []Stmt{statement}}

	AssignNodeIDs(program)
	nodes := []Node{program, declaration, sliceType, integerType, initializer, left, right, statement, call, argument}

	for index, node := range nodes {
		want := NodeID(index + 1)

		if node.GetID() != want {
			t.Fatalf("node %d (%T) ID = %d, want %d", index, node, node.GetID(), want)
		}
	}

	AssignNodeIDs(program)
	for index, node := range nodes {
		want := NodeID(index + 1)

		if node.GetID() != want {
			t.Fatalf("second assignment changed node %T to ID %d, want %d", node, node.GetID(), want)
		}
	}
}

func TestDebugExpr(t *testing.T) {
	identifier := &IdentExpr{Name: "values"}
	one := &LiteralExpr{Value: 1, Type: TypeInt}
	two := &LiteralExpr{Value: 2, Type: TypeInt}
	tests := []struct {
		name string
		expr Expr
		want string
	}{
		{"identifier", identifier, "values"},
		{"literal", one, "1"},
		{"unary", &UnaryExpr{Op: "-", Right: one}, "(-1)"},
		{"binary", &BinaryExpr{Left: one, Op: "+", Right: two}, "(1 + 2)"},
		{"call", &CallExpr{Callee: "print", Args: []Expr{one, two}}, "print(1, 2)"},
		{"field", &FieldExpr{Object: identifier, Name: "length"}, "values.length"},
		{"index", &IndexExpr{Object: identifier, Index: one}, "values[1]"},
		{"slice", &SliceExpr{Object: identifier, Low: one, High: two}, "values[1:2]"},
		{"open slice", &SliceExpr{Object: identifier}, "values[:]"},
		{"make", &MakeExpr{}, "make(...)"},
		{"composite", &CompositeExpr{}, "composite{...}"},
		{"invalid", nil, "<invalid>"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := DebugExpr(test.expr); got != test.want {
				t.Fatalf("DebugExpr() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestAssignNodeIDsPreservesSpans(t *testing.T) {
	span := source.Span{Filename: "span.nuru", Start: source.Position{Offset: 1, Line: 1, Column: 2}}
	program := &Program{Base: Base{Span: span}}

	AssignNodeIDs(program)
	if !reflect.DeepEqual(program.GetSpan(), span) {
		t.Fatalf("span changed to %#v", program.GetSpan())
	}
}
