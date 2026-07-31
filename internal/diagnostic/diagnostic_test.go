package diagnostic

import (
	"reflect"
	"testing"

	"github.com/kc-clintone/compilers/internal/source"
)

func TestDiagnosticError(t *testing.T) {
	diagnostic := Diagnostic{
		Span: source.Span{
			Filename: "example.nuru",
			Start:    source.Position{Offset: 12, Line: 3, Column: 5},
			End:      source.Position{Offset: 15, Line: 3, Column: 8},
		},
		Phase:   "checker",
		Message: "unknown variable value",
	}
	if got, want := diagnostic.Error(), "example.nuru:3:5: checker: unknown variable value"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestSortOrdersByFilenameAndOffsetStably(t *testing.T) {
	diagnostics := []Diagnostic{
		makeDiagnostic("b.nuru", 2, "b-two"),
		makeDiagnostic("a.nuru", 8, "a-eight"),
		makeDiagnostic("a.nuru", 3, "a-three-first"),
		makeDiagnostic("b.nuru", 1, "b-one"),
		makeDiagnostic("a.nuru", 3, "a-three-second"),
	}
	Sort(diagnostics)

	got := make([]string, len(diagnostics))
	for index, diagnostic := range diagnostics {
		got[index] = diagnostic.Message
	}
	want := []string{"a-three-first", "a-three-second", "a-eight", "b-one", "b-two"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sorted messages = %v, want %v", got, want)
	}
}

func TestSortAcceptsEmptySlices(t *testing.T) {
	Sort(nil)
	empty := []Diagnostic{}
	Sort(empty)
	if len(empty) != 0 {
		t.Fatalf("empty slice length = %d", len(empty))
	}
}

func makeDiagnostic(filename string, offset int, message string) Diagnostic {
	return Diagnostic{
		Span:    source.Span{Filename: filename, Start: source.Position{Offset: offset}},
		Phase:   "test",
		Message: message,
	}
}
