// Package ast defines Zing's parser-independent abstract syntax tree.
package ast

import (
	"fmt"
	"reflect"

	"github.com/kc-clintone/compilers/internal/source"
)

// NodeID uniquely identifies a node within one parsed program.
type NodeID uint64

// Node is implemented by every source-spanning AST node.
type Node interface {
	GetID() NodeID
	GetSpan() source.Span
}

// Decl is implemented by top-level declarations.
type Decl interface {
	Node
	decl()
}

// Stmt is implemented by executable statements.
type Stmt interface {
	Node
	stmt()
}

// Expr is implemented by value-producing expressions.
type Expr interface {
	Node
	expr()
}

// Base supplies identity and source location behavior to AST nodes.
type Base struct {
	NodeID NodeID
	Span   source.Span
}

// GetID returns the node's program-local identity.
func (b Base) GetID() NodeID { return b.NodeID }

// GetSpan returns the node's half-open source span.
func (b Base) GetSpan() source.Span { return b.Span }

// AssignNodeIDs assigns deterministic pre-order IDs to every node and type
// reference in program. Parsers should call it exactly once after construction.
func AssignNodeIDs(program *Program) {
	next := NodeID(1)
	baseType := reflect.TypeFor[Base]()
	var visit func(reflect.Value)
	visit = func(value reflect.Value) {
		if !value.IsValid() {
			return
		}

		if value.Kind() == reflect.Interface {
			if value.IsNil() {
				return
			}
			visit(value.Elem())
			return
		}

		if value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return
			}
			visit(value.Elem())
			return
		}

		switch value.Kind() {
		case reflect.Struct:
			for i := 0; i < value.NumField(); i++ {
				field := value.Field(i)
				if value.Type().Field(i).Type == baseType {
					base := field.Addr().Interface().(*Base)
					base.NodeID = next
					next++
					continue
				}
				visit(field)
			}
		case reflect.Slice:
			for i := 0; i < value.Len(); i++ {
				visit(value.Index(i))
			}
		}
	}

	visit(reflect.ValueOf(program))
}

// Program is a complete Zing compilation unit.
type Program struct {
	Base
	Decls []Decl
	Stmts []Stmt
}

// TypeKind identifies syntactic and resolved Zing type categories.
type TypeKind int

// Zing type kinds.
const (
	TypeInvalid TypeKind = iota
	TypeInt
	TypeChar
	TypeString
	TypeBool
	TypeSlice
	TypeMap
	TypeNamed
	TypeVoid
)

// TypeRef is a syntactic reference to a Zing type.
type TypeRef struct {
	Base
	Kind      TypeKind
	Name      string
	Elem, Key *TypeRef
}

// DebugExpr returns a compact, fully parenthesized expression representation
// intended for parser tests and diagnostics during development.
func DebugExpr(expr Expr) string {
	switch value := expr.(type) {
	case *IdentExpr:
		return value.Name
	case *LiteralExpr:
		return fmt.Sprintf("%v", value.Value)
	case *UnaryExpr:
		return "(" + value.Op + DebugExpr(value.Right) + ")"
	case *BinaryExpr:
		return "(" + DebugExpr(value.Left) + " " + value.Op + " " + DebugExpr(value.Right) + ")"
	case *CallExpr:
		text := value.Callee + "("
		for i, argument := range value.Args {
			if i > 0 {
				text += ", "
			}
			text += DebugExpr(argument)
		}
		return text + ")"
	case *FieldExpr:
		return DebugExpr(value.Object) + "." + value.Name
	case *IndexExpr:
		return DebugExpr(value.Object) + "[" + DebugExpr(value.Index) + "]"
	case *SliceExpr:
		low, high := "", ""
		if value.Low != nil {
			low = DebugExpr(value.Low)
		}
		if value.High != nil {
			high = DebugExpr(value.High)
		}
		return DebugExpr(value.Object) + "[" + low + ":" + high + "]"
	case *MakeExpr:
		return "make(...)"
	case *CompositeExpr:
		return "composite{...}"
	default:
		return "<invalid>"
	}
}

// VarDecl declares a global or local variable.
type VarDecl struct {
	Base
	Name string
	Type *TypeRef
	Init Expr
}

func (*VarDecl) decl() {}
func (*VarDecl) stmt() {}

// Field declares one named struct field.
type Field struct {
	Name string
	Type *TypeRef
	Span source.Span
}

// StructDecl declares a named struct type.
type StructDecl struct {
	Base
	Name   string
	Fields []Field
}

func (*StructDecl) decl() {}

// Param declares one function parameter.
type Param struct {
	Name string
	Type *TypeRef
	Span source.Span
}

// FuncDecl declares a user function.
type FuncDecl struct {
	Base
	Name   string
	Params []Param
	Result *TypeRef
	Body   *BlockStmt
}

func (*FuncDecl) decl() {}

// BlockStmt is a lexically scoped statement sequence.
type BlockStmt struct {
	Base
	Stmts []Stmt
}

func (*BlockStmt) stmt() {}

// ExprStmt evaluates an expression for side effects.
type ExprStmt struct {
	Base
	Expr Expr
}

func (*ExprStmt) stmt() {}

// AssignStmt stores Value in Target.
type AssignStmt struct {
	Base
	Target, Value Expr
}

func (*AssignStmt) stmt() {}

// IfStmt conditionally executes one branch.
type IfStmt struct {
	Base
	Cond Expr
	Then *BlockStmt
	Else Stmt
}

func (*IfStmt) stmt() {}

// CaseClause is one case or default arm of a switch.
type CaseClause struct {
	Base
	Values  []Expr
	Body    []Stmt
	Default bool
}

// SwitchStmt selects the first matching case without fallthrough.
type SwitchStmt struct {
	Base
	Expr  Expr
	Cases []*CaseClause
}

func (*SwitchStmt) stmt() {}

// ForStmt represents both while-style and three-clause loops.
type ForStmt struct {
	Base
	Init Stmt
	Cond Expr
	Post Stmt
	Body *BlockStmt
}

func (*ForStmt) stmt() {}

// ReturnStmt exits the current function, optionally with a value.
type ReturnStmt struct {
	Base
	Value Expr
}

func (*ReturnStmt) stmt() {}

// BreakStmt exits the innermost loop.
type BreakStmt struct{ Base }

func (*BreakStmt) stmt() {}

// ContinueStmt starts the next iteration of the innermost loop.
type ContinueStmt struct{ Base }

func (*ContinueStmt) stmt() {}

// IdentExpr refers to a value by name.
type IdentExpr struct {
	Base
	Name string
}

func (*IdentExpr) expr() {}

// LiteralExpr contains a decoded primitive literal.
type LiteralExpr struct {
	Base
	Value any
	Type  TypeKind
}

func (*LiteralExpr) expr() {}

// UnaryExpr applies a prefix operator.
type UnaryExpr struct {
	Base
	Op    string
	Right Expr
}

func (*UnaryExpr) expr() {}

// BinaryExpr applies an infix operator.
type BinaryExpr struct {
	Base
	Left  Expr
	Op    string
	Right Expr
}

func (*BinaryExpr) expr() {}

// CallExpr invokes a named function or built-in.
type CallExpr struct {
	Base
	Callee string
	Args   []Expr
}

func (*CallExpr) expr() {}

// FieldExpr reads a named field.
type FieldExpr struct {
	Base
	Object Expr
	Name   string
}

func (*FieldExpr) expr() {}

// IndexExpr indexes a string, slice, or map.
type IndexExpr struct {
	Base
	Object, Index Expr
}

func (*IndexExpr) expr() {}

// SliceExpr forms a two-bound slice.
type SliceExpr struct {
	Base
	Object, Low, High Expr
}

func (*SliceExpr) expr() {}

// CompositeElem is one slice, map, or struct literal element.
type CompositeElem struct {
	Key            string
	KeyExpr, Value Expr
	Span           source.Span
}

// CompositeExpr constructs a composite value.
type CompositeExpr struct {
	Base
	Type  *TypeRef
	Elems []CompositeElem
}

func (*CompositeExpr) expr() {}

// MakeExpr allocates a slice or map.
type MakeExpr struct {
	Base
	Type *TypeRef
	Size Expr
}

func (*MakeExpr) expr() {}
