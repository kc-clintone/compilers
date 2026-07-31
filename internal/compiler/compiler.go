/*
===============================================================================
QUEST STAGE 4: THE CODE FORGE (Compiler / Transpiler)
===============================================================================
Overview:
  Transpile Zing AST nodes into clean, 1-to-1 Go source code. In this stage, you will
  implement code generation rules for conditional branches, variable declarations/assignments,
  and function definitions/calls.

Tasks in this file:
  - TASK [GEN-01]: Implement compileIfStmt() for Go code generation of conditional branches.
  - TASK [GEN-02]: Implement compileVarDecl(), compileAssignStmt(), and compileIdentExpr()
                   for Go variable declarations and assignments.
  - TASK [GEN-03]: Implement compileFuncDecl(), compileCallExpr(), and compileReturnStmt()
                   for Go function signatures, calls, and returns.

Commands:
  - Run tests:  go test ./internal/compiler
  - Compile:    ./zing-compiler examples/04-compiled.zing
  - Skip stage: ./savepoint.sh 4
  - Reset stage: ./savepoint.sh 3
===============================================================================
*/

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

// TASK [GEN-03]: Implement compileFuncDecl() to emit 'func <Name>(<params...>) any { <body> }'.
// See HINT [GEN-03-HINT] at the bottom of this file for details.
func (c *Compiler) compileFuncDecl(d *ast.FuncDecl) {
	c.error(d.GetSpan(), "function declaration code generation is not implemented in Stage 0/1/2/3")
}

// TASK [GEN-02]: Implement compileVarDecl() to emit variable declarations in Go.
// See HINT [GEN-02-HINT] at the bottom of this file for details.
func (c *Compiler) compileVarDecl(d *ast.VarDecl, indent string) {
	c.error(d.GetSpan(), "variable declaration code generation is not implemented in Stage 0/1/2/3")
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
		c.compileIfStmt(s, indent)

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

// TASK [GEN-01]: Implement compileIfStmt() to emit Go 'if <cond> { <then> } else { <else> }'.
// See HINT [GEN-01-HINT] at the bottom of this file for details.
func (c *Compiler) compileIfStmt(s *ast.IfStmt, indent string) {
	c.error(s.GetSpan(), "conditional branch code generation is not implemented in Stage 0/1/2/3")
}

// TASK [GEN-02]: Implement compileAssignStmt() to emit '<indent><Name> = <Value>\n'.
// See HINT [GEN-02-HINT] at the bottom of this file for details.
func (c *Compiler) compileAssignStmt(s *ast.AssignStmt, indent string) {
	c.error(s.GetSpan(), "variable assignment code generation is not implemented in Stage 0/1/2/3")
}

// TASK [GEN-03]: Implement compileReturnStmt() to emit '<indent>return <Value>\n'.
// See HINT [GEN-03-HINT] at the bottom of this file for details.
func (c *Compiler) compileReturnStmt(s *ast.ReturnStmt, indent string) {
	c.error(s.GetSpan(), "return statement code generation is not implemented in Stage 0/1/2/3")
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

// TASK [GEN-02]: Implement compileIdentExpr() to emit identifier name 'e.Name'.
// See HINT [GEN-02-HINT] at the bottom of this file for details.
func (c *Compiler) compileIdentExpr(e *ast.IdentExpr) {
	c.error(e.GetSpan(), "identifier code generation is not implemented in Stage 0/1/2/3")
}

// TASK [GEN-03]: Implement compileCallExpr() to emit '<Callee>(<args...>)'.
// See HINT [GEN-03-HINT] at the bottom of this file for details.
func (c *Compiler) compileCallExpr(e *ast.CallExpr) {
	c.error(e.GetSpan(), "function call code generation is not implemented in Stage 0/1/2/3")
}

func (c *Compiler) emit(format string, args ...any) {
	fmt.Fprintf(&c.buf, format, args...)
}

func (c *Compiler) error(span source.Span, msg string) {
	c.diags = append(c.diags, diagnostic.Diagnostic{Span: span, Phase: "compiler", Message: msg})
}

/*
===============================================================================
QUEST HINTS & SOLUTIONS
===============================================================================
HINT [GEN-01-HINT]:
  Implement compileIfStmt:
    func (c *Compiler) compileIfStmt(s *ast.IfStmt, indent string) {
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
    }

HINT [GEN-02-HINT]:
  Implement compileVarDecl, compileAssignStmt, and compileIdentExpr:
    func (c *Compiler) compileVarDecl(d *ast.VarDecl, indent string) {
        if indent == "" {
            c.emit("var %s any", d.Name)
            if d.Init != nil {
                c.emit(" = ")
                c.compileExpr(d.Init)
            }
            c.emit("\n")
        } else {
            c.emit("%svar %s any", indent, d.Name)
            if d.Init != nil {
                c.emit(" = ")
                c.compileExpr(d.Init)
            }
            c.emit("\n")
        }
    }

    func (c *Compiler) compileAssignStmt(s *ast.AssignStmt, indent string) {
        c.emit("%s%s = ", indent, s.Name)
        c.compileExpr(s.Value)
        c.emit("\n")
    }

    func (c *Compiler) compileIdentExpr(e *ast.IdentExpr) {
        c.emit("%s", e.Name)
    }

HINT [GEN-03-HINT]:
  Implement compileFuncDecl, compileCallExpr, and compileReturnStmt:
    func (c *Compiler) compileFuncDecl(d *ast.FuncDecl) {
        c.emit("func %s(", d.Name)
        for i, param := range d.Params {
            if i > 0 {
                c.emit(", ")
            }
            c.emit("%s any", param)
        }
        c.emit(") any")

        c.compileStmt(d.Body, "")
        c.emit("\n\n")
    }

    func (c *Compiler) compileCallExpr(e *ast.CallExpr) {
        c.emit("%s(", e.Callee)
        for i, arg := range e.Args {
            if i > 0 {
                c.emit(", ")
            }
            c.compileExpr(arg)
        }
        c.emit(")")
    }

    func (c *Compiler) compileReturnStmt(s *ast.ReturnStmt, indent string) {
        c.emit("%sreturn ", indent)
        if s.Value != nil {
            c.compileExpr(s.Value)
        } else {
            c.emit("nil")
        }
        c.emit("\n")
    }
===============================================================================
*/
