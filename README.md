# compiler-workshop

A hands-on workshop for building a tiny programming language in Go from scratch.

This repository is intentionally scaffolded so participants can implement the core compiler pipeline step by step during the workshop. The goal is to learn the structure of a small language implementation without needing to start from a blank slate.

## Workshop overview

Participants will build a tiny language with the following pieces:

- a lexer
- a parser
- an abstract syntax tree (AST)
- an interpreter
- a compiler that emits simple pseudo-assembly

The project is designed for beginners and emphasizes clarity over complexity. Each stage introduces a small, focused concept that builds on the previous one.

## Learning objectives

By the end of the workshop, participants should be able to:

- explain the role of each stage in a compiler pipeline
- implement a simple tokenizer and parser
- represent programs using an AST
- evaluate programs with an interpreter
- emit simple instructions from a compiler
- understand how a small language is structured in Go

## Repository layout

```text
compiler-workshop/
├── README.md
├── LICENSE
├── .gitignore
├── go.mod
├── Makefile
├── verify.go
├── cmd/
│   └── tiny/
│       └── main.go
├── internal/
│   ├── token/
│   ├── lexer/
│   ├── parser/
│   ├── ast/
│   ├── interpreter/
│   ├── compiler/
│   ├── object/
│   └── repl/
├── examples/
├── exercises/
├── tests/
├── docs/
└── solutions/
```

## Prerequisites

Before starting, make sure you have:

- Go 1.22 or newer installed
- a terminal or command prompt
- a text editor or IDE such as VS Code
- basic familiarity with Go syntax

## Installation

Clone the repository:

```bash
git clone <repository-url>
cd compiler-workshop
```

Install dependencies:

```bash
go mod tidy
```

## How to run verify.go

Use the following command to verify that the repository structure is present:

```bash
go run verify.go
```

You can also use:

```bash
make verify
```

## How to start the REPL

The REPL is planned as the interactive entry point for the language. To start it once implemented, use:

```bash
go run ./cmd/tiny
```

Or:

```bash
make run
```

## Workshop roadmap

The workshop progresses through the following milestones:

1. Tokens and the lexer
2. Parsing expressions and statements
3. Building the AST
4. Interpreting programs
5. Compiling to pseudo-assembly
6. Optional bonus work

## Exercises overview

Each exercise is organized in the exercises directory:

- 01-tokenizer
- 02-parser
- 03-ast
- 04-interpreter
- 05-compiler
- 06-bonus

Each exercise includes a README with objectives, expected outcomes, hints, and files to modify.

## Branch strategy

The repository is intended to follow a simple branch workflow:

- main: starter code and workshop materials
- exercise-1: first milestone
- exercise-2: second milestone
- exercise-3: third milestone
- exercise-4: fourth milestone
- exercise-5: fifth milestone
- complete: final reference state

## Useful Go commands

```bash
go test ./...
go fmt ./...
go run ./cmd/tiny
```

## Project milestones

- Milestone 1: token definitions and lexical analysis
- Milestone 2: parsing expressions and statements
- Milestone 3: AST construction
- Milestone 4: runtime evaluation through the interpreter
- Milestone 5: code generation through the compiler
- Milestone 6: optional improvements and extensions
