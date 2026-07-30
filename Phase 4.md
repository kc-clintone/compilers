# Phase 4: Go Transpiler, Binary Builder, and Unified CLI

## Context

Zing's compiler is a source-to-source compiler: it converts a checked Zing AST into readable Go and can invoke the Go toolchain to create a standalone executable. Keeping the intermediate Go visible makes the compilation process approachable for Go beginners and provides a practical bootstrap target.

## Intended Outcome

At the end of this phase, one CLI can check, interpret, transpile, or build a Zing program. Generated Go is formatted, uses only the standard library, and matches interpreter behavior.

## Dependencies

- Phase 2 supplies the checked AST and `checker.Info` used for code generation.
- Phase 3 defines the observable runtime semantics that generated programs must match.
- The compiler does not depend on interpreter implementation types.

## In Scope

- Go generation for the full v1 AST.
- Generated runtime helpers for Zing-specific built-ins and semantics.
- Temporary-file management and safe `go build` invocation.
- A unified `zing` command with stable diagnostics and exit codes.

## Deferred

- Optimization, minification, source maps, multi-package output, cross-compilation flags, or alternate back ends.
- Translating Go compiler diagnostics back to Zing spans; static Zing errors should already be caught by the checker.
- Installing binaries globally or publishing modules.

## Go Mapping Decisions

- Emit one formatted `package main` source file.
- Map `int`, `char`, `string`, and `bool` to Go `int`, `byte`, `string`, and `bool`.
- Map Zing slices and maps directly to Go slices and maps with recursively mapped element types.
- Emit each Zing struct as a Go struct and map named struct types to pointers. Constructors emit address-taking composite literals.
- Emit globals and functions at package scope. Wrap Zing top-level statements in generated `func main()`.
- Preserve expression precedence with an emission precedence table and add parentheses only when required.
- Lower special operations when direct Go syntax would differ, including conversions, string formatting, failure behavior, and checked file I/O.
- Use deterministic declaration, field, import, and helper ordering so generated-source tests are stable.

## Generated Runtime

- Inject a helper only when the checked program uses the corresponding built-in or semantic operation.
- Use `fmt.Println` behavior only through a Zing display helper where direct formatting would differ.
- Use `os.Args[1:]` for compiled-program `args()`.
- Implement file helpers with `os.ReadFile` and `os.WriteFile`; failures print a stable diagnostic to stderr and terminate non-zero.
- Implement integer/string conversions with `strconv` where required.
- Implement `fail` through the same generated failure helper as file errors.
- Import only referenced standard-library packages.

## Builder Behavior

- `compiler.Generate` returns `go/format`-formatted source or an internal compiler error.
- `compiler.Build` creates a dedicated temporary directory, writes the generated source, and invokes `go build -o <requested-path> <source-file>` using `exec.CommandContext` without a shell.
- Capture Go stdout/stderr and include them in a concise build error.
- Clean the temporary directory on success or failure. The explicit `transpile` command is how users retain generated source.
- Never overwrite the input `.zing` file. Validate that the output path is different from the input before building or transpiling.

## CLI Contract

```text
zing check <file>
zing run <file> [-- program-args...]
zing transpile -o <output.go> <file>
zing build -o <binary> <file>
```

- `check` scans, parses, and checks without execution.
- `run` executes through the interpreter and passes only arguments after `--` to `args()`.
- `transpile` writes formatted Go to the required output path.
- `build` generates temporary Go and creates the requested binary.
- Source/runtime/build failures exit 1; usage and unknown-command errors exit 2.
- CLI code owns filesystem reads, diagnostic rendering, process exit decisions, and signal-derived context cancellation.
- Internal packages return errors and must never call `os.Exit`.

## Ordered Implementation Tasks

1. Implement a structured Go writer with indentation and precedence-aware expression emission.
2. Generate types, structs, globals, functions, and top-level `main` statements.
3. Generate all statements, expressions, composites, mutation, and control flow.
4. Add usage tracking and deterministic injection of runtime helpers/imports.
5. Format and syntactically parse generated Go in tests.
6. Implement the temporary builder and toolchain error handling.
7. Implement the four CLI subcommands and their shared front-end pipeline.
8. Add compiled-binary tests and compare them with interpreter results.

## Tests and Completion Criteria

- Golden tests cover readable Go for each declaration, statement, expression, and built-in.
- Every generated fixture passes `go/parser` and `go/format` checks.
- Representative generated programs compile with `go build` and use no non-standard imports.
- Helper/import tests prove unused runtime code is not emitted and ordering is deterministic.
- CLI tests cover commands, missing/extra arguments, `--` handling, output paths, diagnostics, and exit codes.
- Build cancellation and missing Go toolchain errors are reported without panics or leaked temporary directories.
- Interpreter and compiled executions agree for all fixtures selected for differential testing.

## Phase Gate

Advance only when every checker-valid v1 feature can be transpiled, generated sources build, and the CLI exposes all four modes with documented exit behavior.

## Commit Strategy

1. Add core Go generation for declarations, expressions, and statements.
2. Add composite/reference semantics and generated runtime helpers.
3. Add formatting, golden tests, and the safe binary builder.
4. Add the unified CLI and end-to-end command tests.

