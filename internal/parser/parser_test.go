package parser

import (
	"testing"
)

func TestParserScaffold(t *testing.T) {
	input := `+ -`
	prog, _ := Parse("test.zing", []byte(input))
	if prog == nil {
		t.Fatalf("expected non-nil program")
	}
}
