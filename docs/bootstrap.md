# Bootstrap-readiness matrix

| Future compiler need | Zing v1 capability |
| --- | --- |
| Read source and arguments | `args`, `readFile` |
| Lex bytes | `char`, string indexing/slicing, loops |
| Store tokens and AST nodes | structs and slices |
| Symbol tables | maps |
| Recursive parsing and checking | typed recursive functions |
| Report errors | `print`, conversions, `fail` |
| Emit target source | strings, slices, `writeFile` |

The Go implementation is the seed implementation. Rewriting it in Zing is a
separate future milestone rather than part of the v1 reference.
