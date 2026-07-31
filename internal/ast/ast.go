package ast

import "github.com/kc-clintone/compilers/internal/source"

// Node is implemented by all AST nodes.
type Node interface {
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

// Base supplies source location behavior to AST nodes.
type Base struct {
	Span source.Span
}

func (b Base) GetSpan() source.Span { return b.Span }

// Program represents a complete Nuru compilation unit.
type Program struct {
	Base
	Decls []Decl
	Stmts []Stmt
}

// PrintStmt outputs values to standard output: print(expr1, expr2, ...)
type PrintStmt struct {
	Base
	Args []Expr
}

func (*PrintStmt) stmt() {}

// BlockStmt is a sequence of statements wrapped in braces.
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

// IfStmt represents conditional execution.
type IfStmt struct {
	Base
	Cond Expr
	Then *BlockStmt
	Else Stmt
}

func (*IfStmt) stmt() {}

// ForStmt represents a conditional loop: for (cond) { body }
type ForStmt struct {
	Base
	Cond Expr
	Body *BlockStmt
}

func (*ForStmt) stmt() {}

// BreakStmt exits the innermost loop.
type BreakStmt struct{ Base }

func (*BreakStmt) stmt() {}

// ContinueStmt jumps to the next iteration of the innermost loop.
type ContinueStmt struct{ Base }

func (*ContinueStmt) stmt() {}

// LiteralExpr represents primitive constants (integers, strings, booleans).
type LiteralExpr struct {
	Base
	Value any
}

func (*LiteralExpr) expr() {}

// UnaryExpr applies prefix operators (-x, !x).
type UnaryExpr struct {
	Base
	Op    string
	Right Expr
}

func (*UnaryExpr) expr() {}

// BinaryExpr applies infix operators (x + y, x == y).
type BinaryExpr struct {
	Base
	Left  Expr
	Op    string
	Right Expr
}

func (*BinaryExpr) expr() {}

// VarDecl represents variable declarations: var <Name> = <Init>
type VarDecl struct {
	Base
	Name string
	Init Expr
}

func (*VarDecl) decl() {}
func (*VarDecl) stmt() {}

// AssignStmt represents variable assignments: <Name> = <Value>
type AssignStmt struct {
	Base
	Name  string
	Value Expr
}

func (*AssignStmt) stmt() {}

// IdentExpr represents variable lookups by identifier name.
type IdentExpr struct {
	Base
	Name string
}

func (*IdentExpr) expr() {}

// FuncDecl represents function declarations: func <Name>(<Params>) { <Body> }
type FuncDecl struct {
	Base
	Name   string
	Params []string
	Body   *BlockStmt
}

func (*FuncDecl) decl() {}

// CallExpr represents function calls: <Callee>(<Args>)
type CallExpr struct {
	Base
	Callee string
	Args   []Expr
}

func (*CallExpr) expr() {}

// ReturnStmt represents returning a value from a function: return <Value>
type ReturnStmt struct {
	Base
	Value Expr
}

func (*ReturnStmt) stmt() {}
