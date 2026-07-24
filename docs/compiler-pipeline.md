# Compiler pipeline

The tiny language follows a simple pipeline:

```text
Source Code
↓
Lexer
↓
Tokens
↓
Parser
↓
AST
↓
Interpreter OR Compiler
```

## Overview

Each stage has a specific responsibility:

- The lexer reads source text and produces tokens.
- The parser turns tokens into an AST.
- The interpreter evaluates the AST directly.
- The compiler translates the AST into pseudo-assembly instructions.

## Flow summary

The source code moves through the system one stage at a time. The lexer produces the smallest meaningful pieces of input, the parser gives them structure, and the runtime layer executes or translates that structure.
