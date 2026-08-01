# Nuru

Nuru is a small, statically typed teaching language for learning how a compiler
front end, interpreter, and source-to-source compiler fit together. Its
standard-library-only Go seed implementation uses a handwritten lexer and
parser, performs a separate semantic check, and then either interprets the
checked AST or transpiles it to readable Go.

The implementation is bootstrap-ready, not self-hosted. The
[`source-analyzer.nuru`](examples/source-analyzer.nuru) example demonstrates
the data structures and operations needed by a future Nuru-written front end.

## Prerequisites

- Go 1.22 or newer.
- No network access, ANTLR installation, or third-party dependency is needed.

## Build and use the tools

From the repository root:

```sh
go test ./...
go build -o ./nuru-interpreter ./cmd/nuru-interpreter
go build -o ./nuru-compiler ./cmd/nuru-compiler

./nuru-interpreter check examples/04-functions.nuru
./nuru-interpreter examples/04-functions.nuru
./nuru-compiler check examples/04-functions.nuru
./nuru-compiler transpile examples/04-functions.nuru # writes ./04-functions.nuru.go
./nuru-compiler examples/04-functions.nuru           # writes ./nuru.out
/tmp/functions
```

Interpreter program arguments must follow `--`:

```sh
./nuru-interpreter examples/source-analyzer.nuru -- examples/fixtures/analyzer-input.nuru
```

A compiled program receives its arguments normally:

```sh
./nuru-compiler -o /tmp/source-analyzer examples/source-analyzer.nuru
/tmp/source-analyzer examples/fixtures/analyzer-input.nuru
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
back ends. `nuru.g4` remains a readable parser-generator reference; production
builds use the handwritten lexer and a pure **Recursive Descent** parser for expressions to maximize readability for learners.

Tests include lexer and parser fuzz targets, semantic and runtime unit tests,
generated-Go tests, both CLI contracts, and interpreter-versus-compiled
differential tests for successful and failing programs.

## Teaching and learning material

- [Nuru v1 language reference](docs/language.md)
- [Language evolution roadmap](docs/ROADMAP.md)
- [Guided teaching walkthrough](docs/teaching-guide.md)
- [Latest verification record](docs/verification.md)

Progressive examples and deterministic output are under `examples/`. A
checked-in generated source-analyzer fallback is available at
[`examples/generated/source-analyzer.go`](examples/generated/source-analyzer.go).

The repository is intentionally complete so it can be read stage by stage.
When using it as an exercise, an instructor or self-directed learner can stop
after lexing and parsing, omit static checking, or compare a smaller
implementation against this reference. Rewriting the seed implementation in
Nuru remains a later milestone.
