/*
===============================================================================
QUEST STAGE 4: THE CODE FORGE (Compiler / Transpiler)
===============================================================================
Overview:
  The compiler is the second consumer of the parser's AST. It emits equivalent,
  readable Go source and delegates native-code generation to go build. This is
  source-to-source compilation rather than a bytecode or assembly backend.

  This QUEST adds Go emission for the same branches, variables, identifiers,
  functions, calls, and returns already supported by the interpreter.

Tasks in this file:
  - TASK [GEN-01]: Emit conditional branches.
  - TASK [GEN-02]: Emit variable declarations, assignments, and identifiers.
  - TASK [GEN-03]: Emit functions, calls, and returns.

Commands:
  - Run tests:  go test ./internal/compiler
  - Compile:    ./nuru-compiler examples/04-compiled.nuru
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

// Compiler transpiles Nuru AST programs directly into clean Go source code.
type Compiler struct {
	buf   bytes.Buffer
	diags []diagnostic.Diagnostic
}

// New creates a compiler with an empty output buffer and diagnostic list.
func New() *Compiler {
	return &Compiler{}
}

// Compile generates valid Go code from a Nuru AST program.
func (c *Compiler) Compile(prog *ast.Program) (string, []diagnostic.Diagnostic) {
	c.emit("package main\n\n")
	c.emit("import (\n\t\"fmt\"\n\t\"reflect\"\n)\n\n")
	c.emit("%s", nuruRuntime)

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

// compileDecl dispatches one top-level AST declaration.
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

// compileFuncDecl emits a Go function whose parameters and result use any, then
// delegates body emission to compileStmt.
//
// TASK [GEN-03]: Emit func Name(params any) any and the function body.
// Use c.emit, c.compileStmt, d.Name, d.Params, and d.Body.
// See HINT [GEN-03-HINT] at the bottom of this file for details.
func (c *Compiler) compileFuncDecl(d *ast.FuncDecl) {
	c.error(d.GetSpan(), "function declaration code generation is not implemented in Stage 0/1/2/3")
}

// compileVarDecl emits an indented Go var declaration using any and includes an
// initializer only when the Nuru declaration has one.
//
// TASK [GEN-02]: Emit var Name any, its optional initializer, and a newline.
// Use c.emit, c.compileExpr, d.Name, d.Init, and indent.
// See HINT [GEN-02-HINT] at the bottom of this file for details.
func (c *Compiler) compileVarDecl(d *ast.VarDecl, indent string) {
	c.error(d.GetSpan(), "variable declaration code generation is not implemented in Stage 0/1/2/3")
}

// compileStmt emits one statement with the supplied indentation and dispatches
// nested expressions or statements to their dedicated emitters.
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
		c.emit("%sfor nuruTruthy(", indent)
		c.compileExpr(s.Cond)
		c.emit(")")
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

// compileIfStmt emits an if statement whose condition passes through
// nuruTruthy. It preserves Go's required adjacency for else and else-if.
//
// TASK [GEN-01]: Emit if nuruTruthy(condition), then, and optional else branches.
// Use c.emit, c.compileExpr, c.compileStmt, s.Cond, s.Then, and s.Else.
// See HINT [GEN-01-HINT] at the bottom of this file for details.
func (c *Compiler) compileIfStmt(s *ast.IfStmt, indent string) {
	c.error(s.GetSpan(), "conditional branch code generation is not implemented in Stage 0/1/2/3")
}

// compileAssignStmt emits an indented assignment and trailing newline.
//
// TASK [GEN-02]: Use c.emit, c.compileExpr, indent, s.Name, and s.Value.
// See HINT [GEN-02-HINT] at the bottom of this file for details.
func (c *Compiler) compileAssignStmt(s *ast.AssignStmt, indent string) {
	c.error(s.GetSpan(), "variable assignment code generation is not implemented in Stage 0/1/2/3")
}

// compileReturnStmt emits a return value, using nil for a bare Nuru return so
// the generated Go function always satisfies its any result type.
//
// TASK [GEN-03]: Use c.emit, c.compileExpr, indent, and s.Value.
// See HINT [GEN-03-HINT] at the bottom of this file for details.
func (c *Compiler) compileReturnStmt(s *ast.ReturnStmt, indent string) {
	c.error(s.GetSpan(), "return statement code generation is not implemented in Stage 0/1/2/3")
}

// compileExpr emits one value-producing expression using runtime helpers where
// Go's static operators do not directly model Nuru's dynamic semantics.
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
		c.emit("nuruUnary(%q, ", e.Op)
		c.compileExpr(e.Right)
		c.emit(")")

	case *ast.BinaryExpr:
		c.emit("nuruBinary(%q, ", e.Op)
		c.compileExpr(e.Left)
		c.emit(", ")
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

// compileIdentExpr emits the identifier spelling stored in the AST.
//
// TASK [GEN-02]: Use c.emit and e.Name.
// See HINT [GEN-02-HINT] at the bottom of this file for details.
func (c *Compiler) compileIdentExpr(e *ast.IdentExpr) {
	c.error(e.GetSpan(), "identifier code generation is not implemented in Stage 0/1/2/3")
}

// compileCallExpr emits a named callee and comma-separated argument expressions.
//
// TASK [GEN-03]: Use c.emit, c.compileExpr, e.Callee, and e.Args.
// See HINT [GEN-03-HINT] at the bottom of this file for details.
func (c *Compiler) compileCallExpr(e *ast.CallExpr) {
	c.error(e.GetSpan(), "function call code generation is not implemented in Stage 0/1/2/3")
}

// emit appends formatted Go source to the compiler buffer.
func (c *Compiler) emit(format string, args ...any) {
	fmt.Fprintf(&c.buf, format, args...)
}

// error records a compiler diagnostic at span.
func (c *Compiler) error(span source.Span, msg string) {
	c.diags = append(c.diags, diagnostic.Diagnostic{Span: span, Phase: "compiler", Message: msg})
}

// nuruRuntime is prepended to generated programs to preserve Nuru's dynamic
// truthiness, integer, unary, and binary semantics in ordinary Go.
const nuruRuntime = `// Runtime helpers preserve Nuru's dynamic value semantics in generated Go.
var _ = fmt.Println
var _ = reflect.DeepEqual

func nuruTruthy(value any) bool {
	if value == nil {
		return false
	}
	if boolean, ok := value.(bool); ok {
		return boolean
	}
	if integer, ok := nuruInt(value); ok {
		return integer != 0
	}
	return true
}

func nuruInt(value any) (int, bool) {
	switch integer := value.(type) {
	case int:
		return integer, true
	case int64:
		return int(integer), true
	default:
		return 0, false
	}
}

func nuruInts(left, right any) (int, int, bool) {
	l, leftOK := nuruInt(left)
	r, rightOK := nuruInt(right)
	return l, r, leftOK && rightOK
}

func nuruUnary(operator string, value any) any {
	switch operator {
	case "-":
		if integer, ok := nuruInt(value); ok {
			return -integer
		}
	case "!":
		return !nuruTruthy(value)
	}
	panic(fmt.Sprintf("invalid Nuru unary operation %s %T", operator, value))
}

func nuruBinary(operator string, left, right any) any {
	switch operator {
	case "+":
		if l, r, ok := nuruInts(left, right); ok {
			return l + r
		}
		if l, leftOK := left.(string); leftOK {
			if r, rightOK := right.(string); rightOK {
				return l + r
			}
		}
	case "-":
		if l, r, ok := nuruInts(left, right); ok { return l - r }
	case "*":
		if l, r, ok := nuruInts(left, right); ok { return l * r }
	case "/":
		if l, r, ok := nuruInts(left, right); ok {
			if r == 0 { panic("division by zero") }
			return l / r
		}
	case "%":
		if l, r, ok := nuruInts(left, right); ok {
			if r == 0 { panic("modulo by zero") }
			return l % r
		}
	case "==":
		return reflect.DeepEqual(left, right)
	case "!=":
		return !reflect.DeepEqual(left, right)
	case "<":
		if l, r, ok := nuruInts(left, right); ok { return l < r }
	case "<=":
		if l, r, ok := nuruInts(left, right); ok { return l <= r }
	case ">":
		if l, r, ok := nuruInts(left, right); ok { return l > r }
	case ">=":
		if l, r, ok := nuruInts(left, right); ok { return l >= r }
	case "&&":
		return nuruTruthy(left) && nuruTruthy(right)
	case "||":
		return nuruTruthy(left) || nuruTruthy(right)
	}
	panic(fmt.Sprintf("invalid Nuru binary operation %T %s %T", left, operator, right))
}

`

/*
===============================================================================
QUEST HINTS
===============================================================================
HINT [GEN-01-HINT]:
  1. Emit the indentation and opening Go if with a nuruTruthy call.
  2. Compile the condition, close the call, and compile the then block.
  3. When an else branch exists, emit " else" immediately after the block.
  4. Compile an else-if without additional indentation; compile an else block
     with the current indentation.
  5. Finish the complete branch with one newline.

HINT [GEN-02-HINT]:
  Variable declaration:
  1. Emit indentation, var, the name, and the any type.
  2. When an initializer exists, emit equals and compile the expression.
  3. End the declaration with a newline.

  Assignment:
  1. Emit indentation, the name, and equals.
  2. Compile the value and emit a newline.

  Identifier:
  1. Emit e.Name unchanged.

HINT [GEN-03-HINT]:
  Function declaration:
  1. Emit func, the function name, and an opening parenthesis.
  2. Emit comma-separated parameter names, adding the any type to each.
  3. Close the signature with an any result, compile the body, and add spacing
     before the next top-level declaration.

  Function call:
  1. Emit the callee and opening parenthesis.
  2. Compile comma-separated argument expressions and close the call.

  Return:
  1. Emit indentation and return.
  2. Compile s.Value when present; otherwise emit nil.
  3. End the statement with a newline.
===============================================================================
*/
