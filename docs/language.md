# Zing v1 Language Reference

Zing is a small, statically typed teaching language. Programs contain type,
function, and global variable declarations followed by executable top-level
statements. The transpiler turns those statements into Go's `main` function.

The primitive types are `int`, `char`, `string`, and `bool`. Composite types are
slices (`[]T`), maps (`map[K]V`), and named structs. Characters and string
indexing are byte-oriented. Structs, slices, and maps have reference semantics.

Variables use `var name Type` with an optional initializer. Functions have typed
parameters and zero or one result. Zing supports `if`, non-fallthrough `switch`,
while-style and three-clause `for`, `break`, `continue`, and `return`.

Composite values use Go-like syntax: `Token{kind: 1}`, `[]int{1, 2}`,
`map[string]int{"x": 1}`, `make([]Token, 0)`, and `make(map[string]int)`.

The built-ins are `print`, `args`, `readFile`, `writeFile`, `len`, `append`,
`int`, `char`, `string`, and `fail`. These operations are sufficient for a
future Zing implementation to read source, tokenize it, construct an AST, check
it, and emit Go. Self-hosting is a later milestone, not part of v1.

Conversions are explicit. `int(char)` returns the byte value and `int(string)`
parses a signed base-10 Go-sized integer. `char(int)` accepts only values from
0 through 255. `string(int)` formats a decimal integer, `string(char)` produces
one byte, and `string(bool)` produces `true` or `false`. Invalid or overflowing
conversions fail at runtime with a source-located diagnostic.

Diagnostics have the form `file:line:column: phase: message`.
