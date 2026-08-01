# Compilers and Interpreters in Go: Building Nuru

Follow Nuru source through a lexer and recursive-descent parser, then execute
the same Abstract Syntax Tree (AST) in two ways: directly with a tree-walking
interpreter and indirectly by transpiling it into readable Go.

The workshop has four 15-minute QUESTS:

1. **The Lexical Conduit:** turn source bytes into tokens.
2. **The Structural Weaver:** turn tokens into an AST.
3. **The Living Engine:** walk the AST and produce values.
4. **The Code Forge:** turn the AST into Go and a native executable.

The styled version of this guide is
[`compilers-and-interpreters-in-go.html`](compilers-and-interpreters-in-go.html).
Both guides contain the same complete solutions.

## The pipeline

```text
Nuru source (.nuru) -> lexer -> recursive-descent parser -> AST
                                                        |-> interpreter
                                                        `-> Go transpiler -> go build
```

The lexer gives meaning to byte sequences by emitting tokens. The parser gives
those tokens structure without running them. The interpreter and transpiler
then consume the same AST, which keeps syntax decisions separate from execution.

This workshop implements a small dynamically typed subset of Nuru: identifiers,
variables, branches, top-level functions, calls, and returns. It does not add a
bytecode format, virtual machine, assembly backend, closures, or static types.
The larger, bootstrap-ready reference implementation remains available on the
`lugha` branch for later study.

## Map and equipment

- Install Go 1.22 or newer.
- Be comfortable reading Go structs, pointers, slices, maps, interfaces, and
  multiple packages.
- Start from `main` or a branch based on it.
- Do not create new learner solution files; every TASK location already exists.

Build both tools from the repository root:

```console
$ go build -o ./nuru-interpreter ./cmd/nuru-interpreter
$ go build -o ./nuru-compiler ./cmd/nuru-compiler
```

A successful build is silent. The interpreter executes files directly or opens
a REPL when invoked without arguments:

```text
nuru-interpreter(1)

NAME
    nuru-interpreter - inspect and directly execute Nuru programs

SYNOPSIS
    nuru-interpreter
    nuru-interpreter FILE
    nuru-interpreter tokens FILE
    nuru-interpreter ast FILE
    nuru-interpreter repl
    nuru-interpreter (-h | --help)

DESCRIPTION
    Runs a Nuru source file directly. Invoking the interpreter without a file,
    or with the repl command, starts an interactive read-eval-print loop.

COMMANDS
    FILE
        Execute FILE directly.

    tokens FILE, lex FILE
        Print the tokens produced by the lexer. "lex" is an alias for "tokens".

    ast FILE, parse FILE
        Print the syntax tree produced by the parser. "parse" is an alias for
        "ast".

    repl
        Start the interactive REPL.

    (no arguments)
        Start the interactive REPL.

OPTIONS
    -h, --help
        Print this help page and exit.
```

The compiler exposes the same front-end stages and can emit Go or build a
native executable:

```text
nuru-compiler(1)

NAME
    nuru-compiler - inspect, transpile, and build Nuru programs

SYNOPSIS
    nuru-compiler [-o BINARY] FILE
    nuru-compiler tokens FILE
    nuru-compiler ast FILE
    nuru-compiler transpile [-o GO-FILE] FILE
    nuru-compiler (-h | --help)

DESCRIPTION
    Compiles a Nuru source file to a native executable by default. The same
    command can expose intermediate compiler stages or emit Go source.

COMMANDS
    FILE
        Build FILE as a native executable. This is the default operation.

    tokens FILE, lex FILE
        Print the tokens produced by the lexer. "lex" is an alias for "tokens".

    ast FILE, parse FILE
        Print the syntax tree produced by the parser. "parse" is an alias for
        "ast".

    transpile [-o GO-FILE] FILE
        Write generated Go source. The default output is <input-basename>.go.

OPTIONS
    -o PATH
        Set the generated Go file or executable path. The default executable is
        ./nuru.out.

    -h, --help
        Print this help page and exit.
```

## Follow the QUEST loop

For every stage:

1. Read the guide's stage overview.
2. Find the matching QUEST banner at the top of the source file.
3. Read the documented behavior of each learner-implemented function.
4. Complete its high-level TASK, using the HINT only when needed.
5. Compare with the complete guide solution if you become stuck.
6. Run the stage test and checkpoint command.

Find all banners with:

```console
$ grep -rn "QUEST STAGE" internal
internal/interpreter/interpreter.go:3:QUEST STAGE 3: THE LIVING ENGINE (Interpreter)
internal/compiler/compiler.go:3:QUEST STAGE 4: THE CODE FORGE (Compiler / Transpiler)
internal/lexer/lexer.go:3:QUEST STAGE 1: THE LEXICAL CONDUIT (Lexer)
internal/parser/parser.go:3:QUEST STAGE 2: THE STRUCTURAL WEAVER (Parser)
internal/token/token.go:3:QUEST STAGE 1: THE LEXICAL CONDUIT (Token Vocabulary)
```

Search by TASK identifier rather than relying only on line numbers. The line
numbers below refer to the `0-base` starter checkpoint and are included to help
you orient yourself before making edits.

## QUEST 1: The Lexical Conduit

The lexer is the first compiler stage. It advances through the source one byte
at a time and groups related bytes into tokens. Tokens preserve a kind, the
original lexeme, an optional decoded literal, and a source span. The parser
never needs to classify characters itself.

This QUEST makes the lexer recognize integer digits and names, distinguish
identifiers from reserved words, and add `var` and `func` to the shared token
vocabulary.

### TASK LEX-01: Classify digits and identifier characters

In `internal/lexer/lexer.go`, replace `isDigit` and `isIdentStart` beginning at
starter line 281. A name begins with an ASCII letter or underscore and may use
digits only after that first character.

```go
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isIdentStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
```

### TASK LEX-02: Emit identifiers and keywords

Replace `identifier` at starter line 227. Use `token.Keywords` and `l.add`.
Preserve `true` and `false` as Go `bool` literals because later stages evaluate
their token literals directly.

```go
func (l *Lexer) identifier() {
	for isIdentPart(l.peek()) {
		l.advance()
	}
	raw := string(l.src[l.start:l.current])

	if kind, ok := token.Keywords[raw]; ok {
		switch kind {
		case token.True:
			l.add(kind, true)
		case token.False:
			l.add(kind, false)
		default:
			l.add(kind, nil)
		}
		return
	}
	l.add(token.Ident, raw)
}
```

### TASK LEX-03: Add `var` and `func`

In `internal/token/token.go`, add the kinds at starter line 70:

```go
Var  Kind = "var"
Func Kind = "func"
```

Then register their spellings in `Keywords` at starter line 109:

```go
"var":  Var,
"func": Func,
```

Verify the stage:

```console
$ go test ./internal/token ./internal/lexer
$ go build -o ./nuru-interpreter ./cmd/nuru-interpreter
$ ./nuru-interpreter tokens examples/01-tokens.nuru
POSITION     KIND       LITERAL      LEXEME
---------------------------------------------------------
1:1          var                     var
1:5          IDENT      x            x
1:7          =                       =
1:9          INT        10           10
...
5:1          EOF
```

## QUEST 2: The Structural Weaver

The parser consumes the lexer token stream and records program structure in an
AST. Statement routines recognize declarations and control flow. Expression
routines form a precedence ladder from logical OR down to primary expressions,
so multiplication nests more tightly than addition.

The parser constructs nodes and source spans only; it does not decide runtime
values. This QUEST adds branches, variables, functions, calls, returns, and
identifier lookup nodes.

### TASK PARSE-01: Parse branches

Replace `parseIfStmt` in `internal/parser/parser.go` at starter line 178. The
function consumes its own `if`, parses a required block, and recursively handles
`else if` so every branch has the correct source span.

```go
func (p *Parser) parseIfStmt() ast.Stmt {
	start := p.consume(token.If, "expected 'if'")
	cond := p.expression()
	p.consume(token.LBrace, "expected '{' after if condition")
	thenBranch := p.parseBlockStmt()

	var elseBranch ast.Stmt
	if p.match(token.Else) {
		if p.check(token.If) {
			elseBranch = p.parseIfStmt()
		} else if p.match(token.LBrace) {
			elseBranch = p.parseBlockStmt()
		} else {
			p.error("expected '{' or 'if' after 'else'")
		}
	}

	span := mergeSpan(start.Span, thenBranch.GetSpan())
	if elseBranch != nil {
		span = mergeSpan(start.Span, elseBranch.GetSpan())
	}
	return &ast.IfStmt{
		Base: ast.Base{Span: span},
		Cond: cond,
		Then: thenBranch,
		Else: elseBranch,
	}
}
```

### TASK PARSE-02: Parse variables and assignments

Replace `parseVarDecl` at starter line 164 and `parseAssignStmt` at line 191.
Both accept an optional semicolon and extend their span through the value.

```go
func (p *Parser) parseVarDecl() *ast.VarDecl {
	start := p.consume(token.Var, "expected 'var'")
	name := p.consume(token.Ident, "expected variable name after 'var'")
	p.consume(token.Assign, "expected '=' after variable name")
	init := p.expression()
	p.match(token.Semicolon)

	span := start.Span
	if init != nil {
		span = mergeSpan(start.Span, init.GetSpan())
	}
	return &ast.VarDecl{
		Base: ast.Base{Span: span},
		Name: name.Lexeme,
		Init: init,
	}
}
```

```go
func (p *Parser) parseAssignStmt() ast.Stmt {
	name := p.consume(token.Ident, "expected variable name")
	p.consume(token.Assign, "expected '='")
	value := p.expression()
	p.match(token.Semicolon)

	span := name.Span
	if value != nil {
		span = mergeSpan(name.Span, value.GetSpan())
	}
	return &ast.AssignStmt{
		Base:  ast.Base{Span: span},
		Name:  name.Lexeme,
		Value: value,
	}
}
```

### TASK PARSE-03: Parse functions, calls, and returns

Replace `parseFuncDecl` at starter line 108, `parseReturnStmt` at line 205, and
`parseCallExpr` at line 385. Parameters and arguments are comma-separated;
return values are optional.

```go
func (p *Parser) parseFuncDecl() ast.Decl {
	start := p.consume(token.Func, "expected 'func'")
	name := p.consume(token.Ident, "expected function name")
	p.consume(token.LParen, "expected '(' after function name")

	var params []string
	if !p.check(token.RParen) {
		for {
			param := p.consume(token.Ident, "expected parameter name")
			params = append(params, param.Lexeme)
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
		Name:   name.Lexeme,
		Params: params,
		Body:   body,
	}
}
```

```go
func (p *Parser) parseReturnStmt() ast.Stmt {
	start := p.advance()
	var value ast.Expr
	if !p.check(token.RBrace) && !p.check(token.Semicolon) && !p.atEnd() {
		value = p.expression()
	}
	p.match(token.Semicolon)

	span := start.Span
	if value != nil {
		span = mergeSpan(start.Span, value.GetSpan())
	}
	return &ast.ReturnStmt{Base: ast.Base{Span: span}, Value: value}
}
```

```go
func (p *Parser) parseCallExpr() ast.Expr {
	callee := p.consume(token.Ident, "expected function name")
	p.consume(token.LParen, "expected '('")

	var args []ast.Expr
	if !p.check(token.RParen) {
		for {
			if arg := p.expression(); arg != nil {
				args = append(args, arg)
			}
			if !p.match(token.Comma) {
				break
			}
		}
	}
	end := p.consume(token.RParen, "expected ')' after arguments")
	return &ast.CallExpr{
		Base:   ast.Base{Span: mergeSpan(callee.Span, end.Span)},
		Callee: callee.Lexeme,
		Args:   args,
	}
}
```

### TASK PARSE-04: Parse identifier lookups

Replace `parseIdentExpr` at starter line 372:

```go
func (p *Parser) parseIdentExpr() ast.Expr {
	tok := p.consume(token.Ident, "expected identifier")
	return &ast.IdentExpr{
		Base: ast.Base{Span: tok.Span},
		Name: tok.Lexeme,
	}
}
```

Verify the stage:

```console
$ go test ./internal/parser
$ ./nuru-interpreter ast examples/02-ast.nuru
Program
  FuncDecl: max(a, b)
    BlockStmt:
      IfStmt:
        Cond:
          BinaryExpr (>):
            Ident: a
            Ident: b
        Then:
          BlockStmt:
            ReturnStmt:
              Ident: a
      ReturnStmt:
        Ident: b
  VarDecl: res
    CallExpr: max()
      Literal: 10
      Literal: 20
  PrintStmt:
    Ident: res
```

## QUEST 3: The Living Engine

The interpreter walks the AST directly. Environments map identifier names to
runtime values, with parent links for outward lookup. Top-level functions are
registered before statements run, and every call receives a fresh environment
whose parent is global state.

### TASK EVAL-01: Choose a branch

Replace `evalIfStmt` in `internal/interpreter/interpreter.go` at starter line 236:

```go
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
```

### TASK EVAL-02: Store, update, and retrieve variables

Replace `evalVarDecl` at starter line 151, `evalAssignStmt` at line 246, and
`evalIdentExpr` at line 373:

```go
func (in *Interpreter) evalVarDecl(d *ast.VarDecl, env *Environment) {
	var value any
	if d.Init != nil {
		value = in.evalExpr(d.Init, env)
	}
	env.Define(d.Name, value)
}
```

```go
func (in *Interpreter) evalAssignStmt(s *ast.AssignStmt, env *Environment) any {
	value := in.evalExpr(s.Value, env)
	if !env.Set(s.Name, value) {
		in.error(s.GetSpan(), fmt.Sprintf(
			"cannot assign to undefined variable '%s'", s.Name,
		))
	}
	return nil
}
```

```go
func (in *Interpreter) evalIdentExpr(e *ast.IdentExpr, env *Environment) any {
	value, ok := env.Get(e.Name)
	if !ok {
		in.error(e.GetSpan(), fmt.Sprintf("undefined variable '%s'", e.Name))
		return nil
	}
	return value
}
```

### TASK EVAL-03: Register and call functions

Replace `evalFuncDecl` at starter line 142, `evalReturnStmt` at line 256, and
`evalCallExpr` at line 385:

```go
func (in *Interpreter) evalFuncDecl(d *ast.FuncDecl, env *Environment) {
	in.funcs[d.Name] = &UserFunction{Decl: d}
}
```

```go
func (in *Interpreter) evalReturnStmt(s *ast.ReturnStmt, env *Environment) any {
	var value any
	if s.Value != nil {
		value = in.evalExpr(s.Value, env)
	}
	return ReturnVal{Value: value}
}
```

```go
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
```

Verify the stage:

```console
$ go test ./internal/interpreter
$ ./nuru-interpreter examples/03-interpreter.nuru
Factorial of 5: 120
```

## QUEST 4: The Code Forge

The transpiler walks the same AST but writes Go instead of producing values.
Generated helpers preserve Nuru's dynamic truthiness and operators while Go
provides parsing, machine-code generation, and executable linking.

### TASK GEN-01: Generate branches

Replace `compileIfStmt` in `internal/compiler/compiler.go` at starter line 170:

```go
func (c *Compiler) compileIfStmt(s *ast.IfStmt, indent string) {
	c.emit("%sif nuruTruthy(", indent)
	c.compileExpr(s.Cond)
	c.emit(")")
	c.compileStmt(s.Then, indent)
	if s.Else != nil {
		c.emit(" else")
		if _, isIf := s.Else.(*ast.IfStmt); isIf {
			c.compileStmt(s.Else, "")
		} else {
			c.compileStmt(s.Else, indent)
		}
	}
	c.emit("\n")
}
```

### TASK GEN-02: Generate variables and identifiers

Replace `compileVarDecl` at starter line 100, `compileAssignStmt` at line 178,
and `compileIdentExpr` at line 236:

```go
func (c *Compiler) compileVarDecl(d *ast.VarDecl, indent string) {
	c.emit("%svar %s any", indent, d.Name)
	if d.Init != nil {
		c.emit(" = ")
		c.compileExpr(d.Init)
	}
	c.emit("\n")
}
```

```go
func (c *Compiler) compileAssignStmt(s *ast.AssignStmt, indent string) {
	c.emit("%s%s = ", indent, s.Name)
	c.compileExpr(s.Value)
	c.emit("\n")
}
```

```go
func (c *Compiler) compileIdentExpr(e *ast.IdentExpr) {
	c.emit("%s", e.Name)
}
```

### TASK GEN-03: Generate functions, calls, and returns

Replace `compileFuncDecl` at starter line 90, `compileReturnStmt` at line 187,
and `compileCallExpr` at line 244:

```go
func (c *Compiler) compileFuncDecl(d *ast.FuncDecl) {
	c.emit("func %s(", d.Name)
	for index, parameter := range d.Params {
		if index > 0 {
			c.emit(", ")
		}
		c.emit("%s any", parameter)
	}
	c.emit(") any")
	c.compileStmt(d.Body, "")
	c.emit("\n\n")
}
```

```go
func (c *Compiler) compileReturnStmt(s *ast.ReturnStmt, indent string) {
	c.emit("%sreturn ", indent)
	if s.Value != nil {
		c.compileExpr(s.Value)
	} else {
		c.emit("nil")
	}
	c.emit("\n")
}
```

```go
func (c *Compiler) compileCallExpr(e *ast.CallExpr) {
	c.emit("%s(", e.Callee)
	for index, argument := range e.Args {
		if index > 0 {
			c.emit(", ")
		}
		c.compileExpr(argument)
	}
	c.emit(")")
}
```

Verify the stage:

```console
$ go test ./internal/compiler
$ ./nuru-compiler transpile examples/04-compiled.nuru
Transpiled Go code written to 04-compiled.nuru.go
$ ./nuru-compiler -o ./04-compiled examples/04-compiled.nuru
Transpiled Go code written to /tmp/nuru-build-.../main.go
Compiled binary output written to ./04-compiled
$ ./04-compiled
Fibonacci of 10: 55
```

The generated program's focused excerpt is:

```go
func fib(n any) any {
	if nuruTruthy(nuruBinary("<=", n, 1)) {
		return n
	}
	return nuruBinary("+", fib(nuruBinary("-", n, 1)), fib(nuruBinary("-", n, 2)))
}
```

## Checkpoints and savepoints

The stages are recorded as annotated Git tags:

```text
0-base          Nuru workshop starter — all QUEST tasks ready
1-lexer         Nuru lexer QUEST complete — digits, identifiers, and var/func tokens
2-parser        Nuru parser QUEST complete — branches, variables, functions, calls, and returns
3-interpreter   Nuru interpreter QUEST complete — environments, branches, and function call frames
4-compiler      Nuru transpiler QUEST complete — Go code generation
```

Jump to a checkpoint with:

```console
$ ./savepoint.sh 0  # starter scaffold
$ ./savepoint.sh 1  # lexer complete
$ ./savepoint.sh 2  # parser complete
$ ./savepoint.sh 3  # interpreter complete
$ ./savepoint.sh 4  # transpiler complete
```

`savepoint.sh` runs `git reset --hard` and `git clean -fd`. It deliberately
discards tracked and untracked workshop progress, so commit anything you want
to keep before using it.

At any checkpoint, run the complete package suite with:

```console
$ go test ./...
```

## Continue the QUEST

After completing the scaffold, possible next steps include differential tests
between the interpreter and generated binaries, static types and symbol
checking, collections and structs, file I/O, self-hosting, modules, formatting,
an LSP, an intermediate representation, or a later native backend. The `lugha`
branch contains a broader language reference, roadmap, and teaching guide for
those investigations.
