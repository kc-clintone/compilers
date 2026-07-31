package compiler

import (
	"bytes"
	"fmt"

	"github.com/kc-clintone/compilers/internal/ast"
	"github.com/kc-clintone/compilers/internal/diagnostic"
	"github.com/kc-clintone/compilers/internal/source"
)

// Compiler transpiles Zing AST programs directly into clean Go source code.
type Compiler struct {
	buf   bytes.Buffer
	diags []diagnostic.Diagnostic
}

func New() *Compiler {
	return &Compiler{}
}

// Compile generates valid Go code from a Zing AST program.
func (c *Compiler) Compile(prog *ast.Program) (string, []diagnostic.Diagnostic) {
	c.emit("package main\n\n")
	c.emit("import \"fmt\"\n\n")
	c.emit("// Suppress unused import warning if fmt is unused\n")
	c.emit("var _ = fmt.Println\n\n")

	// Emit top-level function declarations
	for _, decl := range prog.Decls {
		c.compileDecl(decl)
	}

	// Emit top-level main function
	c.emit("func main() {\n")
	for _, stmt := range prog.Stmts {
		c.compileStmt(stmt, "\t")
	}
	c.emit("}\n")

	if len(c.diags) > 0 {
		return "", c.diags
	}
	return c.buf.String(), nil
}

func (c *Compiler) compileDecl(decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		c.compileFuncDecl(d)
	case *ast.VarDecl:
		c.compileVarDecl(d, "")
	default:
		c.error(decl.GetSpan(), fmt.Sprintf("unsupported declaration type %T", decl))
	}
}

// TODO: Compiler - Implement compileFuncDecl() for Stage 4!
// Instructions: Emit 'func <Name>(<params...>) any { <body> }'.
func (c *Compiler) compileFuncDecl(d *ast.FuncDecl) {
	// TODO: Compiler - Function declaration code generation goes here in Stage 4.
	c.error(d.GetSpan(), "function declaration code generation is not implemented in Stage 0")
}

// TODO: Compiler - Implement compileVarDecl() for Stage 4!
// Instructions: Emit '<name> := <init>' or 'var <name> = <init>'.
func (c *Compiler) compileVarDecl(d *ast.VarDecl, indent string) {
	// TODO: Compiler - Variable declaration code generation goes here in Stage 4.
	c.error(d.GetSpan(), "variable declaration code generation is not implemented in Stage 0")
}

func (c *Compiler) compileStmt(stmt ast.Stmt, indent string) {
	if stmt == nil {
		return
	}

	switch s := stmt.(type) {
	case *ast.PrintStmt:
		c.emit("%sfmt.Println(", indent)
		for i, arg := range s.Args {
			if i > 0 {
				c.emit(", ")
			}
			c.compileExpr(arg)
		}
		c.emit(")\n")

	case *ast.BlockStmt:
		c.emit(" {\n")
		for _, st := range s.Stmts {
			c.compileStmt(st, indent+"\t")
		}
		c.emit("%s}", indent)

	case *ast.IfStmt:
		c.emit("%sif ", indent)
		c.compileExpr(s.Cond)
		c.compileStmt(s.Then, indent)
		if s.Else != nil {
			c.emit(" else")
			if _, isIf := s.Else.(*ast.IfStmt); isIf {
				c.compileStmt(s.Else, "")
			} else {
				c.compileStmt(s.Else, indent)
			}
		}
		c.emit("\n")

	case *ast.ForStmt:
		c.emit("%sfor ", indent)
		c.compileExpr(s.Cond)
		c.compileStmt(s.Body, indent)
		c.emit("\n")

	case *ast.BreakStmt:
		c.emit("%sbreak\n", indent)

	case *ast.ContinueStmt:
		c.emit("%scontinue\n", indent)

	case *ast.ExprStmt:
		c.emit("%s", indent)
		c.compileExpr(s.Expr)
		c.emit("\n")

	case *ast.VarDecl:
		c.compileVarDecl(s, indent)

	case *ast.AssignStmt:
		c.compileAssignStmt(s, indent)

	case *ast.ReturnStmt:
		c.compileReturnStmt(s, indent)

	default:
		c.error(stmt.GetSpan(), fmt.Sprintf("unsupported statement type %T", stmt))
	}
}

// TODO: Compiler - Implement compileAssignStmt() for Stage 4!
// Instructions: Emit '<indent><Name> = <Value>\n'.
func (c *Compiler) compileAssignStmt(s *ast.AssignStmt, indent string) {
	// TODO: Compiler - Variable assignment code generation goes here in Stage 4.
	c.error(s.GetSpan(), "variable assignment code generation is not implemented in Stage 0")
}

// TODO: Compiler - Implement compileReturnStmt() for Stage 4!
// Instructions: Emit '<indent>return <Value>\n'.
func (c *Compiler) compileReturnStmt(s *ast.ReturnStmt, indent string) {
	// TODO: Compiler - Return statement code generation goes here in Stage 4.
	c.error(s.GetSpan(), "return statement code generation is not implemented in Stage 0")
}

func (c *Compiler) compileExpr(expr ast.Expr) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.LiteralExpr:
		switch v := e.Value.(type) {
		case string:
			c.emit("%q", v)
		case bool:
			c.emit("%t", v)
		default:
			c.emit("%v", v)
		}

	case *ast.UnaryExpr:
		c.emit("(%s", e.Op)
		c.compileExpr(e.Right)
		c.emit(")")

	case *ast.BinaryExpr:
		c.emit("(")
		c.compileExpr(e.Left)
		c.emit(" %s ", e.Op)
		c.compileExpr(e.Right)
		c.emit(")")

	case *ast.IdentExpr:
		c.compileIdentExpr(e)

	case *ast.CallExpr:
		c.compileCallExpr(e)

	default:
		c.error(expr.GetSpan(), fmt.Sprintf("unsupported expression type %T", expr))
	}
}

// TODO: Compiler - Implement compileIdentExpr() for Stage 4!
// Instructions: Emit identifier name 'e.Name'.
func (c *Compiler) compileIdentExpr(e *ast.IdentExpr) {
	// TODO: Compiler - Identifier code generation goes here in Stage 4.
	c.error(e.GetSpan(), "identifier code generation is not implemented in Stage 0")
}

// TODO: Compiler - Implement compileCallExpr() for Stage 4!
// Instructions: Emit '<Callee>(<args...>)'.
func (c *Compiler) compileCallExpr(e *ast.CallExpr) {
	// TODO: Compiler - Function call code generation goes here in Stage 4.
	c.error(e.GetSpan(), "function call code generation is not implemented in Stage 0")
}

func (c *Compiler) emit(format string, args ...any) {
	fmt.Fprintf(&c.buf, format, args...)
}

func (c *Compiler) error(span source.Span, msg string) {
	c.diags = append(c.diags, diagnostic.Diagnostic{Span: span, Phase: "compiler", Message: msg})
}
