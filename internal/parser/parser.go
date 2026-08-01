/*
===============================================================================
QUEST STAGE 2: THE STRUCTURAL WEAVER (Parser)
===============================================================================
Overview:
  The parser sits between the lexer and both execution backends. It consumes the
  token stream and builds an Abstract Syntax Tree (AST) that records program
  structure without executing it.

  This handwritten recursive-descent parser selects statement rules from the
  next token and uses a precedence ladder for expressions. This QUEST adds AST
  construction for branches, variables, functions, calls, returns, and names.

Tasks in this file:
  - TASK [PARSE-01]: Build if, else-if, and else statement trees.
  - TASK [PARSE-02]: Build variable declarations and assignments.
  - TASK [PARSE-03]: Build function declarations, calls, and returns.
  - TASK [PARSE-04]: Build identifier lookup expressions.

Commands:
  - Run tests:  go test ./internal/parser
  - Inspect:    ./nuru-interpreter ast examples/02-ast.nuru
  - Skip stage: ./savepoint.sh 2
  - Reset stage: ./savepoint.sh 1
===============================================================================
*/

package parser

import (
	"fmt"

	"github.com/kc-clintone/compilers/internal/ast"
	"github.com/kc-clintone/compilers/internal/diagnostic"
	"github.com/kc-clintone/compilers/internal/lexer"
	"github.com/kc-clintone/compilers/internal/source"
	"github.com/kc-clintone/compilers/internal/token"
)

// Parser holds the token stream, current token index, filename, and diagnostics
// used by recursive-descent grammar routines.
type Parser struct {
	filename string
	tokens   []token.Token
	current  int
	diags    []diagnostic.Diagnostic
}

// Parse tokenizes input and parses it into an AST Program. Lexer diagnostics
// stop parsing; parser diagnostics are returned with the partial program.
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

// program parses top-level declarations and statements until EOF.
func (p *Parser) program() *ast.Program {
	start := p.peek().Span
	prog := &ast.Program{Base: ast.Base{Span: start}}

	for !p.atEnd() {
		if p.check(token.Kind("func")) {
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

// declaration dispatches a top-level declaration from the next token.
func (p *Parser) declaration() ast.Decl {
	if p.check(token.Kind("func")) {
		return p.parseFuncDecl()
	}
	p.error("expected declaration")
	p.advance()
	return nil
}

// parseFuncDecl parses a named top-level function, its parameters, and body.
// The returned span starts at func and ends after the closing body brace.
//
// TASK [PARSE-03]: Build an ast.FuncDecl for func name(params) { body }.
// Use token.Func, token.Ident, token.LParen, token.RParen, token.Comma,
// token.LBrace, consume, check, match, parseBlockStmt, and mergeSpan.
// See HINT [PARSE-03-HINT] at the bottom of this file for details.
func (p *Parser) parseFuncDecl() ast.Decl {
	p.error("function declarations ('func') are not implemented yet in Stage 0/1")
	p.advance()
	return nil
}

// statement dispatches the next statement form. Rules that need their leading
// token for a source span consume that token inside their own parser.
func (p *Parser) statement() ast.Stmt {
	if p.check(token.Kind("var")) {
		return p.parseVarDecl()
	}
	if p.check(token.If) {
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

	if p.check(token.Ident) && p.peek().Lexeme == "return" {
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

// parseVarDecl parses var name = initializer with an optional trailing semicolon.
// Its span covers the var keyword through the initializer.
//
// TASK [PARSE-02]: Build an ast.VarDecl for var name = expression.
// Use token.Var, token.Ident, token.Assign, consume, expression, match,
// token.Semicolon, and mergeSpan.
// See HINT [PARSE-02-HINT] at the bottom of this file for details.
func (p *Parser) parseVarDecl() *ast.VarDecl {
	p.error("variable declarations ('var') are not implemented yet in Stage 0/1")
	p.advance()
	return nil
}

// parseIfStmt parses an if condition and required block plus an optional else
// block or recursively parsed else-if. Its span includes the leading if keyword
// and the final branch.
//
// TASK [PARSE-01]: Build an ast.IfStmt for if, else-if, and else branches.
// Use token.If, token.Else, token.LBrace, consume, expression, parseBlockStmt,
// check, match, mergeSpan, and recursive parseIfStmt calls.
// See HINT [PARSE-01-HINT] at the bottom of this file for details.
func (p *Parser) parseIfStmt() ast.Stmt {
	p.error("conditional branch statements ('if') are not implemented yet in Stage 0/1")
	p.advance()
	return nil
}

// parseAssignStmt parses name = value with an optional trailing semicolon. Its
// span covers the assigned identifier through the value expression.
//
// TASK [PARSE-02]: Build an ast.AssignStmt for name = expression.
// Use token.Ident, token.Assign, consume, expression, match, token.Semicolon,
// and mergeSpan.
// See HINT [PARSE-02-HINT] at the bottom of this file for details.
func (p *Parser) parseAssignStmt() ast.Stmt {
	p.error("variable assignments are not implemented yet in Stage 0/1")
	p.advance()
	p.advance()
	return nil
}

// parseReturnStmt parses return with an optional value and semicolon. Return is
// currently recognized by its identifier lexeme rather than a dedicated kind.
//
// TASK [PARSE-03]: Build an ast.ReturnStmt and preserve bare returns.
// Use advance, check, atEnd, expression, match, token.RBrace, token.Semicolon,
// and mergeSpan.
// See HINT [PARSE-03-HINT] at the bottom of this file for details.
func (p *Parser) parseReturnStmt() ast.Stmt {
	p.error("return statements are not implemented yet in Stage 0/1")
	p.advance()
	return nil
}

// parsePrintStmt parses the built-in print call as a dedicated statement node.
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

// parseForStmt parses a condition and required body after an already consumed for.
func (p *Parser) parseForStmt() ast.Stmt {
	start := p.previous()
	cond := p.expression()
	p.consume(token.LBrace, "expected '{' after for condition")
	body := p.parseBlockStmt()

	return &ast.ForStmt{Base: ast.Base{Span: mergeSpan(start.Span, body.GetSpan())}, Cond: cond, Body: body}
}

// parseBlockStmt parses statements through a closing brace. The opening brace
// must already have been consumed so previous identifies the span start.
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

// expression parses the lowest-precedence expression grammar rule.
func (p *Parser) expression() ast.Expr {
	return p.logicalOr()
}

// logicalOr parses left-associative || expressions.
func (p *Parser) logicalOr() ast.Expr {
	expr := p.logicalAnd()
	for p.match(token.Or) {
		op := p.previous().Lexeme
		right := p.logicalAnd()
		expr = &ast.BinaryExpr{Base: ast.Base{Span: mergeSpan(expr.GetSpan(), right.GetSpan())}, Left: expr, Op: op, Right: right}
	}
	return expr
}

// logicalAnd parses left-associative && expressions.
func (p *Parser) logicalAnd() ast.Expr {
	expr := p.equality()
	for p.match(token.And) {
		op := p.previous().Lexeme
		right := p.equality()
		expr = &ast.BinaryExpr{Base: ast.Base{Span: mergeSpan(expr.GetSpan(), right.GetSpan())}, Left: expr, Op: op, Right: right}
	}
	return expr
}

// equality parses left-associative == and != expressions.
func (p *Parser) equality() ast.Expr {
	expr := p.comparison()
	for p.match(token.Equal) || p.match(token.NotEqual) {
		op := p.previous().Lexeme
		right := p.comparison()
		expr = &ast.BinaryExpr{Base: ast.Base{Span: mergeSpan(expr.GetSpan(), right.GetSpan())}, Left: expr, Op: op, Right: right}
	}
	return expr
}

// comparison parses relational expressions after their operands are parsed.
func (p *Parser) comparison() ast.Expr {
	expr := p.additive()
	for p.match(token.Less) || p.match(token.LessEqual) || p.match(token.Greater) || p.match(token.GreaterEqual) {
		op := p.previous().Lexeme
		right := p.additive()
		expr = &ast.BinaryExpr{Base: ast.Base{Span: mergeSpan(expr.GetSpan(), right.GetSpan())}, Left: expr, Op: op, Right: right}
	}
	return expr
}

// additive parses left-associative + and - expressions.
func (p *Parser) additive() ast.Expr {
	expr := p.multiplicative()
	for p.match(token.Plus) || p.match(token.Minus) {
		op := p.previous().Lexeme
		right := p.multiplicative()
		expr = &ast.BinaryExpr{Base: ast.Base{Span: mergeSpan(expr.GetSpan(), right.GetSpan())}, Left: expr, Op: op, Right: right}
	}
	return expr
}

// multiplicative parses left-associative *, /, and % expressions.
func (p *Parser) multiplicative() ast.Expr {
	expr := p.unary()
	for p.match(token.Star) || p.match(token.Slash) || p.match(token.Percent) {
		op := p.previous().Lexeme
		right := p.unary()
		expr = &ast.BinaryExpr{Base: ast.Base{Span: mergeSpan(expr.GetSpan(), right.GetSpan())}, Left: expr, Op: op, Right: right}
	}
	return expr
}

// unary parses prefix - and ! expressions before falling through to primary.
func (p *Parser) unary() ast.Expr {
	if p.match(token.Minus) || p.match(token.Bang) {
		opTok := p.previous()
		right := p.unary()
		return &ast.UnaryExpr{Base: ast.Base{Span: mergeSpan(opTok.Span, right.GetSpan())}, Op: opTok.Lexeme, Right: right}
	}
	return p.primary()
}

// primary parses literals, grouped expressions, identifiers, and calls.
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
		if p.peekNext().Kind == token.LParen {
			return p.parseCallExpr()
		}
		return p.parseIdentExpr()
	}

	p.error(fmt.Sprintf("unexpected token '%s'", p.peek().Lexeme))
	p.advance()
	return nil
}

// parseIdentExpr consumes a name and preserves its spelling and source span in
// an ast.IdentExpr.
//
// TASK [PARSE-04]: Build an ast.IdentExpr for the next identifier.
// Use token.Ident and consume.
// See HINT [PARSE-04-HINT] at the bottom of this file for details.
func (p *Parser) parseIdentExpr() ast.Expr {
	tok := p.advance()
	p.error(fmt.Sprintf("identifier expressions ('%s') are not implemented yet in Stage 0/1", tok.Lexeme))
	return nil
}

// parseCallExpr parses a named callee and zero or more comma-separated
// arguments, preserving the span through the closing parenthesis.
//
// TASK [PARSE-03]: Build an ast.CallExpr for name(arguments).
// Use token.Ident, token.LParen, token.RParen, token.Comma, consume, check,
// expression, match, and mergeSpan.
// See HINT [PARSE-03-HINT] at the bottom of this file for details.
func (p *Parser) parseCallExpr() ast.Expr {
	tok := p.advance()
	p.error(fmt.Sprintf("function calls ('%s(...)') are not implemented yet in Stage 0/1", tok.Lexeme))
	p.consume(token.LParen, "")
	p.consume(token.RParen, "")
	return nil
}

// atEnd reports whether the next token is EOF.
func (p *Parser) atEnd() bool { return p.peek().Kind == token.EOF }

// peek returns the next token without consuming it. Once past the slice it
// returns the final EOF token.
func (p *Parser) peek() token.Token {
	if p.current >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.current]
}

// peekNext returns the token after peek without consuming either token. At the
// boundary it returns the final EOF token.
func (p *Parser) peekNext() token.Token {
	if p.current+1 >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.current+1]
}

// previous returns the most recently consumed token. Callers use it only after
// a successful advance, match, or consume.
func (p *Parser) previous() token.Token { return p.tokens[p.current-1] }

// check reports whether the next non-EOF token has kind k without consuming it.
func (p *Parser) check(k token.Kind) bool {
	if p.atEnd() {
		return false
	}
	return p.peek().Kind == k
}

// advance consumes and returns the next token, stopping at EOF.
func (p *Parser) advance() token.Token {
	if !p.atEnd() {
		p.current++
	}
	return p.previous()
}

// match consumes kind k when it is next and reports whether it matched.
func (p *Parser) match(k token.Kind) bool {
	if p.check(k) {
		p.advance()
		return true
	}
	return false
}

// consume returns the next token when it has kind k. On mismatch it records msg
// at the current token and returns that token without consuming it.
func (p *Parser) consume(k token.Kind, msg string) token.Token {
	if p.check(k) {
		return p.advance()
	}
	p.error(msg)
	return p.peek()
}

// error records a parser diagnostic at the next token.
func (p *Parser) error(msg string) {
	p.diags = append(p.diags, diagnostic.Diagnostic{Span: p.peek().Span, Phase: "parser", Message: msg})
}

// mergeSpan returns a half-open span from a's start through b's end.
func mergeSpan(a, b source.Span) source.Span {
	return source.Span{Filename: a.Filename, Start: a.Start, End: b.End}
}

/*
===============================================================================
QUEST HINTS
===============================================================================
HINT [PARSE-01-HINT]:
  1. Consume token.If and keep that token as the span start.
  2. Parse the condition, require an opening brace, and parse the then block.
  3. If token.Else follows, either parse another if recursively or consume an
     opening brace and parse the else block. Diagnose any other token.
  4. Extend the span to the last parsed branch, falling back to the then block.
  5. Return an ast.IfStmt containing the condition and both branch fields.

HINT [PARSE-02-HINT]:
  Variable declaration:
  1. Consume token.Var, the variable name, and token.Assign in order.
  2. Parse the initializer and optionally consume token.Semicolon.
  3. Merge the var token's span through the initializer and return ast.VarDecl.

  Assignment:
  1. Consume the identifier and token.Assign, then parse the value.
  2. Optionally consume token.Semicolon.
  3. Merge the identifier span through the value and return ast.AssignStmt.

HINT [PARSE-03-HINT]:
  Function declaration:
  1. Consume func, the function name, and the opening parenthesis.
  2. Until the closing parenthesis, consume comma-separated parameter names.
  3. Require the closing parenthesis and opening brace, then parse the body.
  4. Return ast.FuncDecl spanning from func through the body.

  Function call:
  1. Consume the callee and opening parenthesis.
  2. Parse comma-separated expressions unless the argument list is empty.
  3. Consume the closing parenthesis and return an ast.CallExpr spanning it.

  Return:
  1. Consume the return lexeme.
  2. Parse a value unless the next token closes the statement, block, or file.
  3. Optionally consume a semicolon and return ast.ReturnStmt with the full span.

HINT [PARSE-04-HINT]:
  1. Consume token.Ident with a useful diagnostic message.
  2. Copy its span into ast.Base and its lexeme into Name.
  3. Return the new ast.IdentExpr.
===============================================================================
*/
