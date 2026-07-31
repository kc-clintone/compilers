package compiler

import (
	"strings"
	"testing"

	"github.com/kc-clintone/compilers/internal/ast"
)

func TestCompilerScaffold(t *testing.T) {
	prog := &ast.Program{}
	c := New()
	goCode, diags := c.Compile(prog)
	if len(diags) > 0 {
		t.Fatalf("unexpected compiler diagnostics: %v", diags)
	}

	if !strings.Contains(goCode, "package main") {
		t.Errorf("expected package main in output: %s", goCode)
	}
}
