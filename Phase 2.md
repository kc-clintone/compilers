# Phase 2: Symbol Resolution and Static Type Checking

## Context

Parsing establishes structure but cannot determine whether names exist, expressions have compatible types, or control flow is valid. Zing uses a separate static semantic pass so the interpreter and transpiler accept the same programs and do not duplicate language rules.

The attendee exercise may omit this phase for simplicity. In the reference project it is a visible, separately demonstrable compiler stage.

## Intended Outcome

At the end of this phase, every valid AST has resolved symbols and type information. Invalid programs receive source-located diagnostics before interpretation or Go generation begins.

## Dependencies

- Phase 1 provides the final AST, stable node IDs, source spans, and parser contract.
- No checker behavior may depend on interpreter or generated-Go behavior.

## In Scope

- Type representation and equality.
- Global and lexical symbol tables.
- Name resolution, type checking, assignment validation, and control-flow checks.
- Static signatures for language built-ins.
- Ordered, recoverable semantic diagnostics.

## Deferred

- Type inference beyond validating explicitly declared types and composite literal elements.
- Generics, overloads, user-defined conversions, interfaces, or union types.
- Definite assignment analysis for slices and maps.
- Optimization or constant folding.

## Checker Interface and Output

`checker.Check` performs all semantic work and returns an immutable `checker.Info` when successful. `Info` contains maps keyed by AST node ID for:

- The type of each expression.
- The symbol resolved by each identifier.
- The struct declaration resolved by each named type and constructor.
- The selected built-in operation for special calls such as `len`, `append`, and conversions.

The interpreter and compiler must consume this information rather than resolving names or re-deriving types independently.

## Symbol and Scope Rules

- Use distinct symbol kinds for types, variables, functions, parameters, fields, and built-ins.
- Register struct names, function signatures, and global variables in a first pass so forward function calls and recursion work.
- Resolve global initializers, function bodies, and top-level statements in a second pass.
- Create scopes for functions and blocks. Function parameters begin in the function body scope.
- Shadowing an outer local is allowed; duplicate declarations in the same namespace and scope are errors.
- Type names and value names use separate namespaces. Struct fields are resolved only against their owning struct.
- Top-level statements share the global value scope but cannot introduce declarations visible to earlier global initializers.

## Type Rules

- Assignment and return require identical types; there are no implicit conversions.
- Arithmetic accepts `int`, except `+` also accepts two strings.
- Ordering accepts matching integers, characters, or strings. Equality accepts matching primitive types and struct references.
- Logical operators accept booleans and short-circuit at runtime.
- Conditions in `if` and `for` must be Boolean.
- Slice indices and bounds must be integers; indexing a string returns `char`; slicing a string returns `string`.
- Map keys must match the declared key type. Map assignment and lookup use the declared value type.
- Struct constructors must name every field exactly once and may list fields in any order.
- `make` accepts only slice and map types. A slice length, when supplied, must be an integer; maps take no size argument in v1.
- `append` accepts a slice and one element of its element type and returns that same slice type.
- `len` accepts strings, slices, or maps and returns `int`.
- Conversion calls are limited to the signatures frozen in Phase 0.
- `print` accepts zero or more values of any v1 type and returns no value.
- `readFile`, `writeFile`, `args`, and `fail` use their fixed Phase 0 signatures.

## Control-Flow Rules

- `break` and `continue` are legal only inside a loop.
- `return` is legal only inside a function and must match its declared result.
- A no-result function cannot return a value; a value-returning function cannot use a bare return.
- A value-returning function must return on every reachable path. A block guarantees return after a return statement, an `if` only when both branches guarantee return, and a `switch` only when it has `default` and every clause guarantees return. Loops are not assumed to terminate.
- Each switch case expression must have the same type as the switch expression; duplicate literal cases and multiple defaults are errors.
- Unreachable statements after a guaranteed return are diagnosed as warnings only if warnings are introduced later; they do not block v1.

## Diagnostic Policy

- Continue after an error only when a sentinel invalid type prevents cascading messages.
- Prefer one primary diagnostic at the narrowest relevant span.
- Sort diagnostics by filename, start offset, then emission order.
- Never expose Go type names or generated-code details in Zing diagnostics.

## Ordered Implementation Tasks

1. Define canonical primitive, slice, map, struct-reference, no-value, and invalid types.
2. Implement symbols, nested scopes, and duplicate-name tests.
3. Implement the global declaration pass and named-type validation.
4. Implement expression typing and identifier resolution.
5. Implement statement, assignment, function, and control-flow checking.
6. Add special checking for built-ins, composites, indexing, and slicing.
7. Add diagnostic recovery, sorting, and full invalid-program fixtures.
8. Expose checker information through read-only query methods used by later phases.

## Tests and Completion Criteria

- Valid tests cover recursion, forward calls, shadowing, every operator, every composite type, all built-ins, and all control-flow forms.
- Invalid tests cover unknown names/types, duplicates, wrong arity, wrong operands, invalid targets, missing/duplicate fields, bad returns, and illegal loop control.
- Return-path tests cover nested blocks, both-sided branches, switches with/without default, and loops.
- Diagnostic tests assert source locations and suppress avoidable cascades.
- Every parser fixture is classified explicitly as checker-valid or checker-invalid.
- `checker.Info` is sufficient for later phases without AST mutation or a second resolver.

## Phase Gate

Interpretation and transpilation must reject any program that does not produce a successful `checker.Info`. Advance only when checker-valid fixtures cover the complete v1 syntax.

## Commit Strategy

1. Add types, symbols, scopes, and the declaration pass.
2. Add expression resolution and type checking.
3. Add statement/control-flow validation and return analysis.
4. Add built-in/composite rules and comprehensive diagnostic fixtures.

