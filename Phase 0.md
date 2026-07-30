# Phase 0: Freeze the Zing v1 Language and Initialize the Project

## Context

Zing is a small, statically typed teaching language whose reference implementation is written in Go. The August 1, 2026 workshop will demonstrate a complete interpreter and a source-to-source compiler that emits Go. Attendees will work on a smaller exercise implementation, while this repository provides the complete reference pipeline.

The v1 language must be capable of expressing a future Zing implementation of its own scanner, parser, checker, and transpiler. Writing those components in Zing is explicitly deferred; this phase only fixes the language and runtime capabilities they will need.

`mini.g4` remains the readable language specification and may later be used for ANTLR experiments. The production v1 parser will be handwritten and must not depend on ANTLR or generated parser code.

## Intended Outcome

At the end of this phase, the repository builds as a Go module, the Zing v1 syntax and semantics are documented without unresolved design choices, and later phases have stable package contracts to implement.

## Dependencies

None. This phase is the prerequisite for every implementation phase.

## In Scope

- Initializing the Go module and reference implementation layout.
- Freezing v1 syntax, types, runtime semantics, diagnostics, and CLI behavior.
- Updating `mini.g4` as a specification in the implementation commit for this phase.
- Documenting the exact feature set that makes a future self-hosted compiler practical.

## Deferred

- Generating or using an ANTLR parser.
- Rewriting the interpreter, checker, or compiler in Zing.
- Floats, pointers, interfaces, methods, closures, modules, user imports, concurrency, and garbage-collection controls.
- Optimizations or native machine-code generation.

## Fixed Language Decisions

### Program structure

- Source files use the `.zing` extension.
- Type, function, and global variable declarations precede executable top-level statements.
- Top-level statements form the program entry point. The Go transpiler wraps them in `func main()`.
- `main` is a reserved identifier and cannot be declared by user code.
- Functions may recurse, accept typed parameters, and return either one typed value or no value.
- Nested function declarations and first-class function values are not supported in v1.

### Types and values

- Primitive types are `int`, `char`, `string`, and `bool`.
- Composite types are slices (`[]T`), maps (`map[K]V`), and named structs.
- `char` is one byte. String indexing and slicing are byte-oriented so the interpreter and generated Go agree exactly.
- Primitives have value semantics. Slices, maps, and struct instances have reference semantics.
- Map keys are restricted to `int`, `char`, `string`, and `bool`.
- Uninitialized primitive values use Go-compatible zero values. Named struct variables require an initializer.
- A missing map key returns the zero value of the map's value type.

### Required syntax

- Variable declarations: `var name Type` and `var name Type = expression`.
- Struct declarations and named-field construction: `Token{kind: 1, text: "var"}`.
- Slice and map literals: `[]int{1, 2}` and `map[string]int{"x": 1}`.
- Empty/dynamic allocation: `make([]Token, 0)` and `make(map[string]int)`.
- Field access, indexing, and two-bound slicing: `value.field`, `value[index]`, and `value[low:high]`.
- Assignments may target variables, fields, nested fields, slice elements, or map entries.
- Control flow includes `if/else`, non-fallthrough `switch`, while-style `for`, three-clause `for`, `break`, `continue`, and `return`.
- Operators retain the precedence shown in `mini.g4`; `&&` and `||` short-circuit.
- There is no implicit numeric or string conversion.

### Built-in runtime surface

- `print(values...)` writes values separated by spaces followed by a newline.
- `args() []string` returns program arguments, excluding the Zing source path or executable name.
- `readFile(path string) string` reads a complete file or terminates with a runtime diagnostic.
- `writeFile(path string, contents string)` writes a complete file or terminates with a runtime diagnostic.
- `len(value) int` accepts strings, slices, and maps.
- `append(slice, value)` returns the resulting slice and does not mutate the caller's slice header implicitly.
- Explicit conversions are `int(char|string)`, `char(int)`, and `string(int|char|bool)`.
- `fail(message string)` terminates execution with a diagnosed failure and non-zero exit status.

### Diagnostics and command line

- Diagnostics use `file:line:column: phase: message` and are ordered by source position.
- Scanner and parser errors may recover to report additional independent errors. No phase runs after its prerequisite phase reports errors.
- The final CLI commands are `check`, `run`, `transpile`, and `build`.
- Exit code 0 means success, 1 means a source/runtime/build failure, and 2 means invalid CLI usage.

## Project Layout and Package Contracts

- `cmd/zing` owns argument parsing and user-facing output.
- `internal` packages separate diagnostics/source positions, tokens/scanning, AST/parsing, checking, interpretation, and Go generation/building.
- Examples and test fixtures live outside implementation packages and are usable directly by the workshop commands.
- The module is `github.com/kc-clintone/compilers` and uses only the Go standard library.
- All phases target Go 1.22 or newer.

The major pipeline contracts are fixed as follows:

```go
parser.Parse(filename string, source []byte) (*ast.Program, []diagnostic.Diagnostic)
checker.Check(program *ast.Program) (*checker.Info, []diagnostic.Diagnostic)
interpreter.Run(ctx context.Context, program *ast.Program, info *checker.Info, opts interpreter.Options) error
compiler.Generate(program *ast.Program, info *checker.Info) ([]byte, error)
compiler.Build(ctx context.Context, goSource []byte, outputPath string) error
```

## Ordered Implementation Tasks

1. Initialize `go.mod`, the command package, internal package skeletons, and a smoke test.
2. Record the fixed semantics above in a concise language reference.
3. Revise `mini.g4` to cover the v1 syntax while keeping it valid ANTLR grammar.
4. Add a bootstrap-readiness matrix mapping scanner, parser, checker, and transpiler needs to Zing features.
5. Add placeholder package documentation describing each pipeline boundary without implementing later phases.

## Tests and Completion Criteria

- `go test ./...` succeeds from a clean checkout.
- The grammar and language reference agree on every v1 construct.
- Every item in the bootstrap-readiness matrix maps to a fixed language or built-in feature.
- There are no ANTLR dependencies, generated parser files, or external Go dependencies.
- Package contracts compile and contain no competing representations of source positions or diagnostics.

## Phase Gate

Do not begin scanner or AST work until syntax, byte-oriented string behavior, reference semantics, built-in signatures, diagnostics, and entry-point behavior are documented and reflected in `mini.g4`.

## Commit Strategy

1. Initialize the Go module and package layout with a passing smoke test.
2. Add the Zing v1 language reference and bootstrap-readiness matrix.
3. Update `mini.g4` to the frozen v1 syntax.

