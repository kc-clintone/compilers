package interpreter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kc-clintone/compilers/internal/checker"
	"github.com/kc-clintone/compilers/internal/parser"
)

func TestValuesScopesAndControlFlow(t *testing.T) {
	source := `
type Box struct { value int; }
var box Box = Box{value: 1};
var alias Box = box;
var values []int = []int{1};
var sliceAlias []int = values;
var counts map[string]int = make(map[string]int);
var calls int = 0;
func touch() bool { calls = calls + 1; return true; }
func factorial(n int) int { if n <= 1 { return 1; } return n * factorial(n - 1); }
alias.value = 2;
values = append(values, 3);
false && touch();
true || touch();
for var i int = 0; i < 3; i = i + 1 { if i == 1 { continue; } counts["sum"] = counts["sum"] + i; }
print(box.value, len(sliceAlias), len(values), counts["missing"], counts["sum"], calls, factorial(5));`

	output, err := runSource(context.Background(), source, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := output, "2 1 2 0 2 0 120\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestAssignmentTargetIsEvaluatedBeforeValueExactlyOnce(t *testing.T) {
	source := `
type Box struct { value int; }
var box Box = Box{value: 0};
var order string = "";
func target() Box { order = order + "T"; return box; }
func value() int { order = order + "V"; return 7; }
target().value = value();
print(order, box.value);`
	output, err := runSource(context.Background(), source, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if output != "TV 7\n" {
		t.Fatalf("output = %q, want target-before-value ordering", output)
	}
}

func TestInjectedArgumentsAndFileSystem(t *testing.T) {
	files := &memoryFiles{data: map[string][]byte{"input": []byte("hello")}}
	source := `
var values []string = args();
var text string = readFile(values[0]);
writeFile(values[1], text + "!");
print(text);`
	output, err := runSource(context.Background(), source, Options{Args: []string{"input", "output"}, Files: files})
	if err != nil {
		t.Fatal(err)
	}
	if output != "hello\n" || string(files.data["output"]) != "hello!" {
		t.Fatalf("output = %q, files = %#v", output, files.data)
	}
}

func TestRuntimeErrorsCarrySourceSpans(t *testing.T) {
	tests := map[string]struct {
		source, message string
	}{
		"division":                  {"print(1 / 0);", "division by zero"},
		"bounds":                    {"var x []int = []int{1}; print(x[2]);", "index out of bounds"},
		"slice bounds":              {"var x string = \"a\"; print(x[1:0]);", "slice bounds out of range"},
		"integer conversion":        {`print(int("no"));`, "invalid integer conversion"},
		"character conversion low":  {`print(char(-1));`, "invalid character conversion"},
		"character conversion high": {`print(char(256));`, "invalid character conversion"},
		"explicit fail":             {`fail("stopped");`, "stopped"},
		"nil map assignment":        {`var values map[string]int; values["x"] = 1;`, "assignment to uninitialized map"},
		"file":                      {`print(readFile("missing"));`, "missing file"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			options := Options{}
			if name == "file" {
				options.Files = failingFiles{}
			}
			_, err := runSource(context.Background(), test.source, options)
			var runtimeErr *RuntimeError
			if !errors.As(err, &runtimeErr) || !strings.Contains(runtimeErr.Message, test.message) {
				t.Fatalf("error = %v, want runtime error containing %q", err, test.message)
			}
			if runtimeErr.Span.Filename != "test.zing" || runtimeErr.Span.Start.Line < 1 || runtimeErr.Span.Start.Column < 1 {
				t.Fatalf("runtime span = %v", runtimeErr.Span)
			}
		})
	}
}

func TestCancellationStopsInfiniteLoop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := runSource(ctx, `for true { }`, Options{})
	var runtimeErr *RuntimeError
	if !errors.As(err, &runtimeErr) || !strings.Contains(runtimeErr.Message, context.DeadlineExceeded.Error()) {
		t.Fatalf("error = %v, want cancellation runtime error", err)
	}
}

func runSource(ctx context.Context, source string, options Options) (string, error) {
	program, diagnostics := parser.Parse("test.zing", []byte(source))
	if len(diagnostics) != 0 {
		return "", fmt.Errorf("parse: %v", diagnostics)
	}
	info, diagnostics := checker.Check(program)
	if len(diagnostics) != 0 {
		return "", fmt.Errorf("check: %v", diagnostics)
	var output bytes.Buffer
	options.Stdout = &output
	err := Run(ctx, program, info, options)
	return output.String(), err
}

func TestExportScopeUpgrading(t *testing.T) {
	source := `
export var x int = 100;
func foo() {
	export var y int = 200;
}
foo();
print(x, y);
`
	output, err := runSource(context.Background(), source, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := output, "100 200\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestRunREPL(t *testing.T) {
	input := "var a int = 10;\na + 5;\nexit\n"
	var outBuf, errBuf bytes.Buffer
	options := Options{
		Stdin:  strings.NewReader(input),
		Stdout: &outBuf,
		Stderr: &errBuf,
	}
	err := RunREPL(context.Background(), nil, nil, options)
	if err != nil {
		t.Fatalf("unexpected REPL error: %v", err)
	}
	if !strings.Contains(outBuf.String(), "15") {
		t.Fatalf("REPL output = %q, want it to contain 15", outBuf.String())
	}
}

type memoryFiles struct{ data map[string][]byte }

func (files *memoryFiles) ReadFile(name string) ([]byte, error) {
	data, ok := files.data[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return append([]byte(nil), data...), nil
}

func (files *memoryFiles) WriteFile(name string, data []byte, _ os.FileMode) error {
	files.data[name] = append([]byte(nil), data...)
	return nil
}

type failingFiles struct{}

func (failingFiles) ReadFile(string) ([]byte, error) { return nil, errors.New("missing file") }
func (failingFiles) WriteFile(string, []byte, os.FileMode) error {
	return errors.New("write failed")
}
