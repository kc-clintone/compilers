# Phase 5: Differential Testing, Examples, and Workshop Rehearsal

## Context

The reference implementation is complete only when the interpreter and compiled binaries demonstrably agree and the workshop path works from a clean environment. This phase turns the implementation into a reliable teaching artifact and proves bootstrap readiness with a compiler-shaped Zing program rather than only small algorithms.

## Intended Outcome

At the end of this phase, the repository contains progressive examples, comprehensive end-to-end tests, a Zing source analyzer exercising bootstrap-critical features, and a rehearsed August 1 demonstration.

## Dependencies

- Phases 0–4 are complete and their gates pass.
- The language is frozen; fixes in this phase may correct defects but must not casually add syntax.

## In Scope

- Cross-backend differential testing.
- Beginner-oriented examples and expected output.
- The source-analyzer bootstrap probe.
- Setup documentation, workshop runbook, and clean-machine rehearsal.

## Deferred

- The attendee exercise repository and its intentionally incomplete implementation.
- A self-hosted Zing compiler or interpreter.
- Performance benchmarking beyond preventing obvious hangs.
- CI/CD publishing, package releases, or installers unless separately requested.

## Example Programs

Create examples in increasing complexity:

1. `01-basics.zing`: variables, types, arithmetic, conversion, strings, and `print`.
2. `02-control-flow.zing`: `if`, `switch`, while-style loops, and three-clause loops.
3. `03-collections.zing`: slices, maps, structs, constructors, indexing, slicing, and mutation.
4. `04-functions.zing`: typed functions, lexical scope, recursion, and returns.
5. `05-files.zing`: arguments, reading, writing, lengths, and diagnosed failures.
6. `source-analyzer.zing`: the bootstrap-readiness acceptance program.

Each example includes a short header comment describing the concept and a checked-in expected stdout file where output is deterministic.

## Source Analyzer Requirements

The analyzer is deliberately smaller than a compiler but follows the same data flow:

1. Read a source path from `args()` and fail with a helpful message when absent.
2. Read the file and scan a documented subset of Zing tokens using byte-oriented string indexing.
3. Represent tokens with a struct containing kind, lexeme, line, and column.
4. Accumulate tokens in a slice using `append`.
5. Count token kinds in a map.
6. Use helper functions, branching, both loop forms, slicing, conversions, and struct mutation.
7. Print a deterministic token listing and summary.

The program must run unchanged through `zing run`, `zing transpile`, and a binary produced by `zing build`. It is the v1 evidence that Zing can express the core data manipulation required by a future scanner/parser implementation; it is not presented as self-hosting itself.

## Differential Test Harness

- For each valid fixture, capture interpreter stdout/stderr, exit classification, and declared output files.
- Build the same fixture, execute it with the same arguments and isolated filesystem inputs, then compare observable results.
- Normalize only unavoidable environment data such as temporary absolute paths; do not hide semantic differences.
- Keep compile failures separate from runtime-failure fixtures.
- Include success, explicit `fail`, file failure, bounds failure, conversion failure, and recursive-call cases.
- Run binaries with timeouts and report the fixture name on failure.

## Workshop Documentation

- Add a root README with prerequisites, project purpose, pipeline overview, command examples, and links to phase documents.
- Add a language reference covering syntax and semantics without requiring readers to inspect Go code.
- Add a short presenter runbook with exact commands, expected results, fallback screenshots/output, and time boxes.
- Explain handwritten parsers first, then mention ANTLR and other parser generators as alternative tooling.
- Clearly distinguish the full reference implementation from the smaller attendee exercise.

## Demonstration Script

Reserve a short, repeatable portion of the one-hour workshop for this sequence:

1. Run `zing check` and `zing run` on the source analyzer.
2. Show scanner, AST, checker, interpreter, and transpiler boundaries in the repository.
3. Run `zing transpile`, open the readable Go, and identify the generated `main` and runtime helper.
4. Run `zing build`, execute the binary on the same input, and compare its output with the interpreter.
5. Introduce a small type error and show the Zing diagnostic before Go generation.
6. Explain how a future Zing-written compiler would replace the Go seed one stage at a time.

## Ordered Implementation Tasks

1. Build the reusable differential test harness and migrate representative fixtures into it.
2. Add all progressive examples and expected output.
3. Implement and validate `source-analyzer.zing` against a fixed sample input.
4. Fill coverage gaps found by cross-backend tests and classify any intentional differences.
5. Write the README, language reference, pipeline explanation, and presenter runbook.
6. Test all documented commands from a clean checkout with the documented minimum Go version.
7. Rehearse the timed demo and preserve fallback generated Go and expected output.

## Tests and Completion Criteria

- `go test ./...` passes from a clean checkout.
- Every example passes `check`, `run`, `transpile`, and `build` where applicable.
- Differential tests show identical observable behavior between interpreter and compiled execution.
- The source analyzer uses every bootstrap-critical feature listed above and produces deterministic output.
- Documentation commands are executable verbatim and contain no machine-specific paths.
- Temporary files and test binaries do not remain after tests.
- The complete demo succeeds without network access or ANTLR.
- A timed rehearsal fits the allocated reference-demo segment and has a documented fallback for toolchain failure.

## Phase Gate

The workshop artifact is ready only after the clean-checkout rehearsal passes, interpreter/compiled outputs match, and the presenter runbook has been followed verbatim at least once.

## Commit Strategy

1. Add the differential harness and cross-backend fixtures.
2. Add progressive examples and expected output.
3. Add the source analyzer and bootstrap-readiness acceptance test.
4. Add documentation, presenter runbook, and final clean-checkout verification notes.

