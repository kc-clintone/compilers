# Rehearsal record

## 30 July 2026

The Phase 5 presenter path was rehearsed on Linux x86-64 with Go 1.26.4 in a
network-isolated workspace. The module language version remains Go 1.22.

Results:

- `go vet ./...`, `go test ./...`, and `go test -race ./...` passed.
- Two-second lexer and parser fuzz runs passed after roughly 11,000 and 3,300
  executions respectively.
- Both command binaries built successfully.
- Interpreter and compiled source-analyzer stdout matched the checked-in
  expected output byte for byte.
- Fresh transpilation matched `examples/generated/source-analyzer.go` byte for
  byte.
- The deliberate type error produced a checker diagnostic and exit status 1.
- The automated build/check/run/transpile/build/compare path completed in 0.86
  seconds with warm Go caches. The presenter narration remains allocated the
  documented 12-minute segment.

The checked-in Go fallback and stdout remain available if the workshop machine
lacks a functioning Go toolchain or live shell.
