# Guided teaching walkthrough

This guide presents the repository as a sequence of observable compiler stages.
It can support an instructor-led lesson, a study group, or self-directed
learning. The commands are deliberately repeatable and do not modify tracked
files.

## Prepare the repository

From a clean checkout, run:

```sh
go test ./...
go build -o /tmp/zing-interpreter ./cmd/zing-interpreter
go build -o /tmp/zing-compiler ./cmd/zing-compiler
/tmp/zing-interpreter check examples/source-analyzer.zing
/tmp/zing-interpreter examples/source-analyzer.zing -- examples/fixtures/analyzer-input.zing > /tmp/analyzer-interpreted.stdout
cmp /tmp/analyzer-interpreted.stdout examples/expected/source-analyzer.stdout
```

The final `cmp` exits successfully when interpreter output matches the
checked-in expectation. `examples/expected/source-analyzer.stdout` and
`examples/generated/source-analyzer.go` also make it possible to inspect the
expected results without running the Go toolchain.

## Follow the pipeline

1. **Orient around the pipeline.** Start with the diagram in the README, then
   inspect `internal/lexer`, `internal/parser`, `internal/ast`, and
   `internal/checker`.
   The handwritten implementation exposes the mechanics clearly, while
   `mini.g4` shows how the same syntax can be described for a parser generator.
2. **Interpret a program.** Run the source analyzer command above. Trace its
   byte-oriented lexing, token structs, slice append, struct mutation, and map
   counting.
3. **Inspect generated Go.** Run:

   ```sh
   /tmp/zing-compiler transpile -o /tmp/source-analyzer.go examples/source-analyzer.zing
   sed -n '1,120p' /tmp/source-analyzer.go
   ```

   Find the generated types, `main`, bounds checks, and runtime helpers. Compare
   them with the corresponding interpreter behavior.
4. **Build and compare both back ends.** Run:

   ```sh
   /tmp/zing-compiler -o /tmp/source-analyzer examples/source-analyzer.zing
   /tmp/source-analyzer examples/fixtures/analyzer-input.zing > /tmp/analyzer-compiled.stdout
   cmp /tmp/analyzer-interpreted.stdout /tmp/analyzer-compiled.stdout
   ```

   A successful `cmp` demonstrates that both execution paths produced the same
   observable output.
5. **Explore a static diagnostic.** Create a temporary erroneous program:

   ```sh
   sed 's/return n \* factorial(n - 1);/return "not an int";/' examples/04-functions.zing > /tmp/type-error.zing
   /tmp/zing-compiler check /tmp/type-error.zing
   ```

   The compiler reports a Zing checker diagnostic and exits with status 1
   before generating Go. Use the location and phase label to trace the error
   from syntax into semantic checking.
6. **Discuss bootstrap readiness.** Go remains the seed implementation. Zing
   can express compiler-shaped work, but self-hosting would require replacing
   and differentially verifying each seed stage.

## Suggested teaching variations

- Focus on the lexer and Pratt parser for an introductory front-end lesson.
- Compare AST identity with the node-ID-keyed checker information for a lesson
  on keeping syntax and semantics separate.
- Add a runtime-error fixture and compare interpreter and compiled diagnostics.
- Modify a progressive example, predict its output, then verify it through both
  back ends.

If Go is unavailable, use the checked-in generated analyzer and expected stdout
as static examples. All commands write generated files and binaries under
`/tmp`, so the repository remains clean throughout the walkthrough.
