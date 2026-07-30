// Package source defines source locations shared by every compiler phase.
package source

import "fmt"

// Position is a zero-based byte offset and one-based line/column location.
type Position struct {
	Offset int
	Line   int
	Column int
}

// Span identifies a half-open region of a source file.
type Span struct {
	Filename string
	Start    Position
	End      Position
}

// String formats the span's starting location as file:line:column.
func (s Span) String() string {
	return fmt.Sprintf("%s:%d:%d", s.Filename, s.Start.Line, s.Start.Column)
}
