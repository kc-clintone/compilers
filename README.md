# Compilers Workshop Scaffold: Building Nuru

Welcome to the **Building a Compiler and Interpreter in Go** workshop. You will
complete four short **QUESTS** that add identifiers, variables, branches, and
functions to Nuru, a small dynamically typed teaching language.

## Pipeline

```text
Nuru source (.nuru) -> lexer -> recursive-descent parser / AST
                                      |-> tree-walking interpreter
                                      `-> Go source transpiler -> go build
```

The transpiler emits ordinary Go source. This workshop does not introduce a
bytecode format, virtual machine, assembly backend, or closures.

## Before the first QUEST

- Install Go 1.22 or newer.
- Be comfortable reading Go that uses structs, pointers, slices, maps,
  interfaces, and multiple packages.
- Bring a terminal, an editor, curiosity, and a sense of adventure.

Every file needed for the workshop is already provided. Start from `main` or a
branch based on it; do not create additional files or directories. A larger,
statically typed, bootstrap-ready Nuru reference implementation is available
on the `lugha` branch for study after the workshop.

## Build and use the tools

Build both command-line programs from the repository root:

```console
$ go build -o ./nuru-interpreter ./cmd/nuru-interpreter
$ go build -o ./nuru-compiler ./cmd/nuru-compiler
```

A successful build is silent. Inspect the supported commands with `--help`:

```console
$ ./nuru-interpreter --help
Nuru Interpreter (Workshop Edition)

Usage:
  nuru-interpreter <file.nuru>         Run file using interpreter (default)
  nuru-interpreter tokens <file.nuru>  Inspect Lexer token stream
  nuru-interpreter ast <file.nuru>     Inspect Parser AST tree
  nuru-interpreter repl                Start interactive REPL
  nuru-interpreter -h, --help          Show help message
```

Running `./nuru-interpreter` without arguments also opens the REPL. The
compiler can inspect the same front-end stages, print Go, or build a binary:

```console
$ ./nuru-compiler --help
Nuru Compiler (Workshop Edition)

Usage:
  nuru-compiler [-o binary] <file.nuru>   Compile Nuru program to native executable (default)
  nuru-compiler tokens <file.nuru>         Inspect Lexer token stream
  nuru-compiler ast <file.nuru>            Inspect Parser AST tree
  nuru-compiler transpile [-o out.go] <f> Transpile Nuru program to Go source
  nuru-compiler -h, --help                 Show help message
```

Each QUEST introduces the feature needed by its matching example. The guide
shows when to run these commands:

```console
$ ./nuru-interpreter tokens examples/01-tokens.nuru
$ ./nuru-interpreter ast examples/02-ast.nuru
$ ./nuru-interpreter examples/03-interpreter.nuru
Factorial of 5: 120

$ ./nuru-compiler -o ./04-compiled examples/04-compiled.nuru
Transpiled Go code written to /tmp/nuru-build-.../main.go
Compiled binary output written to ./04-compiled
$ ./04-compiled
Fibonacci of 10: 55
```

## How the workshop works

Use this README or
[`compilers-and-interpreters-in-go.html`](compilers-and-interpreters-in-go.html)
as your guide, then repeat this loop for every stage:

1. Find the next QUEST banner in stage order.
2. Read the complete QUEST instructions at the top of its file.
3. Search for each TASK by its unique identifier, such as `LEX-01`.
4. Complete and test only those TASK locations.
5. Optionally delete the completed `TASK [...]` marker so future searches show
   the remaining work.
6. Return to the guide, verify the checkpoint output, and begin the next QUEST.

Locate all QUEST banners with:

```console
$ grep -rn "QUEST STAGE" internal
internal/interpreter/interpreter.go:3:QUEST STAGE 3: THE LIVING ENGINE (Interpreter)
internal/compiler/compiler.go:3:QUEST STAGE 4: THE CODE FORGE (Compiler / Transpiler)
internal/lexer/lexer.go:3:QUEST STAGE 1: THE LEXICAL CONDUIT (Lexer)
internal/parser/parser.go:3:QUEST STAGE 2: THE STRUCTURAL WEAVER (Parser)
internal/token/token.go:3:QUEST STAGE 1: THE LEXICAL CONDUIT (Token Vocabulary)
```

Then search only for the current QUEST's identifiers. For example:

```console
$ grep -rn "TASK \[LEX-" internal
internal/lexer/lexer.go:10:  - TASK [LEX-01]: Implement character classification helpers:
internal/lexer/lexer.go:13:  - TASK [LEX-02]: Implement keyword lookup in identifier() for reserved words
internal/lexer/lexer.go:204:    // TASK [LEX-02]: Implement keyword matching for reserved words ...
internal/lexer/lexer.go:245:// TASK [LEX-01]: Implement character classification helpers:
internal/token/token.go:9:  - TASK [LEX-03]: Define Var ("var") and Func ("func") token kinds ...
internal/token/token.go:67:    // TASK [LEX-03]: Define Var ("var") and Func ("func") token kinds ...
internal/token/token.go:99:// TASK [LEX-03]: Add "var": Var and "func": Func to the Keywords map ...
```

The HTML guide labels every snippet with its path and starter-checkpoint line
number. Earlier stages modify different files, so the labelled location remains
stable when you reach its QUEST. Search by TASK identifier if local edits have
shifted a line.

The four 15-minute stages are recorded as annotated Git tags:

1. **Stage 1 — Lexer (`1-lexer`):** identifiers and the `var`/`func` keyword vocabulary.
2. **Stage 2 — Parser (`2-parser`):** branches, variables, identifiers, declarations, and calls.
3. **Stage 3 — Interpreter (`3-interpreter`):** environments, branches, function calls, and returns.
4. **Stage 4 — Transpiler (`4-compiler`):** direct Go generation for the same features.

## Reset or skip a QUEST

```console
$ ./savepoint.sh 0  # starter scaffold
$ ./savepoint.sh 1  # lexer complete
$ ./savepoint.sh 2  # parser complete
$ ./savepoint.sh 3  # interpreter complete
$ ./savepoint.sh 4  # transpiler complete
```

Each command reports the selected tag and confirms restoration. For example:

```text
Resetting repository to checkpoint tag: 2-parser...
Successfully restored repository checkpoint 2-parser!
```

`savepoint.sh` runs `git reset --hard` and `git clean -fd`. It intentionally
discards tracked and untracked workshop progress. Commit work you want to keep
before using it.

## Verify your checkpoint

```console
$ go test ./...
```

Successful packages finish with `ok`; packages without tests report
`[no test files]`. Return to the guide after the test run and continue with the
next QUEST.
