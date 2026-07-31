// Package checker resolves names and statically checks Nuru programs.
package checker

import (
	"fmt"

	"github.com/kc-clintone/compilers/internal/ast"
	"github.com/kc-clintone/compilers/internal/diagnostic"
	"github.com/kc-clintone/compilers/internal/source"
)

// Type is a resolved Nuru type.
type Type struct {
	Kind      ast.TypeKind
	Name      string
	Elem, Key *Type
}

// Canonical primitive, void, and invalid checker types.
var (
	Invalid = Type{Kind: ast.TypeInvalid}
	Void    = Type{Kind: ast.TypeVoid}
	Int     = Type{Kind: ast.TypeInt}
	Char    = Type{Kind: ast.TypeChar}
	String  = Type{Kind: ast.TypeString}
	Bool    = Type{Kind: ast.TypeBool}
)

// SymbolKind identifies the namespace role of a resolved name.
type SymbolKind int

// Resolved symbol kinds.
const (
	SymbolVariable SymbolKind = iota
	SymbolParameter
	SymbolFunction
	SymbolField
	SymbolBuiltin
)

// Symbol describes a resolved value name.
type Symbol struct {
	Kind SymbolKind
	Name string
	Type Type
	Decl ast.Node
}

// Builtin identifies a statically selected built-in operation.
type Builtin string

// Built-in operation identifiers.
const (
	BuiltinNone      Builtin = ""
	BuiltinPrint     Builtin = "print"
	BuiltinArgs      Builtin = "args"
	BuiltinReadFile  Builtin = "readFile"
	BuiltinWriteFile Builtin = "writeFile"
	BuiltinFail      Builtin = "fail"
	BuiltinLen       Builtin = "len"
	BuiltinAppend    Builtin = "append"
	BuiltinInt       Builtin = "int"
	BuiltinChar      Builtin = "char"
	BuiltinString    Builtin = "string"
)

// String returns the Nuru spelling of t.
func (t Type) String() string {
	switch t.Kind {
	case ast.TypeInt:
		return "int"
	case ast.TypeChar:
		return "char"
	case ast.TypeString:
		return "string"
	case ast.TypeBool:
		return "bool"
	case ast.TypeVoid:
		return "void"
	case ast.TypeSlice:
		return "[]" + t.Elem.String()
	case ast.TypeMap:
		return "map[" + t.Key.String() + "]" + t.Elem.String()
	case ast.TypeNamed:
		return t.Name
	}

	return "<invalid>"
}

// Equal reports whether t and u are identical Nuru types. Invalid is treated
// as compatible to suppress cascading diagnostics.
func (t Type) Equal(u Type) bool {
	if t.Kind == ast.TypeInvalid || u.Kind == ast.TypeInvalid {
		return true
	}

	if t.Kind != u.Kind || t.Name != u.Name {
		return false
	}

	if t.Kind == ast.TypeSlice {
		return t.Elem.Equal(*u.Elem)
	}

	if t.Kind == ast.TypeMap {
		return t.Key.Equal(*u.Key) && t.Elem.Equal(*u.Elem)
	}

	return true
}

// Func is a resolved user-function signature.
type Func struct {
	Name   string
	Params []Type
	Result Type
	Decl   *ast.FuncDecl
}

// Struct is a resolved named-struct declaration.
type Struct struct {
	Name   string
	Decl   *ast.StructDecl
	fields map[string]Type
}

// Info is immutable semantic information consumed by both back ends.
// Info is immutable semantic information consumed by both back ends.
type Info struct {
	exprTypes map[ast.NodeID]Type
	symbols   map[ast.NodeID]*Symbol
	structsAt map[ast.NodeID]*Struct
	builtins  map[ast.NodeID]Builtin
	structs   map[string]*Struct
	functions map[string]*Func
	globals   map[string]Type
}

// TypeOf returns the checked type of e, or Invalid when e was not typed.
func (i *Info) TypeOf(e ast.Expr) Type {
	if t, ok := i.exprTypes[e.GetID()]; ok {
		return t
	}

	return Invalid
}

// SymbolOf returns the symbol resolved for an identifier expression.
func (i *Info) SymbolOf(e *ast.IdentExpr) (*Symbol, bool) {
	symbol, ok := i.symbols[e.GetID()]
	return symbol, ok
}

// StructOf returns the struct declaration resolved at a named type or
// constructor node.
func (i *Info) StructOf(node ast.Node) (*Struct, bool) {
	structure, ok := i.structsAt[node.GetID()]
	return structure, ok
}

// BuiltinOf returns the built-in operation selected for call.
func (i *Info) BuiltinOf(call *ast.CallExpr) (Builtin, bool) {
	builtin, ok := i.builtins[call.GetID()]
	return builtin, ok
}

// GlobalType returns the declared type of a global variable.
func (i *Info) GlobalType(name string) (Type, bool) {
	t, ok := i.globals[name]
	return t, ok
}

// TypeOfRef returns the resolved type represented by ref. Named references are
// valid because Info is only produced after successful checking.
func (i *Info) TypeOfRef(ref *ast.TypeRef) Type {
	if ref == nil {
		return Void
	}
	switch ref.Kind {
	case ast.TypeInt:
		return Int
	case ast.TypeChar:
		return Char
	case ast.TypeString:
		return String
	case ast.TypeBool:
		return Bool
	case ast.TypeNamed:
		return Type{Kind: ast.TypeNamed, Name: ref.Name}
	case ast.TypeSlice:
		element := i.TypeOfRef(ref.Elem)
		return Type{Kind: ast.TypeSlice, Elem: &element}
	case ast.TypeMap:
		key, element := i.TypeOfRef(ref.Key), i.TypeOfRef(ref.Elem)
		return Type{Kind: ast.TypeMap, Key: &key, Elem: &element}
	default:
		return Invalid
	}
}

type scope struct {
	parent *scope
	values map[string]*Symbol
}

func (s *scope) get(n string) (*Symbol, bool) {
	for x := s; x != nil; x = x.parent {
		if t, ok := x.values[n]; ok {
			return t, true
		}
	}

	return nil, false
}
func (s *scope) put(symbol *Symbol) bool {
	n := symbol.Name
	if _, ok := s.values[n]; ok {
		return false
	}

	s.values[n] = symbol
	return true
}

// Checker holds the mutable state of one semantic check.
type Checker struct {
	info      *Info
	diags     []diagnostic.Diagnostic
	scope     *scope
	current   *Func
	loopDepth int
}

// Check resolves and validates program. Info is returned only when no
// diagnostics were produced.
func Check(program *ast.Program) (*Info, []diagnostic.Diagnostic) {
	c := &Checker{info: &Info{exprTypes: map[ast.NodeID]Type{}, symbols: map[ast.NodeID]*Symbol{}, structsAt: map[ast.NodeID]*Struct{}, builtins: map[ast.NodeID]Builtin{}, structs: map[string]*Struct{}, functions: map[string]*Func{}, globals: map[string]Type{}}}

	c.scope = &scope{values: map[string]*Symbol{}}
	c.declare(program)
	c.defineStructs(program)
	c.checkProgram(program)
	diagnostic.Sort(c.diags)
	if len(c.diags) > 0 {
		return nil, c.diags
	}

	return c.info, nil
}

func unwrapDecl(d ast.Decl) ast.Decl {
	if exp, ok := d.(*ast.ExportStmt); ok {
		if inner, ok := exp.Target.(ast.Decl); ok {
			return unwrapDecl(inner)
		}
	}
	return d
}

func (c *Checker) declare(p *ast.Program) {
	for _, d := range p.Decls {
		if s, ok := unwrapDecl(d).(*ast.StructDecl); ok {
			if s.Name == "main" {
				c.err(s, "reserved name main")
			}
			if _, exists := c.info.structs[s.Name]; exists {
				c.err(s, "duplicate type "+s.Name)
			} else {
				c.info.structs[s.Name] = &Struct{Name: s.Name, Decl: s, fields: map[string]Type{}}
			}
		}
	}

	for _, d := range p.Decls {
		if x, ok := unwrapDecl(d).(*ast.VarDecl); ok {
			if x.Name == "main" {
				c.err(x, "reserved name main")
			}
			if builtin(x.Name) {
				c.err(x, "reserved built-in name "+x.Name)
			}
			t := c.resolveType(x.Type)

			symbol := &Symbol{Kind: SymbolVariable, Name: x.Name, Type: t, Decl: x}
			if !c.scope.put(symbol) {
				c.err(x, "duplicate global "+x.Name)
			} else {
				c.info.globals[x.Name] = t
			}
		}
	}

	for _, d := range p.Decls {
		if x, ok := unwrapDecl(d).(*ast.FuncDecl); ok {
			if x.Name == "main" {
				c.err(x, "reserved name main")
				continue
			}
			if builtin(x.Name) {
				c.err(x, "reserved built-in name "+x.Name)
				continue
			}
			if _, exists := c.scope.get(x.Name); exists {
				c.err(x, "duplicate value name "+x.Name)
				continue
			}

			if _, ok := c.info.functions[x.Name]; ok {
				c.err(x, "duplicate function "+x.Name)
				continue
			}

			var ps []Type

			for _, p := range x.Params {
				ps = append(ps, c.resolveType(p.Type))
			}

			r := Void

			if x.Result != nil {
				r = c.resolveType(x.Result)
			}

			c.info.functions[x.Name] = &Func{Name: x.Name, Params: ps, Result: r, Decl: x}
		}
	}
}
func (c *Checker) defineStructs(p *ast.Program) {
	for _, d := range p.Decls {
		if x, ok := unwrapDecl(d).(*ast.StructDecl); ok {
			s := c.info.structs[x.Name]

			for _, f := range x.Fields {
				if f.Name == "main" {
					c.errSpan(f.Span, "reserved name main")
				}
				if _, exists := s.fields[f.Name]; exists {
					c.errSpan(f.Span, "duplicate field "+f.Name)
				} else {
					s.fields[f.Name] = c.resolveType(f.Type)
				}
			}
		}
	}
}
func (c *Checker) checkProgram(p *ast.Program) {
	for _, d := range p.Decls {
		switch x := unwrapDecl(d).(type) {
		case *ast.VarDecl:
			if x.Init != nil {
				c.expect(x.Init, c.info.globals[x.Name], "initializer")
			}

			if c.info.globals[x.Name].Kind == ast.TypeNamed && x.Init == nil {
				c.err(x, "struct variable requires an initializer")
			}
		case *ast.FuncDecl:
			c.checkFunc(x)
		}
	}

	for _, s := range p.Stmts {
		c.stmt(s)
	}
}
func (c *Checker) checkFunc(d *ast.FuncDecl) {
	f := c.info.functions[d.Name]

	if f == nil {
		return
	}

	oldScope, oldFn := c.scope, c.current

	c.scope = &scope{parent: oldScope, values: map[string]*Symbol{}}
	c.current = f
	for i, p := range d.Params {
		if p.Name == "main" || builtin(p.Name) {
			c.errSpan(p.Span, "reserved name "+p.Name)
		}
		symbol := &Symbol{Kind: SymbolParameter, Name: p.Name, Type: f.Params[i]}
		if !c.scope.put(symbol) {
			c.errSpan(p.Span, "duplicate parameter "+p.Name)
		}
	}

	c.block(d.Body, false)
	if f.Result.Kind != ast.TypeVoid && !guaranteesReturn(d.Body) {
		c.err(d, "function "+d.Name+" does not return on every path")
	}

	c.scope, c.current = oldScope, oldFn
}

func (c *Checker) stmt(s ast.Stmt) {
	switch x := s.(type) {
	case *ast.VarDecl:
		t := c.resolveType(x.Type)
		if x.Name == "main" || builtin(x.Name) {
			c.err(x, "reserved name "+x.Name)
		}

		if x.Init != nil {
			c.expect(x.Init, t, "initializer")
		}

		if t.Kind == ast.TypeNamed && x.Init == nil {
			c.err(x, "struct variable requires an initializer")
		}

		if !c.scope.put(&Symbol{Kind: SymbolVariable, Name: x.Name, Type: t, Decl: x}) {
			c.err(x, "duplicate variable "+x.Name)
		}
	case *ast.BlockStmt:
		c.block(x, true)
	case *ast.ExprStmt:
		c.expr(x.Expr)
	case *ast.AssignStmt:
		lt := c.lvalue(x.Target)
		rt := c.expr(x.Value)

		if !lt.Equal(rt) {
			c.err(x, "cannot assign "+rt.String()+" to "+lt.String())
		}
	case *ast.IfStmt:
		c.require(c.expr(x.Cond), Bool, x.Cond, "if condition")
		c.block(x.Then, true)
		if x.Else != nil {
			c.stmt(x.Else)
		}
	case *ast.SwitchStmt:
		subject := c.expr(x.Expr)
		defaults := 0
		seenCases := map[string]bool{}

		for _, cl := range x.Cases {
			if cl.Default {
				defaults++
			}

			for _, v := range cl.Values {
				c.require(c.expr(v), subject, v, "case value")
				if key, ok := literalKey(v); ok {
					if seenCases[key] {
						c.err(v, "duplicate switch case")
					}
					seenCases[key] = true
				}
			}

			c.blockStmts(cl.Body, true)
		}

		if defaults > 1 {
			c.err(x, "switch has multiple default clauses")
		}
	case *ast.ForStmt:
		old := c.scope

		c.scope = &scope{parent: old, values: map[string]*Symbol{}}
		if x.Init != nil {
			c.stmt(x.Init)
		}

		c.require(c.expr(x.Cond), Bool, x.Cond, "loop condition")
		c.loopDepth++
		c.block(x.Body, true)
		if x.Post != nil {
			c.stmt(x.Post)
		}

		c.loopDepth--
		c.scope = old
	case *ast.BreakStmt:
		if c.loopDepth == 0 {
			c.err(x, "break is only valid inside a loop")
		}
	case *ast.ContinueStmt:
		if c.loopDepth == 0 {
			c.err(x, "continue is only valid inside a loop")
		}
	case *ast.ReturnStmt:
		if c.current == nil {
			c.err(x, "return is only valid inside a function")
			return
		}

		if x.Value == nil {
			if c.current.Result.Kind != ast.TypeVoid {
				c.err(x, "return value required")
			}
		} else {
			if c.current.Result.Kind == ast.TypeVoid {
				c.err(x, "void function cannot return a value")
			} else {
				c.require(c.expr(x.Value), c.current.Result, x.Value, "return value")
			}
		}
	case *ast.ExportStmt:
		if targetStmt, ok := x.Target.(ast.Stmt); ok {
			c.stmt(targetStmt)
		}
		if varDecl, ok := x.Target.(*ast.VarDecl); ok {
			if sym, ok := c.scope.values[varDecl.Name]; ok && c.scope.parent != nil {
				c.scope.parent.put(sym)
			}
		}
	case *ast.ImportStmt:
		// Module scaffolding (no-op in checker)
	}
}
func (c *Checker) block(b *ast.BlockStmt, nested bool) { c.blockStmts(b.Stmts, nested) }
func (c *Checker) blockStmts(ss []ast.Stmt, nested bool) {
	old := c.scope

	if nested {
		c.scope = &scope{parent: old, values: map[string]*Symbol{}}
	}

	for _, s := range ss {
		c.stmt(s)
	}

	if nested {
		c.scope = old
	}
}

func (c *Checker) expr(e ast.Expr) Type {
	if e == nil {
		return Void
	}

	var t Type

	switch x := e.(type) {
	case *ast.LiteralExpr:
		t = Type{Kind: x.Type}
	case *ast.IdentExpr:
		var ok bool

		symbol, ok := c.scope.get(x.Name)
		if !ok {
			c.err(x, "unknown variable "+x.Name)
			t = Invalid
		} else {
			t = symbol.Type
			c.info.symbols[x.GetID()] = symbol
		}
	case *ast.UnaryExpr:
		r := c.expr(x.Right)

		if x.Op == "!" {
			c.require(r, Bool, x.Right, "operand")
			t = Bool
		} else {
			c.require(r, Int, x.Right, "operand")
			t = Int
		}
	case *ast.BinaryExpr:
		t = c.binary(x)
	case *ast.CallExpr:
		t = c.call(x)
	case *ast.FieldExpr:
		o := c.expr(x.Object)

		if o.Kind != ast.TypeNamed {
			c.err(x, "field access requires a struct")
			t = Invalid
		} else if s := c.info.structs[o.Name]; s == nil {
			t = Invalid
		} else if ft, ok := s.fields[x.Name]; ok {
			t = ft
		} else {
			c.err(x, "unknown field "+x.Name)
			t = Invalid
		}
	case *ast.IndexExpr:
		o := c.expr(x.Object)
		idx := c.expr(x.Index)

		switch o.Kind {
		case ast.TypeSlice:
			c.require(idx, Int, x.Index, "index")
			t = *o.Elem
		case ast.TypeString:
			c.require(idx, Int, x.Index, "index")
			t = Char
		case ast.TypeMap:
			c.require(idx, *o.Key, x.Index, "map key")
			t = *o.Elem
		default:
			c.err(x, "value is not indexable")
			t = Invalid
		}
	case *ast.SliceExpr:
		o := c.expr(x.Object)

		if x.Low != nil {
			c.require(c.expr(x.Low), Int, x.Low, "slice bound")
		}

		if x.High != nil {
			c.require(c.expr(x.High), Int, x.High, "slice bound")
		}

		if o.Kind == ast.TypeSlice {
			t = o
		} else if o.Kind == ast.TypeString {
			t = String
		} else {
			c.err(x, "value is not sliceable")
			t = Invalid
		}
	case *ast.MakeExpr:
		t = c.resolveType(x.Type)
		if t.Kind != ast.TypeSlice && t.Kind != ast.TypeMap {
			c.err(x, "make requires a slice or map type")
		}

		if x.Size != nil {
			if t.Kind != ast.TypeSlice {
				c.err(x, "only slices accept a size")
			}

			c.require(c.expr(x.Size), Int, x.Size, "make size")
		}
	case *ast.CompositeExpr:
		t = c.composite(x)
	default:
		t = Invalid
	}

	c.info.exprTypes[e.GetID()] = t
	return t
}
func (c *Checker) binary(x *ast.BinaryExpr) Type {
	l, r := c.expr(x.Left), c.expr(x.Right)

	if !l.Equal(r) {
		c.err(x, "operator operands have different types")
		return Invalid
	}

	switch x.Op {
	case "+":
		if l.Kind == ast.TypeInt || l.Kind == ast.TypeString {
			return l
		}
	case "-", "*", "/", "%":
		if l.Kind == ast.TypeInt {
			return Int
		}
	case "<", "<=", ">", ">=":
		if l.Kind == ast.TypeInt || l.Kind == ast.TypeChar || l.Kind == ast.TypeString {
			return Bool
		}
	case "==", "!=":
		if l.Kind == ast.TypeInt || l.Kind == ast.TypeChar || l.Kind == ast.TypeString || l.Kind == ast.TypeBool || l.Kind == ast.TypeNamed {
			return Bool
		}
	case "&&", "||":
		if l.Kind == ast.TypeBool {
			return Bool
		}
	}

	c.err(x, "operator "+x.Op+" does not accept "+l.String())
	return Invalid
}
func (c *Checker) call(x *ast.CallExpr) Type {
	if f := c.info.functions[x.Callee]; f != nil {
		if len(x.Args) != len(f.Params) {
			c.err(x, fmt.Sprintf("%s expects %d arguments", x.Callee, len(f.Params)))
		}

		for i, a := range x.Args {
			got := c.expr(a)

			if i < len(f.Params) {
				c.require(got, f.Params[i], a, "argument")
			}
		}

		return f.Result
	}
	if id := builtinID(x.Callee); id != BuiltinNone {
		c.info.builtins[x.GetID()] = id
	}

	switch x.Callee {
	case "print":
		for _, a := range x.Args {
			c.expr(a)
		}

		return Void
	case "args":
		c.arity(x, 0)
		e := String

		return Type{Kind: ast.TypeSlice, Elem: &e}
	case "readFile":
		c.args(x, String)
		return String
	case "writeFile":
		c.args(x, String, String)
		return Void
	case "fail":
		c.args(x, String)
		return Void
	case "len":
		if len(x.Args) != 1 {
			c.arity(x, 1)
			return Invalid
		}

		a := c.expr(x.Args[0])

		if a.Kind != ast.TypeString && a.Kind != ast.TypeSlice && a.Kind != ast.TypeMap {
			c.err(x, "len requires string, slice, or map")
		}

		return Int
	case "append":
		if len(x.Args) != 2 {
			c.arity(x, 2)
			return Invalid
		}

		s := c.expr(x.Args[0])
		v := c.expr(x.Args[1])

		if s.Kind != ast.TypeSlice {
			c.err(x, "append requires a slice")
			return Invalid
		}

		c.require(v, *s.Elem, x.Args[1], "append value")
		return s
	case "int":
		if len(x.Args) != 1 {
			c.arity(x, 1)
			return Invalid
		}

		a := c.expr(x.Args[0])

		if a.Kind != ast.TypeChar && a.Kind != ast.TypeString {
			c.err(x, "int conversion requires char or string")
		}

		return Int
	case "char":
		c.args(x, Int)
		return Char
	case "string":
		if len(x.Args) != 1 {
			c.arity(x, 1)
			return Invalid
		}

		a := c.expr(x.Args[0])

		if a.Kind != ast.TypeInt && a.Kind != ast.TypeChar && a.Kind != ast.TypeBool {
			c.err(x, "string conversion requires int, char, or bool")
		}

		return String
	}

	c.err(x, "unknown function "+x.Callee)
	for _, a := range x.Args {
		c.expr(a)
	}

	return Invalid
}
func (c *Checker) composite(x *ast.CompositeExpr) Type {
	t := c.resolveType(x.Type)

	switch t.Kind {
	case ast.TypeNamed:
		s := c.info.structs[t.Name]
		if s == nil {
			return Invalid
		}
		c.info.structsAt[x.GetID()] = s
		seen := map[string]bool{}

		for _, e := range x.Elems {
			ft, ok := s.fields[e.Key]

			if !ok {
				c.errSpan(e.Span, "unknown field "+e.Key)
			} else {
				c.require(c.expr(e.Value), ft, e.Value, "field")
			}

			if seen[e.Key] {
				c.errSpan(e.Span, "duplicate field "+e.Key)
			}

			seen[e.Key] = true
		}

		for n := range s.fields {
			if !seen[n] {
				c.err(x, "missing field "+n)
			}
		}
	case ast.TypeSlice:
		for _, e := range x.Elems {
			if e.KeyExpr != nil {
				c.errSpan(e.Span, "slice literal cannot have keys")
			}

			c.require(c.expr(e.Value), *t.Elem, e.Value, "slice element")
		}
	case ast.TypeMap:
		for _, e := range x.Elems {
			if e.KeyExpr == nil {
				c.errSpan(e.Span, "map element requires key")
			} else {
				c.require(c.expr(e.KeyExpr), *t.Key, e.KeyExpr, "map key")
			}

			c.require(c.expr(e.Value), *t.Elem, e.Value, "map value")
		}
	default:
		c.err(x, "invalid composite literal type")
	}

	return t
}
func (c *Checker) lvalue(e ast.Expr) Type {
	switch e.(type) {
	case *ast.IdentExpr, *ast.FieldExpr, *ast.IndexExpr:
		return c.expr(e)
	}

	c.err(e, "expression is not assignable")
	return Invalid
}

func (c *Checker) resolveType(r *ast.TypeRef) Type {
	if r == nil {
		return Void
	}

	switch r.Kind {
	case ast.TypeInt:
		return Int
	case ast.TypeChar:
		return Char
	case ast.TypeString:
		return String
	case ast.TypeBool:
		return Bool
	case ast.TypeSlice:
		e := c.resolveType(r.Elem)

		return Type{Kind: ast.TypeSlice, Elem: &e}
	case ast.TypeMap:
		k, e := c.resolveType(r.Key), c.resolveType(r.Elem)

		if k.Kind != ast.TypeInt && k.Kind != ast.TypeChar && k.Kind != ast.TypeString && k.Kind != ast.TypeBool {
			c.err(r, "invalid map key type")
		}

		return Type{Kind: ast.TypeMap, Key: &k, Elem: &e}
	case ast.TypeNamed:
		structure, ok := c.info.structs[r.Name]
		if !ok {
			c.err(r, "unknown type "+r.Name)
			return Invalid
		}
		c.info.structsAt[r.GetID()] = structure

		return Type{Kind: ast.TypeNamed, Name: r.Name}
	}

	return Invalid
}

func literalKey(expr ast.Expr) (string, bool) {
	literal, ok := expr.(*ast.LiteralExpr)
	if !ok {
		return "", false
	}

	return fmt.Sprintf("%d:%v", literal.Type, literal.Value), true
}

func builtinID(name string) Builtin {
	switch name {
	case "print":
		return BuiltinPrint
	case "args":
		return BuiltinArgs
	case "readFile":
		return BuiltinReadFile
	case "writeFile":
		return BuiltinWriteFile
	case "fail":
		return BuiltinFail
	case "len":
		return BuiltinLen
	case "append":
		return BuiltinAppend
	case "int":
		return BuiltinInt
	case "char":
		return BuiltinChar
	case "string":
		return BuiltinString
	default:
		return BuiltinNone
	}
}
func (c *Checker) expect(e ast.Expr, want Type, what string) { c.require(c.expr(e), want, e, what) }
func (c *Checker) require(got, want Type, n ast.Node, what string) {
	if !got.Equal(want) {
		c.err(n, what+" has type "+got.String()+", want "+want.String())
	}
}
func (c *Checker) args(x *ast.CallExpr, wants ...Type) {
	if len(x.Args) != len(wants) {
		c.arity(x, len(wants))
	}

	for i, a := range x.Args {
		got := c.expr(a)

		if i < len(wants) {
			c.require(got, wants[i], a, "argument")
		}
	}
}
func (c *Checker) arity(x *ast.CallExpr, n int) {
	if len(x.Args) != n {
		c.err(x, fmt.Sprintf("%s expects %d arguments", x.Callee, n))
	}
}
func (c *Checker) err(n ast.Node, msg string) { c.errSpan(n.GetSpan(), msg) }
func (c *Checker) errSpan(s source.Span, msg string) {
	c.diags = append(c.diags, diagnostic.Diagnostic{Span: s, Phase: "checker", Message: msg})
}
func builtin(s string) bool {
	switch s {
	case "print", "args", "readFile", "writeFile", "len", "append", "int", "char", "string", "fail":
		return true
	}

	return false
}
func guaranteesReturn(b *ast.BlockStmt) bool {
	for _, s := range b.Stmts {
		switch x := s.(type) {
		case *ast.ReturnStmt:
			return true
		case *ast.IfStmt:
			if x.Else != nil && guaranteeStmt(x.Then) && guaranteeStmt(x.Else) {
				return true
			}
		case *ast.SwitchStmt:
			hasDefault := false
			all := len(x.Cases) > 0

			for _, cl := range x.Cases {
				hasDefault = hasDefault || cl.Default
				all = all && guaranteesReturn(&ast.BlockStmt{Stmts: cl.Body})
			}

			if hasDefault && all {
				return true
			}
		}
	}

	return false
}
func guaranteeStmt(s ast.Stmt) bool {
	switch x := s.(type) {
	case *ast.BlockStmt:
		return guaranteesReturn(x)
	case *ast.IfStmt:
		return x.Else != nil && guaranteeStmt(x.Then) && guaranteeStmt(x.Else)
	}

	return false
}
