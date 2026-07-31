package diagnostic

import (
	"fmt"

	"github.com/kc-clintone/compilers/internal/source"
)

// Diagnostic represents an error or warning emitted during lexing, parsing, evaluation, or compilation.
type Diagnostic struct {
	Span    source.Span
	Phase   string
	Message string
}

func (d Diagnostic) String() string {
	if d.Span.Filename != "" {
		return fmt.Sprintf("[%s] %s: %s", d.Phase, d.Span, d.Message)
	}
	return fmt.Sprintf("[%s] %s", d.Phase, d.Message)
}
