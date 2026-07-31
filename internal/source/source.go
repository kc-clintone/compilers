package source

import "fmt"

// Position represents a specific location in a source file.
type Position struct {
	Offset int
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// Span represents a half-open range [Start, End) within a source file.
type Span struct {
	Filename string
	Start    Position
	End      Position
}

func (s Span) String() string {
	if s.Filename == "" {
		return fmt.Sprintf("%s-%s", s.Start, s.End)
	}
	return fmt.Sprintf("%s:%s-%s", s.Filename, s.Start, s.End)
}
