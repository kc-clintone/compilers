# Phase 1: Handwritten Lexer, Parser, and AST

## Context

The reference front end must show beginners how source text becomes a structured program without hiding that process behind generated code. It therefore uses a handwritten lexer and a recursive-descent parser, with Pratt parsing for expressions. `mini.g4` is the syntax contract and future parser-generator reference, not a runtime dependency.

## Intended Outcome

At the end of this phase, valid Zing source is converted into a parser-independent AST with accurate source spans. Invalid input produces stable, useful diagnostics and never panics.

## Dependencies

- Phase 0 is complete and the v1 grammar is frozen.
- The shared source-position and diagnostic types from Phase 0 are authoritative.

## In Scope

- Tokenization, comments, literals, escapes, keywords, and punctuation.
- AST definitions for the complete v1 language.
- Declaration/statement recursive descent and Pratt expression parsing.
- Local parser recovery and front-end unit tests.

## Deferred

- Name resolution and type checking.
- Runtime values or execution.
- Go generation.
- ANTLR generation, visitors, listeners, or runtime libraries.

## Front-End Interfaces and Representations

- A token contains its kind, original lexeme, decoded literal value when applicable, and half-open source span.
- A source span contains filename plus start/end byte offsets, lines, and columns.
- The AST is organized around `Program`, `Decl`, `Stmt`, `Expr`, and `TypeExpr` interfaces.
- Every AST node exposes a stable node ID and source span. Node IDs let the checker attach inferred/resolved information without mutating syntax nodes.
- The parser returns a partial AST only for diagnostic recovery; callers must discard it whenever diagnostics are present.

The AST must represent:

- Global variables, structs and fields, functions, parameters, and return types.
- Blocks, local variables, assignment, expression statements, branches, switch cases, loops, loop control, and return.
- Identifiers, primitive/composite literals, constructors, `make`, calls, unary/binary expressions, indexing, slicing, and field access.
- Assignment targets as expressions, with legality deferred to the checker.

## Lexer Decisions

- Recognize all keywords before emitting an identifier token.
- Decode `\n`, `\r`, `\t`, `\\`, `\"`, and `\'` in string/character literals.
- Require decoded character literals to contain exactly one byte.
- Support decimal non-negative integer tokens; unary minus remains parser syntax.
- Skip whitespace, `//` line comments, and non-nesting `/* ... */` block comments while maintaining positions.
- Diagnose invalid bytes, invalid escapes, integer overflow, and unterminated strings, characters, or comments.
- Always emit EOF, including after lexer errors, to make parser recovery deterministic.

## Parser Decisions

- Parse all declarations before top-level executable statements and diagnose declarations that appear afterward.
- Parse blocks and statements with recursive descent.
- Parse expressions using this precedence, from lowest to highest: `||`, `&&`, equality, comparison, addition/subtraction, multiplication/division/modulo, prefix `!`/`-`, then postfix call/index/slice/field access.
- Binary operators associate left; prefix operators associate right; postfix operations may chain.
- Distinguish a composite literal from an ordinary expression using its leading type syntax.
- Require commas between composite elements and constructor fields; allow an optional trailing comma.
- Parse `switch` without fallthrough and require at most one `default` clause.
- Synchronize after errors at semicolons, closing braces, and declaration/statement-leading keywords.

## Ordered Implementation Tasks

1. Implement token kinds, source spans, and diagnostic formatting tests.
2. Implement the lexer and table-driven lexer tests.
3. Define the complete AST, node IDs, and a compact debug formatter used by tests.
4. Implement type-expression and declaration parsing.
5. Implement Pratt expression parsing, including postfix chains and composite construction.
6. Implement statements, blocks, control flow, and top-level ordering.
7. Add synchronization and multi-error parser tests.
8. Add full-program fixtures covering every grammar production.

## Tests and Completion Criteria

- Lexer tests cover every token, keyword, comment form, escape, invalid character, overflow, and unterminated construct.
- Position tests include multiple lines, tabs, comments, and escaped literals.
- Parser tests lock operator precedence and associativity through AST snapshots.
- Fixtures cover nested field/index targets, empty and populated composites, both loop forms, recursive functions, and top-level statements.
- Malformed-source tests verify diagnostic text, position, ordering, and recovery.
- Fuzz tests for the lexer and parser establish that arbitrary byte input cannot panic or hang.
- `go test ./...` succeeds without ANTLR installed.

## Phase Gate

Proceed only when all valid syntax in `mini.g4` has a corresponding AST fixture and all invalid fixtures produce diagnostics rather than partial execution.

## Commit Strategy

1. Add source positions, diagnostics, tokens, and the lexer with tests.
2. Add the complete AST and its test formatter.
3. Add declaration, type, and Pratt expression parsing with tests.
4. Add statement parsing, recovery, full-program fixtures, and fuzz tests.
