// Package compiler translates checked Nuru programs to readable Go.
package compiler

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/kc-clintone/compilers/internal/ast"
	"github.com/kc-clintone/compilers/internal/checker"
)

type generator struct {
	body    bytes.Buffer
	indent  int
	info    *checker.Info
	imports map[string]bool
	helpers map[string]bool
}

// Generate translates a checked Nuru program into formatted Go source.
func Generate(p *ast.Program, info *checker.Info) ([]byte, error) {
	g := &generator{info: info, imports: map[string]bool{}, helpers: map[string]bool{}}

	for _, d := range p.Decls {
		g.decl(d)
		g.line("")
	}

	g.line("func main() {")
	g.indent++
	for _, s := range p.Stmts {
		g.stmt(s)
	}

	g.indent--
	g.line("}")
	g.emitHelpers()
	var out bytes.Buffer

	out.WriteString("package main\n\n")
	if len(g.imports) > 0 {
		var names []string

		for n := range g.imports {
			names = append(names, n)
		}

		sort.Strings(names)
		out.WriteString("import (\n")
		for _, n := range names {
			fmt.Fprintf(&out, "\t%q\n", n)
		}

		out.WriteString(")\n\n")
	}

	out.Write(g.body.Bytes())
	formatted, err := format.Source(out.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated Go: %w\n%s", err, out.String())
	}

	return formatted, nil
}

// Build compiles generated Go source to output using an isolated temporary
// source directory and the caller's configured Go build cache.
func Build(ctx context.Context, src []byte, output string) error {
	dir, err := os.MkdirTemp("", "nuru-build-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	input := filepath.Join(dir, "main.go")

	if err = os.WriteFile(input, src, 0o600); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "go", "build", "-o", output, input)

	b, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build failed: %w\n%s", err, b)
	}

	return nil
}

func (g *generator) decl(d ast.Decl) {
	switch x := d.(type) {
	case *ast.StructDecl:
		g.line("type " + name(x.Name) + " struct {")
		g.indent++
		for _, f := range x.Fields {
			g.line(name(f.Name) + " " + g.typ(f.Type))
		}

		g.indent--
		g.line("}")
	case *ast.VarDecl:
		s := "var " + name(x.Name) + " " + g.typ(x.Type)

		if x.Init != nil {
			s += " = " + g.expr(x.Init)
		}

		g.line(s)
	case *ast.FuncDecl:
		var ps []string

		for _, p := range x.Params {
			ps = append(ps, name(p.Name)+" "+g.typ(p.Type))
		}

		s := "func " + name(x.Name) + "(" + strings.Join(ps, ", ") + ")"

		if x.Result != nil {
			s += " " + g.typ(x.Result)
		}

		g.rawBlockStart(s)
		for _, st := range x.Body.Stmts {
			g.stmt(st)
		}

		g.rawBlockEnd()
	case *ast.ExportStmt:
		if decl, ok := x.Target.(ast.Decl); ok {
			g.decl(decl)
		} else if stmt, ok := x.Target.(ast.Stmt); ok {
			g.stmt(stmt)
		}
	case *ast.ImportStmt:
		// No-op for imports
	}
}

func (g *generator) stmt(s ast.Stmt) {
	switch x := s.(type) {
	case *ast.VarDecl:
		v := "var " + name(x.Name) + " " + g.typ(x.Type)

		if x.Init != nil {
			v += " = " + g.expr(x.Init)
		}

		g.line(v)
	case *ast.ExprStmt:
		g.line(g.expr(x.Expr))
	case *ast.AssignStmt:
		if index, ok := x.Target.(*ast.IndexExpr); ok {
			switch g.info.TypeOf(index.Object).Kind {
			case ast.TypeSlice:
				g.helpers["setSlice"] = true
				g.line("nuruSetSlice(" + g.expr(index.Object) + ", " + g.expr(index.Index) + ", " + g.expr(x.Value) + ", " + location(index) + ")")
			case ast.TypeMap:
				g.helpers["setMap"] = true
				g.line("nuruSetMap(" + g.expr(index.Object) + ", " + g.expr(index.Index) + ", " + g.expr(x.Value) + ", " + location(index) + ")")
			default:
				g.line(g.expr(x.Target) + " = " + g.expr(x.Value))
			}
		} else {
			g.line(g.expr(x.Target) + " = " + g.expr(x.Value))
		}
	case *ast.BlockStmt:
		g.line("{")
		g.indent++
		for _, s := range x.Stmts {
			g.stmt(s)
		}

		g.indent--
		g.line("}")
	case *ast.IfStmt:
		g.line("if " + g.expr(x.Cond) + " {")
		g.indent++
		for _, s := range x.Then.Stmts {
			g.stmt(s)
		}

		g.indent--
		if x.Else == nil {
			g.line("}")
		} else {
			g.writeIndent()
			g.body.WriteString("} else ")
			switch e := x.Else.(type) {
			case *ast.IfStmt:
				g.ifInline(e)
			case *ast.BlockStmt:
				g.body.WriteString("{\n")
				g.indent++
				for _, s := range e.Stmts {
					g.stmt(s)
				}

				g.indent--
				g.line("}")
			}
		}
	case *ast.SwitchStmt:
		g.line("switch " + g.expr(x.Expr) + " {")
		g.indent++
		for _, cl := range x.Cases {
			if cl.Default {
				g.line("default:")
			} else {
				var vs []string

				for _, v := range cl.Values {
					vs = append(vs, g.expr(v))
				}

				g.line("case " + strings.Join(vs, ", ") + ":")
			}

			g.indent++
			for _, s := range cl.Body {
				g.stmt(s)
			}

			g.indent--
		}

		g.indent--
		g.line("}")
	case *ast.ForStmt:
		if x.Init == nil && x.Post == nil {
			g.line("for " + g.expr(x.Cond) + " {")
		} else {
			initStr, postStr := "", ""

			if x.Init != nil {
				if v, ok := x.Init.(*ast.VarDecl); ok {
					initStr = name(v.Name) + " := " + g.expr(v.Init)
				}
			}

			if x.Post != nil {
				if a, ok := x.Post.(*ast.AssignStmt); ok {
					postStr = g.expr(a.Target) + " = " + g.expr(a.Value)
				}
			}

			g.line("for " + initStr + "; " + g.expr(x.Cond) + "; " + postStr + " {")
		}

		g.indent++
		for _, s := range x.Body.Stmts {
			g.stmt(s)
		}

		g.indent--
		g.line("}")
	case *ast.ReturnStmt:
		if x.Value == nil {
			g.line("return")
		} else {
			g.line("return " + g.expr(x.Value))
		}
	case *ast.BreakStmt:
		g.line("break")
	case *ast.ContinueStmt:
		g.line("continue")
	case *ast.ExportStmt:
		if stmt, ok := x.Target.(ast.Stmt); ok {
			g.stmt(stmt)
		} else if decl, ok := x.Target.(ast.Decl); ok {
			g.decl(decl)
		}
	case *ast.ImportStmt:
		// No-op for imports
	}
}

func (g *generator) ifInline(x *ast.IfStmt) {
	g.body.WriteString("if " + g.expr(x.Cond) + " {\n")
	g.indent++
	for _, s := range x.Then.Stmts {
		g.stmt(s)
	}

	g.indent--
	if x.Else == nil {
		g.line("}")
		return
	}

	g.writeIndent()
	g.body.WriteString("} else ")
	switch e := x.Else.(type) {
	case *ast.IfStmt:
		g.ifInline(e)
	case *ast.BlockStmt:
		g.body.WriteString("{\n")
		g.indent++
		for _, s := range e.Stmts {
			g.stmt(s)
		}

		g.indent--
		g.line("}")
	}
}

func (g *generator) expr(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.LiteralExpr:
		switch x.Type {
		case ast.TypeString:
			return strconv.Quote(x.Value.(string))
		case ast.TypeChar:
			return fmt.Sprintf("byte(%s)", strconv.QuoteRune(rune(x.Value.(byte))))
		case ast.TypeBool:
			return strconv.FormatBool(x.Value.(bool))
		default:
			return fmt.Sprint(x.Value)
		}
	case *ast.IdentExpr:
		return name(x.Name)
	case *ast.UnaryExpr:
		return "(" + x.Op + g.expr(x.Right) + ")"
	case *ast.BinaryExpr:
		if x.Op == "/" || x.Op == "%" {
			helper := "nuruDiv"

			if x.Op == "%" {
				helper = "nuruMod"
			}

			g.helpers["division"] = true
			return helper + "(" + g.expr(x.Left) + ", " + g.expr(x.Right) + ", " + location(x) + ")"
		}

		return "(" + g.expr(x.Left) + " " + x.Op + " " + g.expr(x.Right) + ")"
	case *ast.FieldExpr:
		return g.expr(x.Object) + "." + name(x.Name)
	case *ast.IndexExpr:
		switch g.info.TypeOf(x.Object).Kind {
		case ast.TypeSlice:
			g.helpers["indexSlice"] = true
			return "nuruIndexSlice(" + g.expr(x.Object) + ", " + g.expr(x.Index) + ", " + location(x) + ")"
		case ast.TypeString:
			g.helpers["indexString"] = true
			return "nuruIndexString(" + g.expr(x.Object) + ", " + g.expr(x.Index) + ", " + location(x) + ")"
		default:
			return g.expr(x.Object) + "[" + g.expr(x.Index) + "]"
		}
	case *ast.SliceExpr:
		lo, hi := "", ""

		if x.Low != nil {
			lo = g.expr(x.Low)
		}

		if x.High != nil {
			hi = g.expr(x.High)
		}

		object := g.expr(x.Object)

		if lo == "" {
			lo = "0"
		}

		if hi == "" {
			hi = "-1"
		}

		if g.info.TypeOf(x.Object).Kind == ast.TypeString {
			g.helpers["sliceString"] = true
			return "nuruSliceString(" + object + ", " + lo + ", " + hi + ", " + location(x) + ")"
		}

		g.helpers["sliceSlice"] = true
		return "nuruSliceSlice(" + object + ", " + lo + ", " + hi + ", " + location(x) + ")"
	case *ast.MakeExpr:
		if x.Type.Kind == ast.TypeSlice {
			if x.Size == nil {
				return "make(" + g.typ(x.Type) + ", 0)"
			}

			g.helpers["makeSlice"] = true
			return "nuruMakeSlice[" + g.typ(x.Type.Elem) + "](" + g.expr(x.Size) + ", " + location(x) + ")"
		}

		return "make(" + g.typ(x.Type) + ")"
	case *ast.CompositeExpr:
		var es []string

		for _, el := range x.Elems {
			if x.Type.Kind == ast.TypeNamed {
				es = append(es, name(el.Key)+": "+g.expr(el.Value))
			} else if el.KeyExpr != nil {
				es = append(es, g.expr(el.KeyExpr)+": "+g.expr(el.Value))
			} else {
				es = append(es, g.expr(el.Value))
			}
		}

		prefix := g.typ(x.Type)

		if x.Type.Kind == ast.TypeNamed {
			prefix = "&" + name(x.Type.Name)
		}

		return prefix + "{" + strings.Join(es, ", ") + "}"
	case *ast.CallExpr:
		return g.call(x)
	}

	return "/* unsupported */"
}

func (g *generator) call(x *ast.CallExpr) string {
	var as []string

	for _, a := range x.Args {
		as = append(as, g.expr(a))
	}

	args := strings.Join(as, ", ")

	builtin, _ := g.info.BuiltinOf(x)

	switch builtin {
	case checker.BuiltinPrint:
		g.helpers["print"] = true
		return "nuruPrint(" + args + ")"
	case checker.BuiltinArgs:
		g.imports["os"] = true
		return "os.Args[1:]"
	case checker.BuiltinReadFile:
		g.helpers["read"] = true
		return "nuruReadFile(" + location(x) + ", " + args + ")"
	case checker.BuiltinWriteFile:
		g.helpers["write"] = true
		return "nuruWriteFile(" + location(x) + ", " + args + ")"
	case checker.BuiltinFail:
		g.helpers["fail"] = true
		return "nuruFail(" + location(x) + ", " + args + ")"
	case checker.BuiltinLen, checker.BuiltinAppend:
		return x.Callee + "(" + args + ")"
	case checker.BuiltinChar:
		g.helpers["char"] = true
		return "nuruChar(" + args + ", " + location(x) + ")"
	case checker.BuiltinInt:
		if len(x.Args) == 1 && g.info.TypeOf(x.Args[0]).Kind == ast.TypeString {
			g.helpers["atoi"] = true
			return "nuruAtoi(" + args + ", " + location(x) + ")"
		}

		return "int(" + args + ")"
	case checker.BuiltinString:
		if len(x.Args) == 1 {
			switch g.info.TypeOf(x.Args[0]).Kind {
			case ast.TypeInt:
				g.imports["strconv"] = true
				return "strconv.Itoa(" + args + ")"
			case ast.TypeBool:
				g.imports["strconv"] = true
				return "strconv.FormatBool(" + args + ")"
			}
		}

		return "string(" + args + ")"
	}

	return name(x.Callee) + "(" + args + ")"
}

func (g *generator) typ(t *ast.TypeRef) string {
	switch t.Kind {
	case ast.TypeInt:
		return "int"
	case ast.TypeChar:
		return "byte"
	case ast.TypeString:
		return "string"
	case ast.TypeBool:
		return "bool"
	case ast.TypeNamed:
		return "*" + name(t.Name)
	case ast.TypeSlice:
		return "[]" + g.typ(t.Elem)
	case ast.TypeMap:
		return "map[" + g.typ(t.Key) + "]" + g.typ(t.Elem)
	}

	return "any"
}

func (g *generator) emitHelpers() {
	if g.helpers["atoi"] || g.helpers["char"] || g.helpers["division"] || g.helpers["indexSlice"] || g.helpers["indexString"] || g.helpers["sliceSlice"] || g.helpers["sliceString"] || g.helpers["setSlice"] || g.helpers["setMap"] || g.helpers["makeSlice"] {
		g.helpers["fail"] = true
	}

	if g.helpers["read"] || g.helpers["write"] || g.helpers["fail"] {
		g.imports["fmt"] = true
		g.imports["os"] = true
		g.line("")
		g.line("func nuruFail(location, message string) {")
		g.indent++
		g.line("fmt.Fprintln(os.Stderr, location+\": runtime: \"+message)")
		g.line("os.Exit(1)")
		g.indent--
		g.line("}")
	}

	if g.helpers["read"] {
		g.line("")
		g.line("func nuruReadFile(location, path string) string {")
		g.indent++
		g.line("data, err := os.ReadFile(path)")
		g.line("if err != nil { nuruFail(location, err.Error()) }")
		g.line("return string(data)")
		g.indent--
		g.line("}")
	}

	if g.helpers["write"] {
		g.line("")
		g.line("func nuruWriteFile(location, path, contents string) {")
		g.indent++
		g.line("if err := os.WriteFile(path, []byte(contents), 0644); err != nil { nuruFail(location, err.Error()) }")
		g.indent--
		g.line("}")
	}

	if g.helpers["atoi"] {
		g.imports["strconv"] = true
		g.line("")
		g.line("func nuruAtoi(value, location string) int {")
		g.indent++
		g.line("n, err := strconv.Atoi(value)")
		g.line("if err != nil { nuruFail(location, \"invalid integer conversion\") }")
		g.line("return n")
		g.indent--
		g.line("}")
	}

	if g.helpers["char"] {
		g.line("")
		g.line("func nuruChar(value int, location string) byte {")
		g.indent++
		g.line("if value < 0 || value > 255 { nuruFail(location, \"invalid character conversion\") }")
		g.line("return byte(value)")
		g.indent--
		g.line("}")
	}

	if g.helpers["division"] {
		g.line("")
		g.line("func nuruDiv(left, right int, location string) int { if right == 0 { nuruFail(location, \"division by zero\") }; return left / right }")
		g.line("func nuruMod(left, right int, location string) int { if right == 0 { nuruFail(location, \"division by zero\") }; return left % right }")
	}

	if g.helpers["indexSlice"] {
		g.line("")
		g.line("func nuruIndexSlice[T any](value []T, index int, location string) T { if index < 0 || index >= len(value) { nuruFail(location, \"index out of bounds\") }; return value[index] }")
	}

	if g.helpers["indexString"] {
		g.line("")
		g.line("func nuruIndexString(value string, index int, location string) byte { if index < 0 || index >= len(value) { nuruFail(location, \"index out of bounds\") }; return value[index] }")
	}

	if g.helpers["sliceSlice"] {
		g.line("")
		g.line("func nuruSliceSlice[T any](value []T, low, high int, location string) []T { if high < 0 { high = len(value) }; if low < 0 || high < low || high > len(value) { nuruFail(location, \"slice bounds out of range\") }; return value[low:high] }")
	}

	if g.helpers["sliceString"] {
		g.line("")
		g.line("func nuruSliceString(value string, low, high int, location string) string { if high < 0 { high = len(value) }; if low < 0 || high < low || high > len(value) { nuruFail(location, \"slice bounds out of range\") }; return value[low:high] }")
	}

	if g.helpers["setSlice"] {
		g.line("")
		g.line("func nuruSetSlice[T any](value []T, index int, item T, location string) { if index < 0 || index >= len(value) { nuruFail(location, \"index out of bounds\") }; value[index] = item }")
	}

	if g.helpers["setMap"] {
		g.line("")
		g.line("func nuruSetMap[K comparable, V any](value map[K]V, key K, item V, location string) { if value == nil { nuruFail(location, \"assignment to uninitialized map\") }; value[key] = item }")
	}

	if g.helpers["makeSlice"] {
		g.line("")
		g.line("func nuruMakeSlice[T any](length int, location string) []T { if length < 0 { nuruFail(location, \"negative slice size\") }; return make([]T, length) }")
	}

	if g.helpers["print"] {
		g.imports["fmt"] = true
		g.imports["reflect"] = true
		g.imports["strconv"] = true
		g.imports["strings"] = true
		g.line("")
		g.line("func nuruPrint(values ...any) {")
		g.indent++
		g.line("parts := make([]string, len(values))")
		g.line("for i, value := range values { parts[i] = nuruDisplay(reflect.ValueOf(value)) }")
		g.line("fmt.Println(strings.Join(parts, \" \"))")
		g.indent--
		g.line("}")
		g.line("func nuruDisplay(value reflect.Value) string {")
		g.indent++
		g.line("if !value.IsValid() { return \"<void>\" }")
		g.line("if value.Kind() == reflect.Pointer { if value.IsNil() { return \"<nil>\" }; return strings.TrimPrefix(value.Elem().Type().Name(), \"z_\") }")
		g.line("switch value.Kind() {")
		g.line("case reflect.Int: return strconv.FormatInt(value.Int(), 10)")
		g.line("case reflect.Uint8: return string([]byte{byte(value.Uint())})")
		g.line("case reflect.String: return value.String()")
		g.line("case reflect.Bool: return strconv.FormatBool(value.Bool())")
		g.line("case reflect.Slice: parts := make([]string, value.Len()); for i := range parts { parts[i] = nuruDisplay(value.Index(i)) }; return \"[\"+strings.Join(parts, \" \")+\"]\"")
		g.line("case reflect.Map: return fmt.Sprintf(\"map[%d entries]\", value.Len())")
		g.line("}")
		g.line("return \"<invalid>\"")
		g.indent--
		g.line("}")
	}
}

func (g *generator) line(s string)          { g.writeIndent(); g.body.WriteString(s); g.body.WriteByte('\n') }
func (g *generator) writeIndent()           { g.body.WriteString(strings.Repeat("\t", g.indent)) }
func (g *generator) rawBlockStart(s string) { g.line(s + " {"); g.indent++ }
func (g *generator) rawBlockEnd()           { g.indent--; g.line("}") }
func name(s string) string                  { return "z_" + s }
func location(node ast.Node) string         { return strconv.Quote(node.GetSpan().String()) }
