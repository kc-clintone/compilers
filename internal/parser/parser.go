/*
===============================================================================
QUEST STAGE 2: THE STRUCTURAL WEAVER (Parser)
===============================================================================
Overview:
  Parse a stream of tokens into an Abstract Syntax Tree (AST) using Recursive Descent.
  In this stage, you will implement parsing rules for branches, variables, functions,
  and identifier expressions.

Tasks in this file:
  - TASK [PARSE-01]: Implement parseIfStmt() for conditional branches.
                     Uses ast.IfStmt, token.If, token.Else, parseBlockStmt().
  - TASK [PARSE-02]: Implement parseVarDecl() and parseAssignStmt() for variable
                     declarations and assignments.
                     Uses ast.VarDecl, ast.AssignStmt, token.Var, token.Assign.
  - TASK [PARSE-03]: Implement parseFuncDecl(), parseCallExpr(), and parseReturnStmt()
                     for function declarations, invocations, and returns.
                     Uses ast.FuncDecl, ast.CallExpr, ast.ReturnStmt, token.Func.
  - TASK [PARSE-04]: Implement parseIdentExpr() for variable identifier lookups.
                     Uses ast.IdentExpr and token.Ident.

Commands:
  - Run tests:  go test ./internal/parser
  - Inspect:    ./zing ast examples/02-ast.zing
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

func (p *Parser) declaration() ast.Decl {
	if p.check(token.Kind("func")) {
		return p.parseFuncDecl()
	}
	p.error("expected declaration")
	p.advance()
	return nil
}

// TASK [PARSE-03]: Implement parseFuncDecl() for function declarations: func <name>(<params...>) { <body> }
// Uses ast.FuncDecl, token.Func, token.LParen, token.RParen, token.Comma, token.LBrace, parseBlockStmt().
// See HINT [PARSE-03-HINT] at the bottom of this file for details.
func (p *Parser) parseFuncDecl() ast.Decl {
	p.error("function declarations ('func') are not implemented yet in Stage 0/1")
	p.advance()
	return nil
}

// TASK [PARSE-02]: Implement parseVarDecl() for variable declarations: var <name> = <expr>
// Uses ast.VarDecl, token.Var, token.Ident, token.Assign, expression().
// See HINT [PARSE-02-HINT] at the bottom of this file for details.
func (p *Parser) parseVarDecl() *ast.VarDecl {
	p.error("variable declarations ('var') are not implemented yet in Stage 0/1")
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

func (p *Parser) parseVarDeclAfterKeyword() *ast.VarDecl {
	p.error("variable declarations ('var') are not implemented yet in Stage 0/1")
	return nil
}

// TASK [PARSE-01]: Implement parseIfStmt() for conditional branch statements: if (<cond>) { <then> } else { <else> }
// Uses ast.IfStmt, token.If, token.Else, expression(), parseBlockStmt().
// See HINT [PARSE-01-HINT] at the bottom of this file for details.
func (p *Parser) parseIfStmt() ast.Stmt {
	p.error("conditional branch statements ('if') are not implemented yet in Stage 0/1")
	return nil
}

// TASK [PARSE-02]: Implement parseAssignStmt() for variable assignment: <name> = <expr>
// Uses ast.AssignStmt, token.Ident, token.Assign, expression().
// See HINT [PARSE-02-HINT] at the bottom of this file for details.
func (p *Parser) parseAssignStmt() ast.Stmt {
	p.error("variable assignments are not implemented yet in Stage 0/1")
	p.advance()
	p.advance()
	return nil
}

// TASK [PARSE-03]: Implement parseReturnStmt() for return statements: return <expr>
// Uses ast.ReturnStmt, expression().
// See HINT [PARSE-03-HINT] at the bottom of this file for details.
func (p *Parser) parseReturnStmt() ast.Stmt {
	p.error("return statements are not implemented yet in Stage 0/1")
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
		if p.peekNext().Kind == token.LParen {
			return p.parseCallExpr()
		}
		return p.parseIdentExpr()
	}

	p.error(fmt.Sprintf("unexpected token '%s'", p.peek().Lexeme))
	p.advance()
	return nil
}

// TASK [PARSE-04]: Implement parseIdentExpr() for variable lookups.
// Uses ast.IdentExpr and token.Ident.
// See HINT [PARSE-04-HINT] at the bottom of this file for details.
func (p *Parser) parseIdentExpr() ast.Expr {
	tok := p.advance()
	p.error(fmt.Sprintf("identifier expressions ('%s') are not implemented yet in Stage 0/1", tok.Lexeme))
	return nil
}

// TASK [PARSE-03]: Implement parseCallExpr() for function call expressions: <callee>(<args...>)
// Uses ast.CallExpr, token.Ident, token.LParen, token.RParen, token.Comma.
// See HINT [PARSE-03-HINT] at the bottom of this file for details.
func (p *Parser) parseCallExpr() ast.Expr {
	tok := p.advance()
	p.error(fmt.Sprintf("function calls ('%s(...)') are not implemented yet in Stage 0/1", tok.Lexeme))
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

/*
===============================================================================
QUEST HINTS & SOLUTIONS
===============================================================================
HINT [PARSE-01-HINT]:
  Implement parseIfStmt:
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

HINT [PARSE-02-HINT]:
  Implement parseVarDecl, parseVarDeclAfterKeyword, and parseAssignStmt:
    func (p *Parser) parseVarDecl() *ast.VarDecl {
        p.consume(token.Var, "expected 'var'")
        return p.parseVarDeclAfterKeyword()
    }

    func (p *Parser) parseVarDeclAfterKeyword() *ast.VarDecl {
        start := p.previous()
        nameTok := p.consume(token.Ident, "expected variable name after 'var'")
        p.consume(token.Assign, "expected '=' after variable name")
        init := p.expression()
        p.match(token.Semicolon)

        endSpan := start.Span
        if init != nil {
            endSpan = mergeSpan(start.Span, init.GetSpan())
        }

        return &ast.VarDecl{Base: ast.Base{Span: endSpan}, Name: nameTok.Lexeme, Init: init}
    }

    func (p *Parser) parseAssignStmt() ast.Stmt {
        nameTok := p.consume(token.Ident, "expected variable name")
        p.consume(token.Assign, "expected '='")
        val := p.expression()
        p.match(token.Semicolon)

        endSpan := nameTok.Span
        if val != nil {
            endSpan = mergeSpan(nameTok.Span, val.GetSpan())
        }

        return &ast.AssignStmt{Base: ast.Base{Span: endSpan}, Name: nameTok.Lexeme, Value: val}
    }

HINT [PARSE-03-HINT]:
  Implement parseFuncDecl, parseCallExpr, and parseReturnStmt:
    func (p *Parser) parseFuncDecl() ast.Decl {
        start := p.consume(token.Func, "expected 'func'")
        nameTok := p.consume(token.Ident, "expected function name")
        p.consume(token.LParen, "expected '(' after function name")

        var params []string
        if !p.check(token.RParen) {
            for {
                paramTok := p.consume(token.Ident, "expected parameter name")
                params = append(params, paramTok.Lexeme)
                if !p.match(token.Comma) {
                    break
                }
            }
        }
        p.consume(token.RParen, "expected ')' after parameters")
        p.consume(token.LBrace, "expected '{' before function body")
        body := p.parseBlockStmt()

        return &ast.FuncDecl{
            Base:   ast.Base{Span: mergeSpan(start.Span, body.GetSpan())},
            Name:   nameTok.Lexeme,
            Params: params,
            Body:   body,
        }
    }

    func (p *Parser) parseCallExpr() ast.Expr {
        calleeTok := p.consume(token.Ident, "expected function name")
        p.consume(token.LParen, "expected '('")

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
        endTok := p.consume(token.RParen, "expected ')' after arguments")

        return &ast.CallExpr{
            Base:   ast.Base{Span: mergeSpan(calleeTok.Span, endTok.Span)},
            Callee: calleeTok.Lexeme,
            Args:   args,
        }
    }

    func (p *Parser) parseReturnStmt() ast.Stmt {
        start := p.advance()
        var val ast.Expr
        if !p.check(token.RBrace) && !p.check(token.Semicolon) && !p.atEnd() {
            val = p.expression()
        }
        p.match(token.Semicolon)

        span := start.Span
        if val != nil {
            span = mergeSpan(start.Span, val.GetSpan())
        }
        return &ast.ReturnStmt{Base: ast.Base{Span: span}, Value: val}
    }

HINT [PARSE-04-HINT]:
  Implement parseIdentExpr:
    func (p *Parser) parseIdentExpr() ast.Expr {
        tok := p.consume(token.Ident, "expected identifier")
        return &ast.IdentExpr{Base: ast.Base{Span: tok.Span}, Name: tok.Lexeme}
    }
===============================================================================
*/
