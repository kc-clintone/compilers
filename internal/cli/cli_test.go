package cli

import "testing"

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
