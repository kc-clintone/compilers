/*
===============================================================================
QUEST STAGE 3: THE LIVING ENGINE (Interpreter) [COMPLETED]
===============================================================================
Overview:
  The interpreter is one consumer of the parser's AST. It walks nodes directly,
  computes dynamically typed values, and stores variable bindings in nested
  environments. No Go source or native executable is produced on this path.

  This QUEST adds conditional execution, variable access, top-level function
  registration, isolated call frames, argument binding, and return propagation.

Tasks in this file:
  - [COMPLETED] TASK [EVAL-01]: Evaluate one conditional branch.
  - [COMPLETED] TASK [EVAL-02]: Define, assign, and read variables.
  - [COMPLETED] TASK [EVAL-03]: Register and call functions, including returns.

Commands:
  - Run tests:  go test ./internal/interpreter
  - Run code:   ./nuru-interpreter examples/03-interpreter.nuru
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

// Environment holds variable bindings for one lexical scope and optionally
// links to the next outer scope.
type Environment struct {
	parent *Environment
	values map[string]any
}

// NewEnvironment creates an empty scope whose unresolved names continue in parent.
func NewEnvironment(parent *Environment) *Environment {
	return &Environment{parent: parent, values: make(map[string]any)}
}

// Get returns the nearest binding for name, searching outward through parents.
func (e *Environment) Get(name string) (any, bool) {
	if v, ok := e.values[name]; ok {
		return v, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

// Set updates the nearest existing binding for name. It reports false rather
// than creating a new variable when the name is undefined in every scope.
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

// Define creates or replaces a binding in this environment only.
func (e *Environment) Define(name string, val any) {
	e.values[name] = val
}

// UserFunction stores a top-level function declaration.
type UserFunction struct {
	Decl *ast.FuncDecl
}

// ReturnVal carries a function result through nested statement evaluation.
type ReturnVal struct{ Value any }

// BreakVal signals that the innermost loop should stop.
type BreakVal struct{}

// ContinueVal signals that the innermost loop should start its next iteration.
type ContinueVal struct{}

// Interpreter evaluates Nuru AST nodes.
type Interpreter struct {
	env   *Environment
	funcs map[string]*UserFunction
	diags []diagnostic.Diagnostic
}

// New creates an interpreter with an empty global environment and function table.
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

// evalDecl dispatches a top-level declaration to its evaluator.
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

// evalFuncDecl registers a top-level function declaration by name. Function
// bodies run only when a later call expression invokes them.
func (in *Interpreter) evalFuncDecl(d *ast.FuncDecl, env *Environment) {
	in.funcs[d.Name] = &UserFunction{Decl: d}
}

// evalVarDecl evaluates an optional initializer and defines the result in env;
// declarations without initializers receive nil.
func (in *Interpreter) evalVarDecl(d *ast.VarDecl, env *Environment) {
	var value any
	if d.Init != nil {
		value = in.evalExpr(d.Init, env)
	}
	env.Define(d.Name, value)
}

// evalStmt evaluates one statement in env and returns a control-flow signal or
// expression result when the statement produces one.
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

// evalIfStmt evaluates its condition once, then evaluates at most one branch in
// the supplied environment and passes through that branch's result.
func (in *Interpreter) evalIfStmt(s *ast.IfStmt, env *Environment) any {
	condition := in.evalExpr(s.Cond, env)
	if isTruthy(condition) {
		return in.evalStmt(s.Then, env)
	}
	if s.Else != nil {
		return in.evalStmt(s.Else, env)
	}
	return nil
}

// evalAssignStmt evaluates the new value and updates the nearest existing
// binding, reporting a diagnostic when the variable is undefined.
func (in *Interpreter) evalAssignStmt(s *ast.AssignStmt, env *Environment) any {
	value := in.evalExpr(s.Value, env)
	if !env.Set(s.Name, value) {
		in.error(s.GetSpan(), fmt.Sprintf(
			"cannot assign to undefined variable '%s'", s.Name,
		))
	}
	return nil
}

// evalReturnStmt evaluates an optional return value and wraps it in ReturnVal so
// nested blocks can propagate it to the function call.
func (in *Interpreter) evalReturnStmt(s *ast.ReturnStmt, env *Environment) any {
	var value any
	if s.Value != nil {
		value = in.evalExpr(s.Value, env)
	}
	return ReturnVal{Value: value}
}

// evalExpr evaluates a value-producing AST node in env.
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

// evalBinaryOp applies a Nuru binary operator to already evaluated operands.
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

// evalIdentExpr returns the nearest value bound to the identifier or reports an
// undefined-variable diagnostic.
func (in *Interpreter) evalIdentExpr(e *ast.IdentExpr, env *Environment) any {
	value, ok := env.Get(e.Name)
	if !ok {
		in.error(e.GetSpan(), fmt.Sprintf("undefined variable '%s'", e.Name))
		return nil
	}
	return value
}

// evalCallExpr resolves a registered function, validates arity, evaluates each
// argument in the caller, binds parameters in a fresh call environment, runs
// the body, and unwraps ReturnVal.
func (in *Interpreter) evalCallExpr(e *ast.CallExpr, env *Environment) any {
	function, ok := in.funcs[e.Callee]
	if !ok {
		in.error(e.GetSpan(), fmt.Sprintf("undefined function '%s'", e.Callee))
		return nil
	}
	if len(e.Args) != len(function.Decl.Params) {
		in.error(e.GetSpan(), fmt.Sprintf(
			"function '%s' expects %d arguments, got %d",
			e.Callee, len(function.Decl.Params), len(e.Args),
		))
		return nil
	}

	callEnv := NewEnvironment(in.env)
	for index, argument := range e.Args {
		value := in.evalExpr(argument, env)
		callEnv.Define(function.Decl.Params[index], value)
	}

	result := in.evalStmt(function.Decl.Body, callEnv)
	if returned, ok := result.(ReturnVal); ok {
		return returned.Value
	}
	return nil
}

// error records an interpreter diagnostic at span.
func (in *Interpreter) error(span source.Span, msg string) {
	in.diags = append(in.diags, diagnostic.Diagnostic{Span: span, Phase: "interpreter", Message: msg})
}

// isTruthy implements Nuru truthiness: nil, false, and numeric zero are false;
// every other value is true.
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

// toInt converts the integer representations used by the parser and runtime.
func toInt(v any) (int, bool) {
	if n, ok := v.(int); ok {
		return n, true
	}
	if n, ok := v.(int64); ok {
		return int(n), true
	}
	return 0, false
}

// toIntPair converts both operands and succeeds only when both are integers.
func toIntPair(l, r any) (int, int, bool) {
	li, ok1 := toInt(l)
	ri, ok2 := toInt(r)
	return li, ri, ok1 && ok2
}

/*
===============================================================================
QUEST HINTS
===============================================================================
HINT [EVAL-01-HINT]:
  1. Evaluate s.Cond once in env.
  2. If the result is truthy, evaluate and return s.Then.
  3. Otherwise, evaluate and return s.Else when it exists.
  4. Return nil when no branch runs.

HINT [EVAL-02-HINT]:
  Declaration:
  1. Start with a nil value and evaluate d.Init only when it exists.
  2. Define d.Name in the current environment with that value.

  Assignment:
  1. Evaluate s.Value before changing the environment.
  2. Ask env.Set to update the nearest binding.
  3. If it returns false, record an undefined-assignment diagnostic.
  4. Return nil after handling the statement.

  Identifier:
  1. Ask env.Get for e.Name.
  2. Return the value when found; otherwise report an undefined-variable
     diagnostic and return nil.

HINT [EVAL-03-HINT]:
  Function declaration:
  1. Wrap d in UserFunction and store it under d.Name in the function table.

  Return:
  1. Start with nil and evaluate s.Value only when present.
  2. Wrap the result in ReturnVal so blocks do not mistake it for a normal value.

  Function call:
  1. Find e.Callee in the function table; diagnose an unknown function.
  2. Compare argument and parameter counts; diagnose a mismatch before running.
  3. Create a fresh environment whose parent is the interpreter's global env.
  4. Evaluate arguments in the caller's env and bind them to parameters by index.
  5. Evaluate the function body in the call environment.
  6. Unwrap ReturnVal when present; otherwise return nil.
===============================================================================
*/
