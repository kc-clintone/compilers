# Phase 3: Tree-Walking Interpreter and Bootstrap Runtime

## Context

The interpreter executes a checked AST directly and makes Zing's runtime semantics visible to workshop attendees. It must agree with the generated Go implementation closely enough that the same valid program has the same observable behavior through both paths.

Runtime state must use explicit typed representations. A general `map[string]interface{}` is not the language model: environments may map resolved symbols to values, but each runtime value has a known Zing type and composite values have dedicated representations.

## Intended Outcome

At the end of this phase, every checker-valid v1 program can run through the interpreter with deterministic output, file behavior, argument handling, and source-located runtime failures.

## Dependencies

- Phase 2 supplies a checked AST and immutable `checker.Info`.
- The interpreter must not run parsing or type-checking internally.

## In Scope

- Typed runtime values, lexical environments, function calls, and recursion.
- Expression evaluation, statements, control flow, composites, and mutation.
- Built-in functions required by future bootstrap work.
- Injectable process boundaries and interpreter unit tests.

## Deferred

- REPL support, debugger hooks, profiling, bytecode, or optimization.
- User-defined errors or exception handling.
- Concurrent execution.
- Running unchecked or partially parsed programs.

## Runtime Representations

- `int`, `char`, `string`, and `bool` use explicit value variants rather than unconstrained Go values.
- A slice value contains its element type and a shared backing sequence. `append` returns a new slice header, matching Go behavior.
- A map value contains key/value types and a shared key-value store. Missing keys produce the declared value type's zero value.
- A struct value contains its resolved struct type and shared field storage, giving struct instances reference semantics.
- Environments map resolved variable/parameter symbols to typed values and link to a parent environment for lexical scope.
- Functions are resolved declarations, not first-class runtime values.
- Primitive uninitialized variables receive zero values; slices and maps receive nil-equivalent empty headers. Named structs are always initialized because the checker requires it.

## Execution Semantics

- Evaluate global initializers once in source order, register functions, then execute top-level statements.
- Create a fresh function environment per call, bind evaluated arguments to parameters, and create nested block environments.
- Evaluate call arguments and binary operands left to right.
- Implement `&&` and `||` without evaluating the right operand when the result is already known.
- Use internal control signals, distinct from runtime errors, to unwind `return`, `break`, and `continue`.
- Switch evaluates its subject once, chooses the first equal case, never falls through, and otherwise executes `default`.
- Three-clause loops execute initializer once, condition before each iteration, and assignment after each successful/continued iteration.
- Assignment evaluates its target container/index once before storing the right-hand value.
- Integer division truncates toward zero, matching Go.

## Built-in Behavior

- Implement the exact signatures fixed in Phase 0 and dispatch them using checker-resolved built-in IDs.
- `print` uses the language's stable display formatting rather than Go's default formatting for internal values.
- `args` reads from `interpreter.Options.Args`.
- `readFile` and `writeFile` use an injected filesystem interface; the default implementation uses the OS filesystem.
- `print` writes through an injected `io.Writer`, never directly to process-global stdout in the package.
- `fail`, file failures, division by zero, conversion failures, and bounds failures return a runtime error carrying the active AST span.
- Context cancellation is checked at function calls and loop iterations so runaway demo programs can be stopped.

## Interpreter API

`interpreter.Options` contains program arguments, stdout, and a filesystem implementation. Tests use in-memory implementations. `Run` returns `nil` on success or a structured runtime error; only the CLI decides how to print the error or choose an exit code.

## Ordered Implementation Tasks

1. Implement typed primitive/composite values, zero values, equality, and display formatting.
2. Implement environments and global/local variable lifecycle.
3. Implement literal, identifier, unary, binary, field, index, slice, and composite evaluation.
4. Implement declarations, blocks, assignment, branching, switch, and both loop forms.
5. Implement functions, recursion, and structured control signals.
6. Implement all built-ins with injected arguments, output, filesystem, and context handling.
7. Add structured runtime diagnostics and negative tests.
8. Run all checker-valid language fixtures through the interpreter.

## Tests and Completion Criteria

- Value tests cover zero values, reference aliasing, slice append behavior, map misses, struct mutation, equality, and formatting.
- Scope tests cover globals, nested blocks, shadowing, parameters, recursion, and repeated calls.
- Expression tests cover precedence results, short-circuit side effects, integer behavior, conversions, indexing, and slicing.
- Control-flow tests cover all branches, switch/default, break, continue, return, and loop update ordering.
- Built-in tests use injected arguments, buffers, and an in-memory filesystem.
- Runtime-error tests assert source spans for division by zero, invalid conversion, bounds, file failure, and explicit `fail`.
- Cancellation tests stop an infinite loop promptly.
- The interpreter never uses generated Go or invokes the Go toolchain.

## Phase Gate

Advance only when every checker-valid fixture either completes with its expected observable behavior or produces its expected source-located runtime error.

## Commit Strategy

1. Add typed runtime values, environments, and primitive expression evaluation.
2. Add composites, assignment, and control flow.
3. Add functions, recursion, and control unwinding.
4. Add built-ins, injected process boundaries, runtime diagnostics, and full fixture coverage.

