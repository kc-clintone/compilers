// Package diagnostic defines source-located diagnostics shared by the Zing
// front end and runtime.
package diagnostic

import (
	"fmt"
	"sort"

	"github.com/kc-clintone/compilers/internal/source"
)

// Diagnostic is a source-located lexer, parser, checker, or runtime message.
type Diagnostic struct {
	Span    source.Span
	Phase   string
	Message string
}

// Error formats a diagnostic using Zing's stable user-facing representation.
func (d Diagnostic) Error() string {
	return fmt.Sprintf("%s: %s: %s", d.Span, d.Phase, d.Message)
}

// Sort orders diagnostics by filename and source offset while preserving the
// emission order of diagnostics at the same location.
func Sort(diagnostics []Diagnostic) {
	sort.SliceStable(diagnostics, func(i, j int) bool {
		if diagnostics[i].Span.Filename != diagnostics[j].Span.Filename {
			return diagnostics[i].Span.Filename < diagnostics[j].Span.Filename
		}

		return diagnostics[i].Span.Start.Offset < diagnostics[j].Span.Start.Offset
	})
}
