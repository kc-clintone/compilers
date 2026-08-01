// Package parser implements Nuru's handwritten recursive-descent/Pratt parser.
package parser

import (
	"github.com/kc-clintone/compilers/internal/ast"
	"github.com/kc-clintone/compilers/internal/diagnostic"
	"github.com/kc-clintone/compilers/internal/lexer"
	"github.com/kc-clintone/compilers/internal/source"
	"github.com/kc-clintone/compilers/internal/token"
)

// Parser holds the state of one recursive-descent/Pratt parse.
type Parser struct {
	tokens              []token.Token
	current             int
	diags               []diagnostic.Diagnostic
	allowNamedComposite bool
}

// Parse converts source into an AST. Callers must discard the returned partial
// program whenever diagnostics are present.
func Parse(filename string, input []byte) (*ast.Program, []diagnostic.Diagnostic) {
	toks, ds := lexer.Lex(filename, input)

	if len(ds) > 0 {
		return nil, ds
	}

	p := &Parser{tokens: toks, allowNamedComposite: true}
	program := p.program()

	ast.AssignNodeIDs(program)

	if len(p.diags) > 0 {
		return program, p.diags
	}

	return program, nil
}

func (p *Parser) program() *ast.Program {
	start := p.peek().Span
	out := &ast.Program{}

	for p.isDeclStart() {
		d := p.declaration()
		if d != nil {
			out.Decls = append(out.Decls, d)
		} else {
			p.sync()
		}
	}

	for !p.atEnd() {
		if p.isDeclStart() {
			p.err(p.peek(), "declarations must precede top-level statements")
			// Consume the misplaced declaration before recovering. Calling sync
			// while still positioned on a declaration starter would make no
			// progress because declaration starters are synchronization points.
			p.declaration()
			continue
		}

		s := p.statement()
		if s != nil {
			out.Stmts = append(out.Stmts, s)
		} else {
			p.sync()
		}
	}

	out.Span = merge(start, p.previous().Span)
	return out
}

func (p *Parser) isDeclStart() bool {
	return p.check(token.Type) || p.check(token.Func) || p.check(token.Var) || p.check(token.Export) || p.check(token.Import)
}

func (p *Parser) declaration() ast.Decl {
	if p.match(token.Export) {
		start := p.previous()
		var target ast.Node

		if p.isDeclStart() {
			target = p.declaration()
		} else {
			target = p.statement()
		}

		span := start.Span

		if target != nil {
			span = merge(start.Span, target.GetSpan())
		}

		return &ast.ExportStmt{Base: ast.Base{Span: span}, Target: target}
	}

	if p.match(token.Import) {
		start := p.previous()
		var path string

		if p.match(token.String) {
			path = p.previous().Literal.(string)
		} else if p.match(token.Ident) {
			path = p.previous().Lexeme
		} else {
			p.err(p.peek(), "expected module path or identifier after import")
		}

		end := p.consume(token.Semicolon, "expected ';' after import statement")

		return &ast.ImportStmt{Base: ast.Base{Span: merge(start.Span, end.Span)}, Path: path}
	}

	if p.match(token.Type) {
		return p.structDecl()
	}

	if p.match(token.Func) {
		return p.funcDecl()
	}

	if p.match(token.Var) {
		v := p.varDecl(p.previous())

		p.consume(token.Semicolon, "expected ';' after variable declaration")
		return v
	}

	return nil
}

func (p *Parser) structDecl() ast.Decl {
	start := p.previous()
	name := p.consume(token.Ident, "expected type name")

	p.consume(token.Struct, "expected 'struct'")
	p.consume(token.LBrace, "expected '{'")
	d := &ast.StructDecl{Name: name.Lexeme}

	for !p.check(token.RBrace) && !p.atEnd() {
		n := p.consume(token.Ident, "expected field name")
		t := p.parseType()
		semi := p.consume(token.Semicolon, "expected ';' after field")

		d.Fields = append(d.Fields, ast.Field{Name: n.Lexeme, Type: t, Span: merge(n.Span, semi.Span)})
	}

	end := p.consume(token.RBrace, "expected '}'")

	d.Span = merge(start.Span, end.Span)
	return d
}

func (p *Parser) funcDecl() ast.Decl {
	start := p.previous()
	name := p.consume(token.Ident, "expected function name")

	p.consume(token.LParen, "expected '('")
	var params []ast.Param

	if !p.check(token.RParen) {
		for {
			n := p.consume(token.Ident, "expected parameter name")
			t := p.parseType()

			params = append(params, ast.Param{Name: n.Lexeme, Type: t, Span: merge(n.Span, t.Span)})
			if !p.match(token.Comma) {
				break
			}
		}
	}

	p.consume(token.RParen, "expected ')'")
	var result *ast.TypeRef

	if p.typeStart() {
		result = p.parseType()
	}

	body := p.block()
	d := &ast.FuncDecl{Name: name.Lexeme, Params: params, Result: result, Body: body}

	d.Span = merge(start.Span, body.Span)
	return d
}

func (p *Parser) varDecl(start token.Token) *ast.VarDecl {
	name := p.consume(token.Ident, "expected variable name")
	t := p.parseType()
	var init ast.Expr

	if p.match(token.Assign) {
		init = p.expression()
	}

	d := &ast.VarDecl{Name: name.Lexeme, Type: t, Init: init}

	d.Span = merge(start.Span, lastSpan(t, init))
	return d
}

func (p *Parser) parseType() *ast.TypeRef {
	start := p.peek()
	t := &ast.TypeRef{}

	switch {
	case p.match(token.IntType):
		t.Kind = ast.TypeInt
	case p.match(token.CharType):
		t.Kind = ast.TypeChar
	case p.match(token.StringType):
		t.Kind = ast.TypeString
	case p.match(token.BoolType):
		t.Kind = ast.TypeBool
	case p.match(token.LBracket):
		p.consume(token.RBracket, "expected ']' in slice type")
		t.Kind = ast.TypeSlice
		t.Elem = p.parseType()
	case p.match(token.Map):
		p.consume(token.LBracket, "expected '[' after map")
		t.Kind = ast.TypeMap
		t.Key = p.parseType()
		p.consume(token.RBracket, "expected ']' after map key")
		t.Elem = p.parseType()
	case p.match(token.Ident):
		t.Kind = ast.TypeNamed
		t.Name = p.previous().Lexeme
	default:
		p.err(p.peek(), "expected type")
		p.advance()
		t.Kind = ast.TypeInvalid
	}

	t.Span = merge(start.Span, p.previous().Span)
	return t
}

func (p *Parser) typeStart() bool {
	return p.check(token.IntType) || p.check(token.CharType) || p.check(token.StringType) || p.check(token.BoolType) || p.check(token.LBracket) || p.check(token.Map) || p.check(token.Ident)
}

func (p *Parser) statement() ast.Stmt {
	if p.match(token.Export) {
		start := p.previous()
		var target ast.Node

		if p.isDeclStart() {
			target = p.declaration()
		} else {
			target = p.statement()
		}

		span := start.Span

		if target != nil {
			span = merge(start.Span, target.GetSpan())
		}

		return &ast.ExportStmt{Base: ast.Base{Span: span}, Target: target}
	}

	if p.match(token.Import) {
		start := p.previous()
		var path string

		if p.match(token.String) {
			path = p.previous().Literal.(string)
		} else if p.match(token.Ident) {
			path = p.previous().Lexeme
		} else {
			p.err(p.peek(), "expected module path or identifier after import")
		}

		end := p.consume(token.Semicolon, "expected ';' after import statement")

		return &ast.ImportStmt{Base: ast.Base{Span: merge(start.Span, end.Span)}, Path: path}
	}

	if p.match(token.Var) {
		v := p.varDecl(p.previous())

		p.consume(token.Semicolon, "expected ';' after variable declaration")
		return v
	}

	if p.match(token.If) {
		return p.ifStmt()
	}

	if p.match(token.Switch) {
		return p.switchStmt()
	}

	if p.match(token.For) {
		return p.forStmt()
	}

	if p.match(token.Return) {
		return p.returnStmt()
	}

	if p.match(token.Break) {
		s := &ast.BreakStmt{}
		end := p.consume(token.Semicolon, "expected ';' after break")

		s.Span = merge(p.previous().Span, end.Span)
		return s
	}

	if p.match(token.Continue) {
		s := &ast.ContinueStmt{}
		end := p.consume(token.Semicolon, "expected ';' after continue")

		s.Span = merge(p.previous().Span, end.Span)
		return s
	}

	if p.check(token.LBrace) {
		return p.block()
	}

	start := p.peek()
	e := p.expression()

	if p.match(token.Assign) {
		v := p.expression()
		end := p.consume(token.Semicolon, "expected ';' after assignment")
		s := &ast.AssignStmt{Target: e, Value: v}

		s.Span = merge(start.Span, end.Span)
		return s
	}

	end := p.consume(token.Semicolon, "expected ';' after expression")
	s := &ast.ExprStmt{Expr: e}

	s.Span = merge(start.Span, end.Span)
	return s
}

func (p *Parser) block() *ast.BlockStmt {
	start := p.consume(token.LBrace, "expected '{'")
	b := &ast.BlockStmt{}

	for !p.check(token.RBrace) && !p.atEnd() {
		s := p.statement()
		if s != nil {
			b.Stmts = append(b.Stmts, s)
		} else {
			p.sync()
		}
	}

	end := p.consume(token.RBrace, "expected '}'")

	b.Span = merge(start.Span, end.Span)
	return b
}

func (p *Parser) ifStmt() ast.Stmt {
	start := p.previous()
	cond := p.conditionExpression()
	then := p.block()
	var other ast.Stmt

	if p.match(token.Else) {
		if p.match(token.If) {
			other = p.ifStmt()
		} else {
			other = p.block()
		}
	}

	s := &ast.IfStmt{Cond: cond, Then: then, Else: other}

	s.Span = merge(start.Span, lastStmtSpan(then, other))
	return s
}

func (p *Parser) switchStmt() ast.Stmt {
	start := p.previous()
	subject := p.conditionExpression()

	p.consume(token.LBrace, "expected '{' after switch expression")
	s := &ast.SwitchStmt{Expr: subject}

	for !p.check(token.RBrace) && !p.atEnd() {
		c := &ast.CaseClause{}
		ct := p.peek()

		if p.match(token.Case) {
			c.Values = append(c.Values, p.expression())
			p.consume(token.Colon, "expected ':' after case")
		} else if p.match(token.Default) {
			c.Default = true
			p.consume(token.Colon, "expected ':' after default")
		} else {
			p.err(p.peek(), "expected case or default")
			p.advance()
			continue
		}

		for !p.check(token.Case) && !p.check(token.Default) && !p.check(token.RBrace) && !p.atEnd() {
			c.Body = append(c.Body, p.statement())
		}

		c.Span = merge(ct.Span, p.previous().Span)
		s.Cases = append(s.Cases, c)
	}

	end := p.consume(token.RBrace, "expected '}' after switch")

	s.Span = merge(start.Span, end.Span)
	return s
}

func (p *Parser) forStmt() ast.Stmt {
	start := p.previous()
	f := &ast.ForStmt{}

	if p.match(token.Var) {
		f.Init = p.varDecl(p.previous())
		p.consume(token.Semicolon, "expected ';' after loop initializer")
		f.Cond = p.expression()
		p.consume(token.Semicolon, "expected ';' after loop condition")
		target := p.expression()

		p.consume(token.Assign, "expected '=' in loop update")
		value := p.expression()
		a := &ast.AssignStmt{Target: target, Value: value}

		a.Span = merge(target.GetSpan(), value.GetSpan())
		f.Post = a
	} else {
		f.Cond = p.conditionExpression()
	}

	f.Body = p.block()
	f.Span = merge(start.Span, f.Body.Span)
	return f
}

func (p *Parser) returnStmt() ast.Stmt {
	start := p.previous()
	var v ast.Expr

	if !p.check(token.Semicolon) {
		v = p.expression()
	}

	end := p.consume(token.Semicolon, "expected ';' after return")
	s := &ast.ReturnStmt{Value: v}

	s.Span = merge(start.Span, end.Span)
	return s
}

// --- Recursive Descent Expression Parser ---
//
// Expression parsing enforces operator precedence using a strict function call hierarchy
// to make grammar rules transparent and easy for learners to follow:
//
//   1. expression()  -> logicalOr()
//   2. logicalOr()   -> logicalAnd() ("||" logicalAnd())*
//   3. logicalAnd()  -> equality() ("&&" equality())*
//   4. equality()    -> comparison() (("==" | "!=") comparison())*
//   5. comparison()  -> term() (("<" | "<=" | ">" | ">=") term())*
//   6. term()        -> factor() (("+" | "-") factor())*
//   7. factor()      -> unary() (("*" | "/" | "%") unary())*
//   8. unary()       -> ("!" | "-") unary() | postfix()
//   9. postfix()     -> primary() ("(" args? ")" | "." ident | "[" index/slice "]")*
//  10. primary()     -> literals | identifiers | "(" expression ")" | make(...) | composite lit

// expression is the entry point for parsing expressions.
func (p *Parser) expression() ast.Expr {
	return p.logicalOr()
}

func (p *Parser) conditionExpression() ast.Expr {
	// A named composite literal and a control-flow body both start with an
	// identifier followed by "{". Leave that brace for the statement parser
	// while reading an unparenthesized if, switch, or while-style for header.
	previous := p.allowNamedComposite

	p.allowNamedComposite = false
	defer func() { p.allowNamedComposite = previous }()
	return p.expression()
}

// logicalOr parses logical OR expressions ("||") at the lowest binary precedence level.
// Grammar: LogicalOr -> LogicalAnd ( "||" LogicalAnd )*
func (p *Parser) logicalOr() ast.Expr {
	left := p.logicalAnd()

	for p.match(token.Or) {
		op := p.previous()
		right := p.logicalAnd()
		b := &ast.BinaryExpr{Left: left, Op: op.Lexeme, Right: right}

		b.Span = merge(left.GetSpan(), right.GetSpan())
		left = b
	}

	return left
}

// logicalAnd parses logical AND expressions ("&&").
// Grammar: LogicalAnd -> Equality ( "&&" Equality )*
func (p *Parser) logicalAnd() ast.Expr {
	left := p.equality()

	for p.match(token.And) {
		op := p.previous()
		right := p.equality()
		b := &ast.BinaryExpr{Left: left, Op: op.Lexeme, Right: right}

		b.Span = merge(left.GetSpan(), right.GetSpan())
		left = b
	}

	return left
}

// equality parses equality comparison expressions ("==", "!=").
// Grammar: Equality -> Comparison ( ( "==" | "!=" ) Comparison )*
func (p *Parser) equality() ast.Expr {
	left := p.comparison()

	for p.match(token.Equal, token.NotEqual) {
		op := p.previous()
		right := p.comparison()
		b := &ast.BinaryExpr{Left: left, Op: op.Lexeme, Right: right}

		b.Span = merge(left.GetSpan(), right.GetSpan())
		left = b
	}

	return left
}

// comparison parses relational comparison expressions ("<", "<=", ">", ">=").
// Grammar: Comparison -> Term ( ( "<" | "<=" | ">" | ">=" ) Term )*
func (p *Parser) comparison() ast.Expr {
	left := p.term()

	for p.match(token.Less, token.LessEqual, token.Greater, token.GreaterEqual) {
		op := p.previous()
		right := p.term()
		b := &ast.BinaryExpr{Left: left, Op: op.Lexeme, Right: right}

		b.Span = merge(left.GetSpan(), right.GetSpan())
		left = b
	}

	return left
}

// term parses addition and subtraction expressions ("+", "-").
// Grammar: Term -> Factor ( ( "+" | "-" ) Factor )*
func (p *Parser) term() ast.Expr {
	left := p.factor()

	for p.match(token.Plus, token.Minus) {
		op := p.previous()
		right := p.factor()
		b := &ast.BinaryExpr{Left: left, Op: op.Lexeme, Right: right}

		b.Span = merge(left.GetSpan(), right.GetSpan())
		left = b
	}

	return left
}

// factor parses multiplication, division, and modulo expressions ("*", "/", "%").
// Grammar: Factor -> Unary ( ( "*" | "/" | "%" ) Unary )*
func (p *Parser) factor() ast.Expr {
	left := p.unary()

	for p.match(token.Star, token.Slash, token.Percent) {
		op := p.previous()
		right := p.unary()
		b := &ast.BinaryExpr{Left: left, Op: op.Lexeme, Right: right}

		b.Span = merge(left.GetSpan(), right.GetSpan())
		left = b
	}

	return left
}

func (p *Parser) unary() ast.Expr {
	if p.match(token.Bang, token.Minus) {
		op := p.previous()
		right := p.unary()
		u := &ast.UnaryExpr{Op: op.Lexeme, Right: right}

		u.Span = merge(op.Span, right.GetSpan())
		return u
	}

	return p.postfix()
}

func (p *Parser) postfix() ast.Expr {
	e := p.primary()

	for {
		if p.match(token.LParen) {
			id, ok := e.(*ast.IdentExpr)

			if !ok {
				p.err(p.previous(), "only named functions can be called")
			}

			var args []ast.Expr

			if !p.check(token.RParen) {
				for {
					args = append(args, p.expression())
					if !p.match(token.Comma) {
						break
					}
				}
			}

			end := p.consume(token.RParen, "expected ')' after arguments")
			c := &ast.CallExpr{Args: args}

			if ok {
				c.Callee = id.Name
			}

			c.Span = merge(e.GetSpan(), end.Span)
			e = c
		} else if p.match(token.Dot) {
			name := p.consume(token.Ident, "expected field name")
			f := &ast.FieldExpr{Object: e, Name: name.Lexeme}

			f.Span = merge(e.GetSpan(), name.Span)
			e = f
		} else if p.match(token.LBracket) {
			start := e.GetSpan()

			if p.match(token.Colon) {
				var hi ast.Expr

				if !p.check(token.RBracket) {
					hi = p.expression()
				}

				end := p.consume(token.RBracket, "expected ']'")
				s := &ast.SliceExpr{Object: e, High: hi}

				s.Span = merge(start, end.Span)
				e = s
				continue
			}

			first := p.expression()

			if p.match(token.Colon) {
				var hi ast.Expr

				if !p.check(token.RBracket) {
					hi = p.expression()
				}

				end := p.consume(token.RBracket, "expected ']'")
				s := &ast.SliceExpr{Object: e, Low: first, High: hi}

				s.Span = merge(start, end.Span)
				e = s
			} else {
				end := p.consume(token.RBracket, "expected ']'")
				i := &ast.IndexExpr{Object: e, Index: first}

				i.Span = merge(start, end.Span)
				e = i
			}
		} else {
			break
		}
	}

	return e
}

func (p *Parser) primary() ast.Expr {
	t := p.advance()

	switch t.Kind {
	case token.Integer:
		return literal(t, t.Literal, ast.TypeInt)
	case token.String:
		return literal(t, t.Literal, ast.TypeString)
	case token.Char:
		return literal(t, t.Literal, ast.TypeChar)
	case token.True:
		return literal(t, true, ast.TypeBool)
	case token.False:
		return literal(t, false, ast.TypeBool)
	case token.LParen:
		e := p.expression()

		p.consume(token.RParen, "expected ')'")
		return e
	case token.Make:
		p.consume(token.LParen, "expected '(' after make")
		typ := p.parseType()
		var size ast.Expr

		if p.match(token.Comma) {
			size = p.expression()
		}

		end := p.consume(token.RParen, "expected ')' after make")
		m := &ast.MakeExpr{Type: typ, Size: size}

		m.Span = merge(t.Span, end.Span)
		return m
	case token.LBracket:
		p.current--
		typ := p.parseType()

		return p.composite(typ)
	case token.Map:
		p.current--
		typ := p.parseType()

		return p.composite(typ)
	case token.Ident:
		if p.allowNamedComposite && p.check(token.LBrace) {
			typ := &ast.TypeRef{Kind: ast.TypeNamed, Name: t.Lexeme}

			typ.Span = t.Span
			return p.composite(typ)
		}

		id := &ast.IdentExpr{Name: t.Lexeme}

		id.Span = t.Span
		return id
	case token.IntType, token.CharType, token.StringType, token.BoolType:
		// Primitive type names are expression names only when used as explicit
		// conversion calls, for example string(42) or int("42").
		id := &ast.IdentExpr{Name: t.Lexeme}

		id.Span = t.Span
		return id
	default:
		p.err(t, "expected expression")
		x := &ast.LiteralExpr{Value: 0, Type: ast.TypeInvalid}

		x.Span = t.Span
		return x
	}
}

func (p *Parser) composite(typ *ast.TypeRef) ast.Expr {
	start := typ.Span

	p.consume(token.LBrace, "expected '{' after composite type")
	c := &ast.CompositeExpr{Type: typ}

	for !p.check(token.RBrace) && !p.atEnd() {
		el := ast.CompositeElem{Span: p.peek().Span}

		if typ.Kind == ast.TypeNamed && p.check(token.Ident) && p.peekN(1).Kind == token.Colon {
			el.Key = p.advance().Lexeme
			p.advance()
			el.Value = p.expression()
		} else {
			first := p.expression()

			if p.match(token.Colon) {
				el.KeyExpr = first
				el.Value = p.expression()
			} else {
				el.Value = first
			}
		}

		el.Span = merge(el.Span, el.Value.GetSpan())
		c.Elems = append(c.Elems, el)
		if !p.match(token.Comma) {
			break
		}
	}

	end := p.consume(token.RBrace, "expected '}' after composite literal")

	c.Span = merge(start, end.Span)
	return c
}

func literal(t token.Token, v any, k ast.TypeKind) ast.Expr {
	x := &ast.LiteralExpr{Value: v, Type: k}

	x.Span = t.Span
	return x
}

func (p *Parser) peek() token.Token { return p.tokens[p.current] }
func (p *Parser) peekN(n int) token.Token {
	i := p.current + n

	if i >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}

	return p.tokens[i]
}

func (p *Parser) previous() token.Token {
	if p.current == 0 {
		return p.tokens[0]
	}

	return p.tokens[p.current-1]
}
func (p *Parser) atEnd() bool { return p.peek().Kind == token.EOF }
func (p *Parser) advance() token.Token {
	if !p.atEnd() {
		p.current++
	}

	return p.previous()
}
func (p *Parser) check(k token.Kind) bool { return p.peek().Kind == k }
func (p *Parser) match(ks ...token.Kind) bool {
	for _, k := range ks {
		if p.check(k) {
			p.advance()
			return true
		}
	}

	return false
}

func (p *Parser) consume(k token.Kind, msg string) token.Token {
	if p.check(k) {
		return p.advance()
	}

	p.err(p.peek(), msg)
	return p.peek()
}

func (p *Parser) err(t token.Token, msg string) {
	p.diags = append(p.diags, diagnostic.Diagnostic{Span: t.Span, Phase: "parser", Message: msg})
}

func (p *Parser) sync() {
	for !p.atEnd() {
		if p.previous().Kind == token.Semicolon {
			return
		}

		switch p.peek().Kind {
		case token.Var, token.Type, token.Func, token.If, token.For, token.Switch, token.Return, token.RBrace:
			return
		}

		p.advance()
	}
}

func merge(a, b source.Span) source.Span {
	if a.Filename == "" {
		return b
	}

	if b.Filename == "" {
		return a
	}

	return source.Span{Filename: a.Filename, Start: a.Start, End: b.End}
}

func lastSpan(t *ast.TypeRef, e ast.Expr) source.Span {
	if e != nil {
		return e.GetSpan()
	}

	return t.Span
}

func lastStmtSpan(a, b ast.Stmt) source.Span {
	if b != nil {
		return b.GetSpan()
	}

	return a.GetSpan()
}
