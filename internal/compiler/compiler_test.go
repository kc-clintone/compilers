package compiler

import (
	"strings"
	"testing"

	"github.com/kc-clintone/compilers/internal/parser"
)

func TestCompilerBasic(t *testing.T) {
	input := `print(1 + 2)`
	prog, diags := parser.Parse("test.zing", []byte(input))
	if len(diags) > 0 {
		t.Fatalf("unexpected parser diagnostics: %v", diags)
	}

	c := New()
	goCode, compDiags := c.Compile(prog)
	if len(compDiags) > 0 {
		t.Fatalf("unexpected compiler diagnostics: %v", compDiags)
	}

	if !strings.Contains(goCode, "fmt.Println((1 + 2))") {
		t.Errorf("generated Go code does not contain expected fmt.Println: %s", goCode)
	}
}
