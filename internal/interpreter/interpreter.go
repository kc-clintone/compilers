/*
===============================================================================
QUEST STAGE 3: THE LIVING ENGINE (Interpreter)
===============================================================================
Overview:
  Evaluate the AST dynamically by walking AST nodes and managing lexical environment
  scopes. In this stage, you will implement execution logic for conditional branches,
  variable environments, and function invocation call frames.

Tasks in this file:
  - TASK [EVAL-01]: Implement evalIfStmt() for conditional branches.
                    Uses isTruthy() and evalStmt().
  - TASK [EVAL-02]: Implement evalVarDecl(), evalAssignStmt(), and evalIdentExpr()
                    for environment variable storage and lookup.
                    Uses env.Define(), env.Set(), env.Get().
  - TASK [EVAL-03]: Implement evalFuncDecl(), evalCallExpr(), and evalReturnStmt()
                    for function registration, call stack frames, and returns.
                    Uses UserFunction, NewEnvironment(), and ReturnVal.

Commands:
  - Run tests:  go test ./internal/interpreter
  - Run code:   ./zing run examples/03-interpreter.zing
  - Skip stage: ./savepoint.sh 3
  - Reset stage: ./savepoint.sh 2
===============================================================================
*/

package interpreter

import (
	"fmt"
	"reflect"

	"github.com/kc-clintone/compilers/internal/ast"
	"github.com/kc-clintone/compilers/internal/diagnostic"
	"github.com/kc-clintone/compilers/internal/source"
)

// Environment holds variable bindings for a lexical scope.
type Environment struct {
	parent *Environment
	values map[string]any
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{parent: parent, values: make(map[string]any)}
}

func (e *Environment) Get(name string) (any, bool) {
	if v, ok := e.values[name]; ok {
		return v, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

func (e *Environment) Set(name string, val any) bool {
	if _, ok := e.values[name]; ok {
		e.values[name] = val
		return true
	}
	if e.parent != nil {
		return e.parent.Set(name, val)
	}
	return false
}

func (e *Environment) Define(name string, val any) {
	e.values[name] = val
}

// UserFunction stores a function declaration and its lexical closure environment.
type UserFunction struct {
	Decl    *ast.FuncDecl
	Closure *Environment
}

// Control flow signals
type ReturnVal struct{ Value any }
type BreakVal struct{}
type ContinueVal struct{}

// Interpreter evaluates Zing AST nodes.
type Interpreter struct {
	env   *Environment
	funcs map[string]*UserFunction
	diags []diagnostic.Diagnostic
}

func New() *Interpreter {
	return &Interpreter{
		env:   NewEnvironment(nil),
		funcs: make(map[string]*UserFunction),
	}
}

// Interpret evaluates all declarations and statements in a program.
func (in *Interpreter) Interpret(prog *ast.Program) []diagnostic.Diagnostic {
	// First pass: register function declarations
	for _, decl := range prog.Decls {
		in.evalDecl(decl)
	}

	// Second pass: execute top-level statements
	for _, stmt := range prog.Stmts {
		res := in.evalStmt(stmt, in.env)
		if _, isRet := res.(ReturnVal); isRet {
			break
		}
	}

	return in.diags
}

func (in *Interpreter) evalDecl(decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		in.evalFuncDecl(d, in.env)
	case *ast.VarDecl:
		in.evalVarDecl(d, in.env)
	default:
		in.error(decl.GetSpan(), fmt.Sprintf("unsupported declaration %T", decl))
	}
}

// TASK [EVAL-03]: Implement evalFuncDecl() to register user functions in in.funcs[d.Name].
// Uses UserFunction{Decl: d, Closure: env}.
// See HINT [EVAL-03-HINT] at the bottom of this file for details.
func (in *Interpreter) evalFuncDecl(d *ast.FuncDecl, env *Environment) {
	in.error(d.GetSpan(), "function declarations are not evaluated in Stage 0/1/2")
}

// TASK [EVAL-02]: Implement evalVarDecl() to evaluate initializer and define variable in env.
// Uses env.Define(d.Name, val).
// See HINT [EVAL-02-HINT] at the bottom of this file for details.
func (in *Interpreter) evalVarDecl(d *ast.VarDecl, env *Environment) {
	in.error(d.GetSpan(), "variable declarations are not evaluated in Stage 0/1/2")
}

func (in *Interpreter) evalStmt(stmt ast.Stmt, env *Environment) any {
	if stmt == nil {
		return nil
	}

	switch s := stmt.(type) {
	case *ast.PrintStmt:
		var vals []any
		for _, arg := range s.Args {
			vals = append(vals, in.evalExpr(arg, env))
		}
		for i, v := range vals {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(v)
		}
		fmt.Println()
		return nil

	case *ast.BlockStmt:
		blockEnv := NewEnvironment(env)
		for _, st := range s.Stmts {
			res := in.evalStmt(st, blockEnv)
			if res != nil {
				return res // Pass through ReturnVal, BreakVal, ContinueVal
			}
		}
		return nil

	case *ast.IfStmt:
		return in.evalIfStmt(s, env)

	case *ast.ForStmt:
		for {
			condVal := in.evalExpr(s.Cond, env)
			if !isTruthy(condVal) {
				break
			}
			res := in.evalStmt(s.Body, env)
			if _, isBreak := res.(BreakVal); isBreak {
				break
			}
			if _, isRet := res.(ReturnVal); isRet {
				return res
			}
		}
		return nil

	case *ast.BreakStmt:
		return BreakVal{}

	case *ast.ContinueStmt:
		return ContinueVal{}

	case *ast.ExprStmt:
		return in.evalExpr(s.Expr, env)

	case *ast.VarDecl:
		in.evalVarDecl(s, env)
		return nil

	case *ast.AssignStmt:
		return in.evalAssignStmt(s, env)

	case *ast.ReturnStmt:
		return in.evalReturnStmt(s, env)

	default:
		in.error(stmt.GetSpan(), fmt.Sprintf("unsupported statement type %T", stmt))
		return nil
	}
}

// TASK [EVAL-01]: Implement evalIfStmt() for conditional branches.
// Uses isTruthy(condVal) and evalStmt().
// See HINT [EVAL-01-HINT] at the bottom of this file for details.
func (in *Interpreter) evalIfStmt(s *ast.IfStmt, env *Environment) any {
	in.error(s.GetSpan(), "conditional branch statements are not evaluated in Stage 0/1/2")
	return nil
}

// TASK [EVAL-02]: Implement evalAssignStmt() to evaluate value and update env using env.Set(s.Name, val).
// See HINT [EVAL-02-HINT] at the bottom of this file for details.
func (in *Interpreter) evalAssignStmt(s *ast.AssignStmt, env *Environment) any {
	in.error(s.GetSpan(), "variable assignments are not evaluated in Stage 0/1/2")
	return nil
}

// TASK [EVAL-03]: Implement evalReturnStmt() to return ReturnVal{Value: val}.
// See HINT [EVAL-03-HINT] at the bottom of this file for details.
func (in *Interpreter) evalReturnStmt(s *ast.ReturnStmt, env *Environment) any {
	in.error(s.GetSpan(), "return statements are not evaluated in Stage 0/1/2")
	return nil
}

func (in *Interpreter) evalExpr(expr ast.Expr, env *Environment) any {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.LiteralExpr:
		return e.Value

	case *ast.UnaryExpr:
		rightVal := in.evalExpr(e.Right, env)
		switch e.Op {
		case "-":
			if n, ok := toInt(rightVal); ok {
				return -n
			}
			in.error(e.GetSpan(), "operator '-' requires integer operand")
		case "!":
			return !isTruthy(rightVal)
		}
		return nil

	case *ast.BinaryExpr:
		leftVal := in.evalExpr(e.Left, env)
		rightVal := in.evalExpr(e.Right, env)
		return in.evalBinaryOp(e, leftVal, rightVal)

	case *ast.IdentExpr:
		return in.evalIdentExpr(e, env)

	case *ast.CallExpr:
		return in.evalCallExpr(e, env)

	default:
		in.error(expr.GetSpan(), fmt.Sprintf("unsupported expression type %T", expr))
		return nil
	}
}

func (in *Interpreter) evalBinaryOp(e *ast.BinaryExpr, left, right any) any {
	switch e.Op {
	case "+":
		if l, ok1 := toInt(left); ok1 {
			if r, ok2 := toInt(right); ok2 {
				return l + r
			}
		}
		if l, ok1 := left.(string); ok1 {
			if r, ok2 := right.(string); ok2 {
				return l + r
			}
		}
		in.error(e.GetSpan(), "operator '+' requires matching integers or strings")
	case "-":
		if l, r, ok := toIntPair(left, right); ok {
			return l - r
		}
	case "*":
		if l, r, ok := toIntPair(left, right); ok {
			return l * r
		}
	case "/":
		if l, r, ok := toIntPair(left, right); ok {
			if r == 0 {
				in.error(e.GetSpan(), "division by zero")
				return 0
			}
			return l / r
		}
	case "%":
		if l, r, ok := toIntPair(left, right); ok {
			if r == 0 {
				in.error(e.GetSpan(), "modulo by zero")
				return 0
			}
			return l % r
		}
	case "==":
		return reflect.DeepEqual(left, right)
	case "!=":
		return !reflect.DeepEqual(left, right)
	case "<":
		if l, r, ok := toIntPair(left, right); ok {
			return l < r
		}
	case "<=":
		if l, r, ok := toIntPair(left, right); ok {
			return l <= r
		}
	case ">":
		if l, r, ok := toIntPair(left, right); ok {
			return l > r
		}
	case ">=":
		if l, r, ok := toIntPair(left, right); ok {
			return l >= r
		}
	case "&&":
		return isTruthy(left) && isTruthy(right)
	case "||":
		return isTruthy(left) || isTruthy(right)
	}
	return nil
}

// TASK [EVAL-02]: Implement evalIdentExpr() to look up variable value using env.Get(e.Name).
// See HINT [EVAL-02-HINT] at the bottom of this file for details.
func (in *Interpreter) evalIdentExpr(e *ast.IdentExpr, env *Environment) any {
	in.error(e.GetSpan(), fmt.Sprintf("undefined variable '%s'", e.Name))
	return nil
}

// TASK [EVAL-03]: Implement evalCallExpr() to look up user function, create argument environment, and execute body.
// Uses NewEnvironment(), evalStmt(), and ReturnVal.
// See HINT [EVAL-03-HINT] at the bottom of this file for details.
func (in *Interpreter) evalCallExpr(e *ast.CallExpr, env *Environment) any {
	in.error(e.GetSpan(), fmt.Sprintf("undefined function '%s'", e.Callee))
	return nil
}

func (in *Interpreter) error(span source.Span, msg string) {
	in.diags = append(in.diags, diagnostic.Diagnostic{Span: span, Phase: "interpreter", Message: msg})
}

func isTruthy(v any) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	if n, ok := toInt(v); ok {
		return n != 0
	}
	return true
}

func toInt(v any) (int, bool) {
	if n, ok := v.(int); ok {
		return n, true
	}
	if n, ok := v.(int64); ok {
		return int(n), true
	}
	return 0, false
}

func toIntPair(l, r any) (int, int, bool) {
	li, ok1 := toInt(l)
	ri, ok2 := toInt(r)
	return li, ri, ok1 && ok2
}

/*
===============================================================================
QUEST HINTS & SOLUTIONS
===============================================================================
HINT [EVAL-01-HINT]:
  Implement evalIfStmt:
    func (in *Interpreter) evalIfStmt(s *ast.IfStmt, env *Environment) any {
        condVal := in.evalExpr(s.Cond, env)
        if isTruthy(condVal) {
            return in.evalStmt(s.Then, env)
        } else if s.Else != nil {
            return in.evalStmt(s.Else, env)
        }
        return nil
    }

HINT [EVAL-02-HINT]:
  Implement evalVarDecl, evalAssignStmt, and evalIdentExpr:
    func (in *Interpreter) evalVarDecl(d *ast.VarDecl, env *Environment) {
        var val any
        if d.Init != nil {
            val = in.evalExpr(d.Init, env)
        }
        env.Define(d.Name, val)
    }

    func (in *Interpreter) evalAssignStmt(s *ast.AssignStmt, env *Environment) any {
        val := in.evalExpr(s.Value, env)
        if !env.Set(s.Name, val) {
            in.error(s.GetSpan(), fmt.Sprintf("cannot assign to undefined variable '%s'", s.Name))
        }
        return nil
    }

    func (in *Interpreter) evalIdentExpr(e *ast.IdentExpr, env *Environment) any {
        val, ok := env.Get(e.Name)
        if !ok {
            in.error(e.GetSpan(), fmt.Sprintf("undefined variable '%s'", e.Name))
            return nil
        }
        return val
    }

HINT [EVAL-03-HINT]:
  Implement evalFuncDecl, evalCallExpr, and evalReturnStmt:
    func (in *Interpreter) evalFuncDecl(d *ast.FuncDecl, env *Environment) {
        in.funcs[d.Name] = &UserFunction{Decl: d, Closure: env}
    }

    func (in *Interpreter) evalReturnStmt(s *ast.ReturnStmt, env *Environment) any {
        var val any
        if s.Value != nil {
            val = in.evalExpr(s.Value, env)
        }
        return ReturnVal{Value: val}
    }

    func (in *Interpreter) evalCallExpr(e *ast.CallExpr, env *Environment) any {
        fn, ok := in.funcs[e.Callee]
        if !ok {
            in.error(e.GetSpan(), fmt.Sprintf("undefined function '%s'", e.Callee))
            return nil
        }

        if len(e.Args) != len(fn.Decl.Params) {
            in.error(e.GetSpan(), fmt.Sprintf("function '%s' expects %d arguments, got %d", e.Callee, len(fn.Decl.Params), len(e.Args)))
            return nil
        }

        callEnv := NewEnvironment(fn.Closure)
        for i, argExpr := range e.Args {
            argVal := in.evalExpr(argExpr, env)
            callEnv.Define(fn.Decl.Params[i], argVal)
        }

        res := in.evalStmt(fn.Decl.Body, callEnv)
        if ret, isRet := res.(ReturnVal); isRet {
            return ret.Value
        }
        return nil
    }
===============================================================================
*/
