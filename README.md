# Zing

Zing is a small, statically typed teaching language for a one-hour compiler
workshop on 1 August 2026. Its standard-library-only Go seed implementation
uses a handwritten lexer and parser, performs a separate semantic check, and
then either interprets the checked AST or transpiles it to readable Go.

The implementation is bootstrap-ready, not self-hosted. The
[`source-analyzer.zing`](examples/source-analyzer.zing) example demonstrates
the data structures and operations needed by a future Zing-written front end.

## Prerequisites

- Go 1.22 or newer.
- No network access, ANTLR installation, or third-party dependency is needed.

## Build and use the tools

From the repository root:

```sh
go test ./...
go build -o ./zing-interpreter ./cmd/zing-interpreter
go build -o ./zing-compiler ./cmd/zing-compiler

./zing-interpreter check examples/04-functions.zing
./zing-interpreter run examples/04-functions.zing
./zing-compiler check examples/04-functions.zing
./zing-compiler transpile -o /tmp/functions.go examples/04-functions.zing
./zing-compiler build -o /tmp/functions examples/04-functions.zing
/tmp/functions
```

Interpreter program arguments must follow `--`:

```sh
./zing-interpreter run examples/source-analyzer.zing -- examples/fixtures/analyzer-input.zing
```

A compiled program receives its arguments normally:

```sh
./zing-compiler build -o /tmp/source-analyzer examples/source-analyzer.zing
/tmp/source-analyzer examples/fixtures/analyzer-input.zing
```

Both tools return 0 for success, 1 for source/runtime/build failures, and 2 for
invalid command-line usage. Compiler output paths must differ from the input
source path.

## Pipeline

```text
source -> lexer -> parser/AST -> static checker -> interpreter
                                           `----> Go generator -> go build
```

The matching stages live under `internal/`. The checker produces immutable
node-ID-keyed type, symbol, struct, and built-in resolutions consumed by both
back ends. `mini.g4` remains a readable parser-generator reference; production
builds use the handwritten lexer and Pratt/recursive-descent parser.

Tests include lexer and parser fuzz targets, semantic and runtime unit tests,
generated-Go tests, both CLI contracts, and interpreter-versus-compiled
differential tests for successful and failing programs.

## Workshop material

- [Zing v1 language reference](docs/language.md)
- [Bootstrap-readiness matrix](docs/bootstrap.md)
- [Presenter runbook](docs/presenter-runbook.md)
- [Latest rehearsal record](docs/rehearsal.md)
- [Phase 0: specification and setup](Phase%200.md)
- [Phase 1: lexer, parser, and AST](Phase%201.md)
- [Phase 2: static checking](Phase%202.md)
- [Phase 3: interpreter](Phase%203.md)
- [Phase 4: transpiler and CLIs](Phase%204.md)
- [Phase 5: validation and rehearsal](Phase%205.md)

Progressive examples and deterministic output are under `examples/`. A
checked-in generated source-analyzer fallback is available at
[`examples/generated/source-analyzer.go`](examples/generated/source-analyzer.go).

This repository is the complete presenter reference. A smaller attendee
exercise may stop after lexing/parsing or omit the checker. Rewriting the seed
implementation in Zing remains a later milestone.
