# Compilers Workshop Scaffold: Building Zing

Welcome to the **Building a Compiler in Go** workshop!

In this 1-hour workshop, you will learn how a programming language compiler, interpreter, and source-to-source generator (transpiler) fit together by completing the implementation of **Variables** and **Functions** in Zing.

---

## 1. Project & Pipeline Overview

Zing is a simplified, dynamic programming language designed for learning compiler front-ends and back-ends. The architecture follows a 4-stage pipeline:

```text
Source Code (.zing)
       │
       ▼
┌──────────────┐
│    Lexer     │  Converts raw source bytes into a sequence of Tokens
└──────┬───────┘
       │
       ▼
┌──────────────┐
│    Parser    │  Parses Tokens into an Abstract Syntax Tree (AST) using Recursive Descent
└──────┬───────┘
       │
       ├──────────────────────────┐
       ▼                          ▼
┌──────────────┐          ┌──────────────┐
│ Interpreter  │          │   Compiler   │
│  (Evaluator) │          │ (Transpiler) │  Generates clean Go source code directly
└──────────────┘          └──────────────┘
```

### Directory Structure

```text
.
├── cmd/
│   └── zing/             # Command line interface (run, transpile, repl)
├── internal/
│   ├── ast/              # Abstract Syntax Tree node definitions
│   ├── compiler/         # Go code transpiler
│   ├── diagnostic/       # Compiler error and diagnostic formatting
│   ├── interpreter/      # Dynamic AST evaluator and lexical environment
│   ├── lexer/            # Token scanner
│   ├── parser/           # Pure Recursive Descent parser
│   ├── source/           # Source positions and spans
│   └── token/            # Token kinds and keyword lookup maps
├── savepoint.sh          # Helper script to jump to checkpoint tags
└── README.md
```

---

## 2. Your Goal

The base repository (`0-base`) provides a working execution engine for arithmetic, booleans, comparisons, control flow (`if`/`else`, `for`), and built-in `print(...)`.

Your mission is to hunt down `// TODO` comments across the pipeline and implement **Variables** and **Functions** to make the language fully Turing complete:

1. **Stage 1 (Lexer):** Recognize `var` and `func` keywords.
2. **Stage 2 (Parser):** Parse variable declarations (`var x = expr`), assignment (`x = expr`), variable lookup, function declarations (`func f(a, b) { ... }`), call expressions (`f(x)`), and return statements (`return expr`).
3. **Stage 3 (Interpreter):** Implement environment variable binding and lookup, function frame creation, argument passing, and return value handling.
4. **Stage 4 (Compiler):** Transpile variable declarations, assignments, function signatures, function calls, and return statements to 1-to-1 Go code.

---

## 3. Workshop Instructions & Checkpoints

You can check your progress or reset your worktree to any checkpoint at any time using `./savepoint.sh`:

```bash
# Reset worktree to initial scaffold (Stage 0)
./savepoint.sh 0

# Jump to Lexer checkpoint (Stage 1)
./savepoint.sh 1

# Jump to Parser checkpoint (Stage 2)
./savepoint.sh 2

# Jump to Interpreter checkpoint (Stage 3)
./savepoint.sh 3

# Jump to complete reference implementation (Stage 4)
./savepoint.sh 4
```

### Running and Testing Zing

```bash
# Run tests
go test ./...

# Build the CLI
go build -o zing ./cmd/zing

# Execute a Zing program with the interpreter
./zing run example.zing

# Transpile a Zing program to Go code
./zing transpile example.zing

# Compile and run a Zing program via Go
./zing compile example.zing

# Start the Zing REPL
./zing repl
```
