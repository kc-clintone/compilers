package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseOutputArgsDefaults(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		fallback   func(string) string
		wantInput  string
		wantOutput string
	}{
		{"transpile", []string{"some/path/demo.nuru"}, func(string) string { return "demo.nuru.go" }, "some/path/demo.nuru", "demo.nuru.go"},
		{"compile", []string{"some/path/demo.nuru"}, func(string) string { return "nuru.out" }, "some/path/demo.nuru", "nuru.out"},
		{"explicit", []string{"-o", "custom", "demo.nuru"}, func(string) string { return "unused" }, "demo.nuru", "custom"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input, output, unrecognized := parseOutputArgs(test.args, test.fallback)
			if input != test.wantInput || output != test.wantOutput || len(unrecognized) != 0 {
				t.Fatalf("input=%q output=%q unrecognized=%q", input, output, unrecognized)
			}
		})
	}
}

func TestHelpDocumentsCommandsAndOptions(t *testing.T) {
	tests := []struct {
		name  string
		print func(*bytes.Buffer)
		want  []string
	}{
		{
			name: "compiler",
			print: func(output *bytes.Buffer) {
				printCompilerHelp(output)
			},
			want: []string{"NAME", "SYNOPSIS", "DESCRIPTION", "COMMANDS", "OPTIONS", "tokens FILE, lex FILE", "ast FILE, parse FILE", "transpile [-o GO-FILE] FILE", "-o PATH", "-h, --help", "nuru.out"},
		},
		{
			name: "interpreter",
			print: func(output *bytes.Buffer) {
				printInterpreterHelp(output)
			},
			want: []string{"NAME", "SYNOPSIS", "DESCRIPTION", "COMMANDS", "OPTIONS", "tokens FILE, lex FILE", "ast FILE, parse FILE", "repl", "-h, --help"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			test.print(&output)
			for _, want := range test.want {
				if !strings.Contains(output.String(), want) {
					t.Errorf("help does not contain %q:\n%s", want, output.String())
				}
			}
		})
	}
}
