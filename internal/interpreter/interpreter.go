// Package interpreter executes checked Zing syntax trees.
package interpreter

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/kc-clintone/compilers/internal/ast"
	"github.com/kc-clintone/compilers/internal/checker"
	"github.com/kc-clintone/compilers/internal/source"
)

// FileSystem is the complete-file I/O boundary used by Zing built-ins.
type FileSystem interface {
	ReadFile(string) ([]byte, error)
	WriteFile(string, []byte, os.FileMode) error
}

type osFileSystem struct{}

// ReadFile delegates to the operating-system filesystem.
func (osFileSystem) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }

// WriteFile delegates to the operating-system filesystem.
func (osFileSystem) WriteFile(name string, data []byte, mode os.FileMode) error {
	return os.WriteFile(name, data, mode)
}

// Options supplies process boundaries for one interpreter run.
type Options struct {
	Args   []string
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
	Files  FileSystem
}

// RuntimeError is a source-located failure raised while executing Zing code.
type RuntimeError struct {
	Span    source.Span
	Message string
}

// Error formats the runtime error as a stable Zing diagnostic.
func (e *RuntimeError) Error() string { return fmt.Sprintf("%s: runtime: %s", e.Span, e.Message) }

type value interface{ zingValue() }
type intValue int
type charValue byte
type stringValue string
type boolValue bool

func (intValue) zingValue()    {}
func (charValue) zingValue()   {}
func (stringValue) zingValue() {}
func (boolValue) zingValue()   {}

type sliceValue struct {
	element checker.Type
	items   []value
}

func (sliceValue) zingValue() {}

type mapValue struct {
	key, element checker.Type
	items        map[value]value
}

func (*mapValue) zingValue() {}

type structValue struct {
	name   string
	fields map[string]value
}

func (*structValue) zingValue() {}

type environment struct {
	parent *environment
	values map[string]value
}

func (environment *environment) get(name string) (value, bool) {
	for current := environment; current != nil; current = current.parent {
		if result, ok := current.values[name]; ok {
			return result, true
		}
	}
	return nil, false
}

func (environment *environment) define(name string, result value) {
	environment.values[name] = result
}

func (environment *environment) assign(name string, result value) bool {
	for current := environment; current != nil; current = current.parent {
		if _, ok := current.values[name]; ok {
			current.values[name] = result
			return true
		}
	}
	return false
}

type signal int

const (
	signalNone signal = iota
	signalReturn
	signalBreak
	signalContinue
)

type execution struct {
	signal signal
	value  value
}

type runner struct {
	context context.Context
	program *ast.Program
	info    *checker.Info
	globals *environment
	current *environment
	options Options
	funcs   map[string]*ast.FuncDecl
}

// Run executes a successfully checked program until completion, cancellation,
// or a structured runtime failure.
func Run(ctx context.Context, program *ast.Program, info *checker.Info, options Options) error {
	if options.Stdout == nil {
		options.Stdout = io.Discard
	}
	if options.Files == nil {
		options.Files = osFileSystem{}
	}
	if ctx == nil {
		ctx = context.Background()
	}

	globals := &environment{values: map[string]value{}}
	runner := &runner{
		context: ctx,
		program: program,
		info:    info,
		globals: globals,
		current: globals,
		options: options,
		funcs:   map[string]*ast.FuncDecl{},
	}
	for _, declaration := range program.Decls {
		decl := declaration
		if exp, ok := declaration.(*ast.ExportStmt); ok {
			if innerDecl, ok := exp.Target.(ast.Decl); ok {
				decl = innerDecl
			}
		}
		if function, ok := decl.(*ast.FuncDecl); ok {
			runner.funcs[function.Name] = function
		}
	}
	for _, declaration := range program.Decls {
		decl := declaration
		if exp, ok := declaration.(*ast.ExportStmt); ok {
			if innerDecl, ok := exp.Target.(ast.Decl); ok {
				decl = innerDecl
			}
		}
		variable, ok := decl.(*ast.VarDecl)
		if !ok {
			continue
		}
		variableType, _ := info.GlobalType(variable.Name)
		result := runner.zero(variableType)
		if variable.Init != nil {
			var err error
			result, err = runner.eval(variable.Init)
			if err != nil {
				return err
			}
		}
		runner.globals.define(variable.Name, result)
	}

	_, err := runner.executeAll(program.Stmts)
	return err
}

func (runner *runner) executeAll(statements []ast.Stmt) (execution, error) {
	for _, statement := range statements {
		result, err := runner.execute(statement)
		if err != nil || result.signal != signalNone {
			return result, err
		}
	}
	return execution{}, nil
}

func (runner *runner) execute(statement ast.Stmt) (execution, error) {
	switch node := statement.(type) {
	case *ast.VarDecl:
		result := runner.zero(runner.info.TypeOfRef(node.Type))
		if node.Init != nil {
			var err error
			result, err = runner.eval(node.Init)
			if err != nil {
				return execution{}, err
			}
		}
		runner.current.define(node.Name, result)
	case *ast.BlockStmt:
		return runner.executeBlock(node)
	case *ast.ExprStmt:
		_, err := runner.eval(node.Expr)
		return execution{}, err
	case *ast.AssignStmt:
		store, err := runner.prepareStore(node.Target)
		if err != nil {
			return execution{}, err
		}
		result, err := runner.eval(node.Value)
		if err != nil {
			return execution{}, err
		}
		return execution{}, store(result)
	case *ast.IfStmt:
		condition, err := runner.eval(node.Cond)
		if err != nil {
			return execution{}, err
		}
		if bool(condition.(boolValue)) {
			return runner.executeBlock(node.Then)
		}
		if node.Else != nil {
			return runner.execute(node.Else)
		}
	case *ast.SwitchStmt:
		return runner.executeSwitch(node)
	case *ast.ForStmt:
		return runner.executeLoop(node)
	case *ast.ReturnStmt:
		var result value
		if node.Value != nil {
			var err error
			result, err = runner.eval(node.Value)
			if err != nil {
				return execution{}, err
			}
		}
		return execution{signal: signalReturn, value: result}, nil
	case *ast.BreakStmt:
		return execution{signal: signalBreak}, nil
	case *ast.ContinueStmt:
		return execution{signal: signalContinue}, nil
	case *ast.ExportStmt:
		if targetStmt, ok := node.Target.(ast.Stmt); ok {
			exec, err := runner.execute(targetStmt)
			if err != nil {
				return exec, err
			}
		}
		if varDecl, ok := node.Target.(*ast.VarDecl); ok {
			if val, ok := runner.current.get(varDecl.Name); ok {
				if runner.current.parent != nil {
					runner.current.parent.define(varDecl.Name, val)
				}
			}
		}
		return execution{}, nil
	case *ast.ImportStmt:
		return execution{}, nil
	}
	return execution{}, nil
}

func (runner *runner) executeSwitch(node *ast.SwitchStmt) (execution, error) {
	subject, err := runner.eval(node.Expr)
	if err != nil {
		return execution{}, err
	}
	var fallback *ast.CaseClause
	for _, clause := range node.Cases {
		if clause.Default {
			fallback = clause
			continue
		}
		for _, expression := range clause.Values {
			candidate, candidateErr := runner.eval(expression)
			if candidateErr != nil {
				return execution{}, candidateErr
			}
			if equal(subject, candidate) {
				return runner.executeScoped(clause.Body)
			}
		}
	}
	if fallback != nil {
		return runner.executeScoped(fallback.Body)
	}
	return execution{}, nil
}

func (runner *runner) executeBlock(block *ast.BlockStmt) (execution, error) {
	return runner.executeScoped(block.Stmts)
}

func (runner *runner) executeScoped(statements []ast.Stmt) (execution, error) {
	previous := runner.current
	runner.current = &environment{parent: previous, values: map[string]value{}}
	defer func() { runner.current = previous }()
	return runner.executeAll(statements)
}

func (runner *runner) executeLoop(loop *ast.ForStmt) (execution, error) {
	previous := runner.current
	runner.current = &environment{parent: previous, values: map[string]value{}}
	defer func() { runner.current = previous }()
	if loop.Init != nil {
		if _, err := runner.execute(loop.Init); err != nil {
			return execution{}, err
		}
	}
	for {
		if err := runner.context.Err(); err != nil {
			return execution{}, runner.runtime(loop, err.Error())
		}
		condition, err := runner.eval(loop.Cond)
		if err != nil {
			return execution{}, err
		}
		if !bool(condition.(boolValue)) {
			return execution{}, nil
		}
		result, err := runner.executeBlock(loop.Body)
		if err != nil {
			return execution{}, err
		}
		switch result.signal {
		case signalReturn:
			return result, nil
		case signalBreak:
			return execution{}, nil
		}
		if loop.Post != nil {
			if _, err = runner.execute(loop.Post); err != nil {
				return execution{}, err
			}
		}
	}
}

func (runner *runner) eval(expression ast.Expr) (value, error) {
	switch node := expression.(type) {
	case *ast.LiteralExpr:
		return primitive(node), nil
	case *ast.IdentExpr:
		result, _ := runner.current.get(node.Name)
		return result, nil
	case *ast.UnaryExpr:
		result, err := runner.eval(node.Right)
		if err != nil {
			return nil, err
		}
		if node.Op == "!" {
			return boolValue(!bool(result.(boolValue))), nil
		}
		return intValue(-int(result.(intValue))), nil
	case *ast.BinaryExpr:
		return runner.binary(node)
	case *ast.CallExpr:
		return runner.call(node)
	case *ast.FieldExpr:
		object, err := runner.eval(node.Object)
		if err != nil {
			return nil, err
		}
		if object == nil {
			return nil, runner.runtime(node, "field access on nil struct")
		}
		return object.(*structValue).fields[node.Name], nil
	case *ast.IndexExpr:
		return runner.index(node)
	case *ast.SliceExpr:
		return runner.slice(node)
	case *ast.MakeExpr:
		return runner.makeValue(node)
	case *ast.CompositeExpr:
		return runner.composite(node)
	default:
		return nil, runner.runtime(expression, "unsupported expression")
	}
}

func primitive(literal *ast.LiteralExpr) value {
	switch literal.Type {
	case ast.TypeInt:
		return intValue(literal.Value.(int))
	case ast.TypeChar:
		return charValue(literal.Value.(byte))
	case ast.TypeString:
		return stringValue(literal.Value.(string))
	case ast.TypeBool:
		return boolValue(literal.Value.(bool))
	default:
		return nil
	}
}

func (runner *runner) binary(node *ast.BinaryExpr) (value, error) {
	left, err := runner.eval(node.Left)
	if err != nil {
		return nil, err
	}
	if node.Op == "&&" && !bool(left.(boolValue)) {
		return boolValue(false), nil
	}
	if node.Op == "||" && bool(left.(boolValue)) {
		return boolValue(true), nil
	}
	right, err := runner.eval(node.Right)
	if err != nil {
		return nil, err
	}
	switch node.Op {
	case "+":
		switch left := left.(type) {
		case intValue:
			return intValue(int(left) + int(right.(intValue))), nil
		case stringValue:
			return stringValue(string(left) + string(right.(stringValue))), nil
		}
	case "-":
		return intValue(int(left.(intValue)) - int(right.(intValue))), nil
	case "*":
		return intValue(int(left.(intValue)) * int(right.(intValue))), nil
	case "/", "%":
		if right.(intValue) == 0 {
			return nil, runner.runtime(node, "division by zero")
		}
		if node.Op == "/" {
			return intValue(int(left.(intValue)) / int(right.(intValue))), nil
		}
		return intValue(int(left.(intValue)) % int(right.(intValue))), nil
	case "==":
		return boolValue(equal(left, right)), nil
	case "!=":
		return boolValue(!equal(left, right)), nil
	case "&&":
		return boolValue(bool(left.(boolValue)) && bool(right.(boolValue))), nil
	case "||":
		return boolValue(bool(left.(boolValue)) || bool(right.(boolValue))), nil
	case "<":
		return boolValue(compare(left, right) < 0), nil
	case "<=":
		return boolValue(compare(left, right) <= 0), nil
	case ">":
		return boolValue(compare(left, right) > 0), nil
	case ">=":
		return boolValue(compare(left, right) >= 0), nil
	}
	return nil, runner.runtime(node, "unknown operator")
}

func (runner *runner) index(node *ast.IndexExpr) (value, error) {
	object, err := runner.eval(node.Object)
	if err != nil {
		return nil, err
	}
	index, err := runner.eval(node.Index)
	if err != nil {
		return nil, err
	}
	switch object := object.(type) {
	case sliceValue:
		position := int(index.(intValue))
		if position < 0 || position >= len(object.items) {
			return nil, runner.runtime(node, "index out of bounds")
		}
		return object.items[position], nil
	case stringValue:
		position := int(index.(intValue))
		if position < 0 || position >= len(object) {
			return nil, runner.runtime(node, "index out of bounds")
		}
		return charValue(object[position]), nil
	case *mapValue:
		if result, ok := object.items[index]; ok {
			return result, nil
		}
		return runner.zero(object.element), nil
	default:
		return nil, runner.runtime(node, "value is not indexable")
	}
}

func (runner *runner) slice(node *ast.SliceExpr) (value, error) {
	object, err := runner.eval(node.Object)
	if err != nil {
		return nil, err
	}
	low, high := 0, -1
	if node.Low != nil {
		result, evalErr := runner.eval(node.Low)
		if evalErr != nil {
			return nil, evalErr
		}
		low = int(result.(intValue))
	}
	if node.High != nil {
		result, evalErr := runner.eval(node.High)
		if evalErr != nil {
			return nil, evalErr
		}
		high = int(result.(intValue))
	}
	switch object := object.(type) {
	case sliceValue:
		if high < 0 {
			high = len(object.items)
		}
		if low < 0 || high < low || high > len(object.items) {
			return nil, runner.runtime(node, "slice bounds out of range")
		}
		return sliceValue{element: object.element, items: object.items[low:high]}, nil
	case stringValue:
		if high < 0 {
			high = len(object)
		}
		if low < 0 || high < low || high > len(object) {
			return nil, runner.runtime(node, "slice bounds out of range")
		}
		return stringValue(object[low:high]), nil
	default:
		return nil, runner.runtime(node, "value is not sliceable")
	}
}

func (runner *runner) makeValue(node *ast.MakeExpr) (value, error) {
	resolved := runner.info.TypeOfRef(node.Type)
	if resolved.Kind == ast.TypeMap {
		return &mapValue{key: *resolved.Key, element: *resolved.Elem, items: map[value]value{}}, nil
	}
	length := 0
	if node.Size != nil {
		result, err := runner.eval(node.Size)
		if err != nil {
			return nil, err
		}
		length = int(result.(intValue))
		if length < 0 {
			return nil, runner.runtime(node, "negative slice size")
		}
	}
	items := make([]value, length)
	for index := range items {
		items[index] = runner.zero(*resolved.Elem)
	}
	return sliceValue{element: *resolved.Elem, items: items}, nil
}

func (runner *runner) composite(node *ast.CompositeExpr) (value, error) {
	resolved := runner.info.TypeOfRef(node.Type)
	switch resolved.Kind {
	case ast.TypeSlice:
		result := sliceValue{element: *resolved.Elem}
		for _, element := range node.Elems {
			item, err := runner.eval(element.Value)
			if err != nil {
				return nil, err
			}
			result.items = append(result.items, item)
		}
		return result, nil
	case ast.TypeMap:
		result := &mapValue{key: *resolved.Key, element: *resolved.Elem, items: map[value]value{}}
		for _, element := range node.Elems {
			key, err := runner.eval(element.KeyExpr)
			if err != nil {
				return nil, err
			}
			item, err := runner.eval(element.Value)
			if err != nil {
				return nil, err
			}
			result.items[key] = item
		}
		return result, nil
	case ast.TypeNamed:
		result := &structValue{name: resolved.Name, fields: map[string]value{}}
		for _, element := range node.Elems {
			item, err := runner.eval(element.Value)
			if err != nil {
				return nil, err
			}
			result.fields[element.Key] = item
		}
		return result, nil
	default:
		return nil, runner.runtime(node, "invalid composite")
	}
}

type storeFunc func(value) error

func (runner *runner) prepareStore(expression ast.Expr) (storeFunc, error) {
	switch node := expression.(type) {
	case *ast.IdentExpr:
		return func(result value) error {
			runner.current.assign(node.Name, result)
			return nil
		}, nil
	case *ast.FieldExpr:
		object, err := runner.eval(node.Object)
		if err != nil {
			return nil, err
		}
		if object == nil {
			return nil, runner.runtime(node, "field access on nil struct")
		}
		instance := object.(*structValue)
		return func(result value) error {
			instance.fields[node.Name] = result
			return nil
		}, nil
	case *ast.IndexExpr:
		object, err := runner.eval(node.Object)
		if err != nil {
			return nil, err
		}
		index, err := runner.eval(node.Index)
		if err != nil {
			return nil, err
		}
		switch object := object.(type) {
		case sliceValue:
			position := int(index.(intValue))
			if position < 0 || position >= len(object.items) {
				return nil, runner.runtime(node, "index out of bounds")
			}
			return func(result value) error {
				object.items[position] = result
				return nil
			}, nil
		case *mapValue:
			if object.items == nil {
				return nil, runner.runtime(node, "assignment to uninitialized map")
			}
			return func(result value) error {
				object.items[index] = result
				return nil
			}, nil
		}
	}
	return nil, runner.runtime(expression, "invalid assignment target")
}

func (runner *runner) call(node *ast.CallExpr) (value, error) {
	if err := runner.context.Err(); err != nil {
		return nil, runner.runtime(node, err.Error())
	}
	arguments := make([]value, 0, len(node.Args))
	for _, expression := range node.Args {
		argument, err := runner.eval(expression)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, argument)
	}
	if builtin, ok := runner.info.BuiltinOf(node); ok {
		return runner.callBuiltin(node, builtin, arguments)
	}

	function := runner.funcs[node.Callee]
	previous := runner.current
	runner.current = &environment{parent: runner.globals, values: map[string]value{}}
	defer func() { runner.current = previous }()
	for index, parameter := range function.Params {
		runner.current.define(parameter.Name, arguments[index])
	}
	result, err := runner.executeAll(function.Body.Stmts)
	if err != nil {
		return nil, err
	}
	return result.value, nil
}

func (runner *runner) callBuiltin(node *ast.CallExpr, builtin checker.Builtin, arguments []value) (value, error) {
	switch builtin {
	case checker.BuiltinPrint:
		parts := make([]string, len(arguments))
		for index, argument := range arguments {
			parts[index] = display(argument)
		}
		_, err := fmt.Fprintln(runner.options.Stdout, strings.Join(parts, " "))
		return nil, err
	case checker.BuiltinArgs:
		items := make([]value, len(runner.options.Args))
		for index, argument := range runner.options.Args {
			items[index] = stringValue(argument)
		}
		return sliceValue{element: checker.String, items: items}, nil
	case checker.BuiltinReadFile:
		data, err := runner.options.Files.ReadFile(string(arguments[0].(stringValue)))
		if err != nil {
			return nil, runner.runtime(node, err.Error())
		}
		return stringValue(data), nil
	case checker.BuiltinWriteFile:
		err := runner.options.Files.WriteFile(string(arguments[0].(stringValue)), []byte(arguments[1].(stringValue)), 0o644)
		if err != nil {
			return nil, runner.runtime(node, err.Error())
		}
		return nil, nil
	case checker.BuiltinFail:
		return nil, runner.runtime(node, string(arguments[0].(stringValue)))
	case checker.BuiltinLen:
		switch argument := arguments[0].(type) {
		case stringValue:
			return intValue(len(argument)), nil
		case sliceValue:
			return intValue(len(argument.items)), nil
		case *mapValue:
			return intValue(len(argument.items)), nil
		}
	case checker.BuiltinAppend:
		slice := arguments[0].(sliceValue)
		return sliceValue{element: slice.element, items: append(slice.items, arguments[1])}, nil
	case checker.BuiltinInt:
		switch argument := arguments[0].(type) {
		case charValue:
			return intValue(argument), nil
		case stringValue:
			result, err := strconv.Atoi(string(argument))
			if err != nil {
				return nil, runner.runtime(node, "invalid integer conversion")
			}
			return intValue(result), nil
		}
	case checker.BuiltinChar:
		integer := int(arguments[0].(intValue))
		if integer < 0 || integer > 255 {
			return nil, runner.runtime(node, "invalid character conversion")
		}
		return charValue(integer), nil
	case checker.BuiltinString:
		switch argument := arguments[0].(type) {
		case intValue:
			return stringValue(strconv.Itoa(int(argument))), nil
		case charValue:
			return stringValue([]byte{byte(argument)}), nil
		case boolValue:
			return stringValue(strconv.FormatBool(bool(argument))), nil
		}
	}
	return nil, runner.runtime(node, "unsupported built-in")
}

func (runner *runner) zero(resolved checker.Type) value {
	switch resolved.Kind {
	case ast.TypeInt:
		return intValue(0)
	case ast.TypeChar:
		return charValue(0)
	case ast.TypeString:
		return stringValue("")
	case ast.TypeBool:
		return boolValue(false)
	case ast.TypeSlice:
		return sliceValue{element: *resolved.Elem}
	case ast.TypeMap:
		return &mapValue{key: *resolved.Key, element: *resolved.Elem}
	default:
		return nil
	}
}

func (runner *runner) runtime(node ast.Node, message string) error {
	return &RuntimeError{Span: node.GetSpan(), Message: message}
}

func equal(left, right value) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	switch left := left.(type) {
	case intValue:
		return left == right.(intValue)
	case charValue:
		return left == right.(charValue)
	case stringValue:
		return left == right.(stringValue)
	case boolValue:
		return left == right.(boolValue)
	case *structValue:
		return left == right
	default:
		return false
	}
}

func compare(left, right value) int {
	switch left := left.(type) {
	case intValue:
		right := right.(intValue)
		if left < right {
			return -1
		}
		if left > right {
			return 1
		}
	case charValue:
		right := right.(charValue)
		if left < right {
			return -1
		}
		if left > right {
			return 1
		}
	case stringValue:
		right := right.(stringValue)
		if left < right {
			return -1
		}
		if left > right {
			return 1
		}
	}
	return 0
}

func display(result value) string {
	switch result := result.(type) {
	case nil:
		return "<void>"
	case intValue:
		return strconv.Itoa(int(result))
	case charValue:
		return string([]byte{byte(result)})
	case stringValue:
		return string(result)
	case boolValue:
		return strconv.FormatBool(bool(result))
	case sliceValue:
		parts := make([]string, len(result.items))
		for index, item := range result.items {
			parts[index] = display(item)
		}
		return "[" + strings.Join(parts, " ") + "]"
	case *mapValue:
		return fmt.Sprintf("map[%d entries]", len(result.items))
	case *structValue:
		return result.name
	default:
		return "<invalid>"
	}
}
