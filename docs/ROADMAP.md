# Language Evolution Roadmap

Welcome to the language evolution roadmap! Building a programming language is an exciting journey that unfolds step by step.
This roadmap outlines how a small initial prototype grows into a production-grade language capable of competing with modern systems like Go, Rust, or Python.

---

## Level 1: The Core Pipeline

*The initial foundation of the language—turning raw text into executable operations!*

- [x] **Lexical Analysis (Lexer/Scanner)**
  Converts raw source code text into a structured stream of meaningful tokens.
  This is the foundation of language processing, enabling the compiler to recognize keywords, identifiers, and symbols.
- [x] **Parsing and Abstract Syntax Tree (AST) Generation**
  Groups tokens into a hierarchical syntax tree that reflects the grammar of the program.
  This converts flat source code into a queryable data structure for downstream analysis and execution.
- [x] **Basic Expressions (Arithmetic & Logic)**
  Evaluates fundamental mathematical operations and logical comparisons (e.g., addition, subtraction, equality).
  It empowers the language to perform calculations and evaluate core conditions.
- [x] **Variable Declaration and Global State**
  Allows storing and retrieving values in named memory locations across execution contexts.
  State management is essential for building dynamic programs that track changing values.
- [x] **Basic Built-in Functions (e.g., `print`)**
  Provides core primitive utilities for outputting data and interacting with the environment.
  This gives developers immediate feedback to verify program output and state.
- [x] **Initial Code Execution (Tree-walking Interpreter & Source-to-Source Go Transpiler)**
  Executes AST nodes directly via an interpreter or translates them into Go code for compilation.
  Having dual execution paths enables rapid prototyping alongside compiled output.

---

## Level 2: Turing Completeness (The Takeaway Challenges)

*Unlocking full algorithmic expressiveness so developers can construct complex applications!*

- [x] **Control Flow (`if`/`else`, `switch`)**
  Enables conditional branching so programs can make decisions and run different code paths based on dynamic data.
  Without conditional logic, programs can only execute linearly.
- [x] **Loops (`for`/`while`)**
  Allows repeating blocks of code based on conditions or data iterations.
  Repetition is fundamental for automating repetitive tasks and processing collections of data efficiently.
- [x] **Functions and Recursion**
  Enables code reuse, scope isolation, and functional decomposition by packaging logic into callable blocks that can call themselves.
  This is critical for managing complexity and modularizing codebases.
- [x] **Complex Data Types (Arrays, Hash Maps, Structs)**
  Organizes primitive values into compound collections and custom record structures.
  Compound types allow developers to model real-world domain objects and build sophisticated software.

---

## Level 3: Bootstrapping (Self-Hosting)

*Teaching the language to compile itself and break free from external host dependencies!*

- [x] **File I/O (Reading Source Code & Writing Compiled Output)**
  Gives the language the ability to read source files from disk and write compiled text files back out.
  This bridges in-memory execution with persistent file system operations.
- [ ] **Self-Hosting (Writing the Compiler in the Language Itself)**
  Rewriting the language's compiler in the language itself so it no longer depends on Go or another host language.
  Reaching self-hosting is the ultimate proof of a language's expressiveness and stability.

---

## Level 4: Architecture & Optimization

*Scaling performance by targeting hardware instructions and optimizing execution!*

- [ ] **Intermediate Representation (IR)**
  Translates the AST into a streamlined, machine-independent instruction set designed for analysis and optimization.
  IR decouples language syntax from target machine details, enabling powerful compiler passes.
- [ ] **LLVM Backend Integration**
  Connects the compiler to LLVM to leverage industry-standard code generation and target optimization pipelines.
  This grants instant support for dozens of CPU architectures without writing custom backends.
- [ ] **Assembly Code Generation**
  Emits native CPU assembly instructions directly (such as x86_64 or ARM64) without relying on external transpilers.
  Direct machine code emission gives maximum compilation speed and low-level control.
- [ ] **Garbage Collector or Memory Management System**
  Implements automatic runtime memory reclamation (or compile-time ownership tracking).
  Automated memory management prevents leaks and memory corruption without burdening developers with manual allocation.

---

## Level 5: Tooling & Developer Experience

*Building the ecosystem tools that turn a language into a joy to write every day!*

- [ ] **Built-in Package Manager & Modules System**
  Provides dependency resolution, package downloading, and module namespacing for sharing code.
  A seamless package manager unlocks community collaboration and code reusability across projects.
- [ ] **Language Server Protocol (LSP) for IDE Support**
  Implements standard LSP features like auto-completion, hover docs, and inline syntax error diagnostics in IDEs like VS Code.
  Superior editor support dramatically boosts developer productivity and adoption.
- [ ] **Code Formatter**
  Provides an opinionated command-line tool to format source code according to standard community guidelines.
  Automated formatting eliminates code style debates and keeps codebases consistent.
- [ ] **Linter**
  Statically analyzes code to flag potential bugs, code smells, and performance pitfalls before execution.
  Static linting helps developers write safer, cleaner, and more idiomatic code.

---

## Level 6: Production & Community

*Growing from a developer tool into a thriving global ecosystem with flagship applications!*

- [ ] **Robust Standard Library**
  Offers rich built-in libraries for networking, HTTP servers, JSON parsing, file systems, and cryptography.
  A comprehensive standard library lets developers build real-world applications out of the box.
- [ ] **Official Documentation & Interactive Guides**
  Delivers thorough reference manuals, API guides, and interactive tutorials for newcomers.
  Clear documentation lowers the barrier to entry and fosters an inclusive learner community.
- [ ] **Package Registry**
  Hosts a public index and distribution network for open-source libraries published by the community.
  A central registry accelerates ecosystem growth and enables developers to share reusable packages.
- [ ] **Finding a "Killer App" (Flagship Framework or Use Case)**
  Establishes a standout framework or core domain (such as web backend frameworks, WebAssembly, or data science) that drives widespread industry adoption.
  A killer app gives developers a compelling reason to choose the language over established giants.
