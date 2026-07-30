# Presenter runbook

## Before the room opens — 5 minutes

From a clean checkout, run:

```sh
go test ./...
go build -o /tmp/zing-interpreter ./cmd/zing-interpreter
go build -o /tmp/zing-compiler ./cmd/zing-compiler
/tmp/zing-interpreter check examples/source-analyzer.zing
/tmp/zing-interpreter run examples/source-analyzer.zing -- examples/fixtures/analyzer-input.zing > /tmp/analyzer-interpreted.stdout
cmp /tmp/analyzer-interpreted.stdout examples/expected/source-analyzer.stdout
```

Keep `examples/expected/source-analyzer.stdout` and
`examples/generated/source-analyzer.go` open as projector/toolchain fallbacks.

## Reference implementation demonstration — 12 minutes

1. **Pipeline — 2 minutes.** Show the README pipeline and briefly open
   `internal/lexer`, `internal/parser`, `internal/ast`, and `internal/checker`.
   Explain why the teaching implementation is handwritten, then identify
   `mini.g4` as an ANTLR-compatible alternative specification.
2. **Interpret — 2 minutes.** Run the source analyzer command above. Point out
   byte-oriented lexing, token structs, slice append, struct mutation, and map
   counting.
3. **Transpile — 3 minutes.** Run:

   ```sh
   /tmp/zing-compiler transpile -o /tmp/source-analyzer.go examples/source-analyzer.zing
   sed -n '1,120p' /tmp/source-analyzer.go
   ```

   Identify generated types, `main`, bounds checks, and runtime helpers.
4. **Build and compare — 2 minutes.** Run:

   ```sh
   /tmp/zing-compiler build -o /tmp/source-analyzer examples/source-analyzer.zing
   /tmp/source-analyzer examples/fixtures/analyzer-input.zing > /tmp/analyzer-compiled.stdout
   cmp /tmp/analyzer-interpreted.stdout /tmp/analyzer-compiled.stdout
   ```

5. **Diagnostic — 2 minutes.** Work on a temporary copy:

   ```sh
   sed 's/return n \* factorial(n - 1);/return "not an int";/' examples/04-functions.zing > /tmp/type-error.zing
   /tmp/zing-compiler check /tmp/type-error.zing
   ```

   Confirm a checker diagnostic and exit status 1 before Go generation.
6. **Bootstrap distinction — 1 minute.** Explain that Go is the seed. Zing can
   express compiler-shaped work, but a future milestone must replace and
   differentially verify each seed stage before claiming self-hosting.

## Fallbacks

- If `go build` is unavailable, show
  `examples/generated/source-analyzer.go` and the checked-in expected stdout.
- If live editing is risky, show the diagnostic command without altering a
  checked-in example.
- If time is short, skip helper details but retain the interpreter/compiled
  comparison and bootstrap distinction.

The reference segment is capped at 12 minutes, leaving the rest of the hour
for lexer/parser explanation and the attendee exercise.
