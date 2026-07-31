# Zing v1 language reference

Zing is a small, statically typed language. A source file contains type,
function, and global-variable declarations followed by executable top-level
statements. The interpreter executes those statements directly; the compiler
places them in generated Go `main`.

## Lexical structure

Files use the `.zing` extension. Identifiers contain ASCII letters, digits, and
underscores and cannot start with a digit. `main` is reserved everywhere, and
built-in names cannot be redeclared in the value namespace. Whitespace, `//`
line comments, and non-nesting `/* ... */` comments are ignored.

Integer literals are non-negative decimal Go-sized integers; negation is a
unary operator. Strings and characters support `\n`, `\r`, `\t`, `\\`, `\"`,
and `\'`. A character contains exactly one decoded byte.

## Types and declarations

Primitive types are `int`, `char`, `string`, and `bool`. Composite types are
slices (`[]T`), maps (`map[K]V`), and named structs. Map keys are restricted to
primitive types. Characters and all string indexing/slicing are byte-oriented.

```zing
type Token struct {
    text string;
    line int;
}

var count int;
var token Token = Token{text: "var", line: 1};
var tokens []Token = []Token{token};
var counts map[string]int = make(map[string]int);
```

Primitive values use value semantics. Slices copy their headers while sharing
backing storage; maps and struct instances have reference semantics. Missing
map keys return the value type's zero value. Assigning through an uninitialized
map is a runtime error. Named struct variables always require an initializer.

Functions have typed parameters and either one result type or no result. They
may recurse and call later-declared functions. Functions are not values and
cannot be nested.

## Statements and expressions

Zing supports local variables, assignment, expression statements, blocks,
`if`/`else`, non-fallthrough `switch`, while-style `for`, three-clause `for`,
`break`, `continue`, and `return`.

```zing
for value > 0 {
    value = value - 1;
}

for var i int = 0; i < len(tokens); i = i + 1 {
    print(tokens[i].text);
}
```

Operator precedence from low to high is `||`, `&&`, equality, ordering,
addition/subtraction, multiplication/division/modulo, prefix `!`/`-`, and
postfix call/index/slice/field access. Binary operators associate left. `&&`
and `||` short-circuit; other operands and function arguments evaluate left to
right. Assignment evaluates its target container and index once, before its
right-hand value.

Arithmetic accepts integers, except `+` also concatenates strings. Ordering
accepts matching integers, characters, or strings. Equality accepts matching
primitive values and compares struct references by identity. Conditions must
be Boolean, and assignments and returns require identical types.

## Built-ins

- `print(values...)` prints stable space-separated representations and a newline.
- `args() []string` returns program arguments.
- `readFile(path string) string` and `writeFile(path, contents string)` perform complete-file I/O.
- `len(value) int` accepts strings, slices, and maps.
- `append(slice, value)` returns a new slice header.
- `fail(message string)` raises a diagnosed runtime failure.

Conversions are explicit. `int(char)` returns the byte value and `int(string)`
parses a signed base-10 Go-sized integer. `char(int)` accepts 0 through 255.
`string(int)` formats decimal, `string(char)` produces one byte, and
`string(bool)` produces `true` or `false`. Invalid and overflowing conversions
are runtime failures.

## Diagnostics and tools

Diagnostics use `file:line:column: phase: message`, where phase is `lexer`,
`parser`, `checker`, or `runtime`. A phase does not run after its prerequisite
reports errors.

Use `zing-interpreter` (or `zing-interpreter check`) for direct execution and
`zing-compiler` (or `zing-compiler check/transpile`) for Go generation. Both tools return 0 on
success, 1 for program/build failures, and 2 for invalid usage.
