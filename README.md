# Compilers Workshop Scaffold: Building Zing

Welcome to the **Building a Compiler in Go** workshop!

In this workshop, you will learn compiler and interpreter engineering by embarking on a step-by-step **QUEST** to implement **Branches (`if`/`else`)**, **Variables**, and **Functions** in the Zing programming language.

---

## 1. Project & Pipeline Overview

Zing is a dynamically typed programming language designed for learning compiler architecture. The pipeline follows 4 stages:

```text
Source Code (.zing)
       │
       ▼
┌──────────────┐
│    Lexer     │  Tokenizes raw source text into a stream of Tokens
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
│  (Evaluator) │          │ (Transpiler) │  Transpiles AST directly into clean Go code
└──────────────┘          └──────────────┘
```

---

## 2. CLI Tools: `zing-interpreter` and `zing-compiler`

The repository provides two separate CLI entry points:

### 1. `zing-interpreter` (Direct Evaluator & REPL)
- **Default Execution:** Interprets a Zing source file directly.
  ```bash
  ./zing-interpreter examples/03-interpreter.zing
  ```
- **Interactive REPL:** Opens REPL when run with no arguments.
  ```bash
  ./zing-interpreter
  ```
- **Inspect Tokens:** `./zing-interpreter tokens examples/01-tokens.zing`
- **Inspect AST:** `./zing-interpreter ast examples/02-ast.zing`

### 2. `zing-compiler` (Native Go Transpiler & Builder)
- **Default Execution:** Transpiles a Zing program and builds a native executable.
  ```bash
  ./zing-compiler examples/04-compiled.zing -o ./04-compiled
  ./04-compiled
  ```
- **Inspect Tokens:** `./zing-compiler tokens examples/01-tokens.zing`
- **Inspect AST:** `./zing-compiler ast examples/02-ast.zing`
- **Transpile Go Source:** `./zing-compiler transpile examples/04-compiled.zing`

*Note: Any unrecognized options or arguments passed to either CLI will trigger a warning on `os.Stderr` and be ignored.*

---

## 3. The QUEST & TASK Adventure Architecture

Each stage of the workshop is structured as a **QUEST**.

### QUEST Banners
Every file that requires modification contains a **QUEST Banner** at the top summarizing the stage goals, tasks involved, and relevant CLI commands.

### TASK Markers & Symbol Discovery
Instead of scattered comments, task markers (`TASK [LEX-01]`, `TASK [PARSE-01]`, `TASK [EVAL-01]`, `TASK [GEN-01]`) are placed *only* at the exact location where code must be written. Required structs, constants, or token kinds are mentioned in the task description so you can search for symbol definitions in the codebase.

Search for tasks using `grep`:
```bash
# Search for Stage 1 (Lexer) tasks
grep -rn "TASK \[LEX-" .

# Search for Stage 2 (Parser) tasks
grep -rn "TASK \[PARSE-" .

# Search for Stage 3 (Interpreter) tasks
grep -rn "TASK \[EVAL-" .

# Search for Stage 4 (Compiler) tasks
grep -rn "TASK \[GEN-" .
```

### Solution HINTs
At the bottom of every task file, you will find a `QUEST HINTS & SOLUTIONS` section with drop-in code snippets for each task (e.g. `HINT [LEX-01-HINT]`).

---

## 4. Resetting or Skipping Quests (`savepoint.sh`)

If you get stuck or want to reset/skip a stage, use `./savepoint.sh`:

```bash
# Reset to initial scaffold (Stage 0)
./savepoint.sh 0

# Jump to Lexer checkpoint (Stage 1 solved)
./savepoint.sh 1

# Jump to Parser checkpoint (Stage 2 solved)
./savepoint.sh 2

# Jump to Interpreter checkpoint (Stage 3 solved)
./savepoint.sh 3

# Jump to complete solution (Stage 4 solved)
./savepoint.sh 4
```

---

## 5. Running Unit Tests

```bash
# Run all tests across the repository
go test ./...
```
