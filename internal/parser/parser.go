package parser

import (
	"fmt"

	"github.com/kc-clintone/compilers/internal/ast"
	"github.com/kc-clintone/compilers/internal/diagnostic"
	"github.com/kc-clintone/compilers/internal/lexer"
	"github.com/kc-clintone/compilers/internal/source"
	"github.com/kc-clintone/compilers/internal/token"
)

// Parser holds state for recursive descent parsing.
type Parser struct {
	filename string
	tokens   []token.Token
	current  int
	diags    []diagnostic.Diagnostic
}

// Parse tokenizes input and parses it into an AST Program.
func Parse(filename string, input []byte) (*ast.Program, []diagnostic.Diagnostic) {
	toks, ds := lexer.Lex(filename, input)
	if len(ds) > 0 {
		return nil, ds
	}

	p := &Parser{filename: filename, tokens: toks}
	prog := p.program()

	if len(p.diags) > 0 {
		return prog, p.diags
	}
	return prog, nil
}

func (p *Parser) program() *ast.Program {
	start := p.peek().Span
	prog := &ast.Program{Base: ast.Base{Span: start}}

	for !p.atEnd() {
		// Top-level item: declaration or statement
		if p.check(token.Kind("func")) || p.check(token.Kind("var")) {
			d := p.declaration()
			if d != nil {
				prog.Decls = append(prog.Decls, d)
			}
		} else {
			s := p.statement()
			if s != nil {
				prog.Stmts = append(prog.Stmts, s)
			}
		}
	}

	if len(p.tokens) > 0 {
		prog.Span = mergeSpan(start, p.previous().Span)
	}
	return prog
}

func (p *Parser) declaration() ast.Decl {
	if p.check(token.Kind("func")) {
		return p.parseFuncDecl()
	}
	if p.check(token.Kind("var")) {
		return p.parseVarDecl()
	}
	p.error("expected declaration")
	p.advance()
	return nil
}

// TODO: Parser - Implement parseFuncDecl() for Stage 2!
// Instructions: Parse 'func <name>(<params...>) { <body> }' and return *ast.FuncDecl.
func (p *Parser) parseFuncDecl() ast.Decl {
	// TODO: Parser - Function declaration parsing goes here in Stage 2.
	p.error("function declarations ('func') are not implemented yet in this checkpoint")
	p.advance()
	return nil
}

// TODO: Parser - Implement parseVarDecl() for Stage 2!
// Instructions: Parse 'var <name> = <expr>' and return *ast.VarDecl.
func (p *Parser) parseVarDecl() *ast.VarDecl {
	// TODO: Parser - Variable declaration parsing goes here in Stage 2.
	p.error("variable declarations ('var') are not implemented yet in this checkpoint")
	p.advance()
	return nil
}

func (p *Parser) statement() ast.Stmt {
	if p.check(token.Kind("var")) {
		p.advance()
		return p.parseVarDeclAfterKeyword()
	}
	if p.match(token.If) {
		return p.parseIfStmt()
	}
	if p.match(token.For) {
		return p.parseForStmt()
	}
	if p.match(token.LBrace) {
		return p.parseBlockStmt()
	}
	if p.match(token.Break) {
		return &ast.BreakStmt{Base: ast.Base{Span: p.previous().Span}}
	}
	if p.match(token.Continue) {
		return &ast.ContinueStmt{Base: ast.Base{Span: p.previous().Span}}
	}
	if p.check(token.Ident) && p.peek().Lexeme == "print" {
		return p.parsePrintStmt()
	}

	// Check for assignment statement: <ident> = <expr>
	if p.check(token.Ident) && p.peekNext().Kind == token.Assign {
		return p.parseAssignStmt()
	}

	// TODO: Parser - Implement parseReturnStmt() for Stage 2!
	if p.check(token.Kind("return")) || (p.check(token.Ident) && p.peek().Lexeme == "return") {
		return p.parseReturnStmt()
	}

	// Fallback: Expression statement
	expr := p.expression()
	if expr == nil {
		p.advance()
		return nil
	}
	return &ast.ExprStmt{Base: ast.Base{Span: expr.GetSpan()}, Expr: expr}
}

func (p *Parser) parseVarDeclAfterKeyword() *ast.VarDecl {
	// TODO: Parser - Complete variable declaration parsing after 'var' keyword.
	p.error("variable declarations ('var') are not implemented yet in this checkpoint")
	return nil
}

// TODO: Parser - Implement parseAssignStmt() for Stage 2!
// Instructions: Parse '<name> = <expr>' and return *ast.AssignStmt.
func (p *Parser) parseAssignStmt() ast.Stmt {
	// TODO: Parser - Variable assignment parsing goes here in Stage 2.
	p.error("variable assignments are not implemented yet in this checkpoint")
	p.advance()
	p.advance()
	return nil
}

// TODO: Parser - Implement parseReturnStmt() for Stage 2!
// Instructions: Parse 'return <expr>' and return *ast.ReturnStmt.
func (p *Parser) parseReturnStmt() ast.Stmt {
	// TODO: Parser - Return statement parsing goes here in Stage 2.
	p.error("return statements are not implemented yet in this checkpoint")
	p.advance()
	return nil
}

func (p *Parser) parsePrintStmt() ast.Stmt {
	start := p.advance() // consume 'print'
	p.consume(token.LParen, "expected '(' after 'print'")

	var args []ast.Expr
	if !p.check(token.RParen) {
		for {
			arg := p.expression()
			if arg != nil {
				args = append(args, arg)
			}
			if !p.match(token.Comma) {
				break
			}
		}
	}
	end := p.consume(token.RParen, "expected ')' after print arguments")
	return &ast.PrintStmt{Base: ast.Base{Span: mergeSpan(start.Span, end.Span)}, Args: args}
}

func (p *Parser) parseIfStmt() ast.Stmt {
	start := p.previous()
	cond := p.expression()
	p.consume(token.LBrace, "expected '{' after if condition")
	thenBranch := p.parseBlockStmt()

	var elseBranch ast.Stmt
	if p.match(token.Else) {
		if p.match(token.If) {
			elseBranch = p.parseIfStmt()
		} else if p.match(token.LBrace) {
			elseBranch = p.parseBlockStmt()
		} else {
			p.error("expected '{' or 'if' after 'else'")
		}
	}

	span := start.Span
	if elseBranch != nil {
		span = mergeSpan(start.Span, elseBranch.GetSpan())
	} else if thenBranch != nil {
		span = mergeSpan(start.Span, thenBranch.GetSpan())
	}

	return &ast.IfStmt{Base: ast.Base{Span: span}, Cond: cond, Then: thenBranch, Else: elseBranch}
}

func (p *Parser) parseForStmt() ast.Stmt {
	start := p.previous()
	cond := p.expression()
	p.consume(token.LBrace, "expected '{' after for condition")
	body := p.parseBlockStmt()

	return &ast.ForStmt{Base: ast.Base{Span: mergeSpan(start.Span, body.GetSpan())}, Cond: cond, Body: body}
}

func (p *Parser) parseBlockStmt() *ast.BlockStmt {
	start := p.previous()
	var stmts []ast.Stmt

	for !p.check(token.RBrace) && !p.atEnd() {
		s := p.statement()
		if s != nil {
			stmts = append(stmts, s)
		}
	}
	end := p.consume(token.RBrace, "expected '}' to close block")
	return &ast.BlockStmt{Base: ast.Base{Span: mergeSpan(start.Span, end.Span)}, Stmts: stmts}
}

// --- Recursive Descent Expression Parser ---

func (p *Parser) expression() ast.Expr {
	return p.logicalOr()
}

func (p *Parser) logicalOr() ast.Expr {
	expr := p.logicalAnd()
	for p.match(token.Or) {
		op := p.previous().Lexeme
		right := p.logicalAnd()
		expr = &ast.BinaryExpr{Base: ast.Base{Span: mergeSpan(expr.GetSpan(), right.GetSpan())}, Left: expr, Op: op, Right: right}
	}
	return expr
}

func (p *Parser) logicalAnd() ast.Expr {
	expr := p.equality()
	for p.match(token.And) {
		op := p.previous().Lexeme
		right := p.equality()
		expr = &ast.BinaryExpr{Base: ast.Base{Span: mergeSpan(expr.GetSpan(), right.GetSpan())}, Left: expr, Op: op, Right: right}
	}
	return expr
}

func (p *Parser) equality() ast.Expr {
	expr := p.comparison()
	for p.match(token.Equal) || p.match(token.NotEqual) {
		op := p.previous().Lexeme
		right := p.comparison()
		expr = &ast.BinaryExpr{Base: ast.Base{Span: mergeSpan(expr.GetSpan(), right.GetSpan())}, Left: expr, Op: op, Right: right}
	}
	return expr
}

func (p *Parser) comparison() ast.Expr {
	expr := p.additive()
	for p.match(token.Less) || p.match(token.LessEqual) || p.match(token.Greater) || p.match(token.GreaterEqual) {
		op := p.previous().Lexeme
		right := p.additive()
		expr = &ast.BinaryExpr{Base: ast.Base{Span: mergeSpan(expr.GetSpan(), right.GetSpan())}, Left: expr, Op: op, Right: right}
	}
	return expr
}

func (p *Parser) additive() ast.Expr {
	expr := p.multiplicative()
	for p.match(token.Plus) || p.match(token.Minus) {
		op := p.previous().Lexeme
		right := p.multiplicative()
		expr = &ast.BinaryExpr{Base: ast.Base{Span: mergeSpan(expr.GetSpan(), right.GetSpan())}, Left: expr, Op: op, Right: right}
	}
	return expr
}

func (p *Parser) multiplicative() ast.Expr {
	expr := p.unary()
	for p.match(token.Star) || p.match(token.Slash) || p.match(token.Percent) {
		op := p.previous().Lexeme
		right := p.unary()
		expr = &ast.BinaryExpr{Base: ast.Base{Span: mergeSpan(expr.GetSpan(), right.GetSpan())}, Left: expr, Op: op, Right: right}
	}
	return expr
}

func (p *Parser) unary() ast.Expr {
	if p.match(token.Minus) || p.match(token.Bang) {
		opTok := p.previous()
		right := p.unary()
		return &ast.UnaryExpr{Base: ast.Base{Span: mergeSpan(opTok.Span, right.GetSpan())}, Op: opTok.Lexeme, Right: right}
	}
	return p.primary()
}

func (p *Parser) primary() ast.Expr {
	if p.match(token.Integer) || p.match(token.String) || p.match(token.True) || p.match(token.False) {
		tok := p.previous()
		return &ast.LiteralExpr{Base: ast.Base{Span: tok.Span}, Value: tok.Literal}
	}

	if p.match(token.LParen) {
		expr := p.expression()
		p.consume(token.RParen, "expected ')' after expression")
		return expr
	}

	if p.check(token.Ident) {
		tok := p.peek()
		if p.peekNext().Kind == token.LParen {
			return p.parseCallExpr()
		}
		return p.parseIdentExpr()
	}

	p.error(fmt.Sprintf("unexpected token '%s'", p.peek().Lexeme))
	p.advance()
	return nil
}

// TODO: Parser - Implement parseIdentExpr() for Stage 2!
// Instructions: Return *ast.IdentExpr containing identifier name.
func (p *Parser) parseIdentExpr() ast.Expr {
	// TODO: Parser - Identifier expression parsing goes here in Stage 2.
	tok := p.advance()
	p.error(fmt.Sprintf("identifier expressions ('%s') are not implemented yet in this checkpoint", tok.Lexeme))
	return nil
}

// TODO: Parser - Implement parseCallExpr() for Stage 2!
// Instructions: Parse '<callee>(<args...>)' and return *ast.CallExpr.
func (p *Parser) parseCallExpr() ast.Expr {
	// TODO: Parser - Function call parsing goes here in Stage 2.
	tok := p.advance()
	p.error(fmt.Sprintf("function calls ('%s(...)') are not implemented yet in this checkpoint", tok.Lexeme))
	p.consume(token.LParen, "")
	p.consume(token.RParen, "")
	return nil
}

// Helper methods for parser state navigation
func (p *Parser) atEnd() bool { return p.peek().Kind == token.EOF }
func (p *Parser) peek() token.Token {
	if p.current >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.current]
}
func (p *Parser) peekNext() token.Token {
	if p.current+1 >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.current+1]
}
func (p *Parser) previous() token.Token { return p.tokens[p.current-1] }
func (p *Parser) check(k token.Kind) bool {
	if p.atEnd() {
		return false
	}
	return p.peek().Kind == k
}
func (p *Parser) advance() token.Token {
	if !p.atEnd() {
		p.current++
	}
	return p.previous()
}
func (p *Parser) match(k token.Kind) bool {
	if p.check(k) {
		p.advance()
		return true
	}
	return false
}
func (p *Parser) consume(k token.Kind, msg string) token.Token {
	if p.check(k) {
		return p.advance()
	}
	p.error(msg)
	return p.peek()
}
func (p *Parser) error(msg string) {
	p.diags = append(p.diags, diagnostic.Diagnostic{Span: p.peek().Span, Phase: "parser", Message: msg})
}

func mergeSpan(a, b source.Span) source.Span {
	return source.Span{Filename: a.Filename, Start: a.Start, End: b.End}
}
