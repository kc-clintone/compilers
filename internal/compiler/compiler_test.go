package compiler

import (
	"bytes"
	"context"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kc-clintone/compilers/internal/checker"
	nuruparser "github.com/kc-clintone/compilers/internal/parser"
)

func TestGenerateIsFormattedDeterministicAndParsable(t *testing.T) {
	source := `
type Box struct { value int; }
var box Box = Box{value: 2};
func double(value int) int { return value * 2; }
print(double(box.value));`
	first := generateSource(t, source)
	second := generateSource(t, source)

	if !bytes.Equal(first, second) {
		t.Fatal("generation is not deterministic")
	}

	formatted, err := format.Source(first)
	if err != nil || !bytes.Equal(first, formatted) {
		t.Fatalf("generated source is not gofmt formatted: %v", err)
	}

	if _, err := parser.ParseFile(token.NewFileSet(), "main.go", first, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, first)
	}

	for _, fragment := range []string{"type z_Box struct", "func z_double", "func main()", "nuruPrint"} {
		if !bytes.Contains(first, []byte(fragment)) {
			t.Fatalf("generated source does not contain %q\n%s", fragment, first)
		}
	}
}

func TestGenerateSelectsOnlyRequiredHelpersAndImports(t *testing.T) {
	plain := generateSource(t, `var value int = 1; value = value + 1;`)

	if bytes.Contains(plain, []byte("import (")) || bytes.Contains(plain, []byte("nuruFail")) {
		t.Fatalf("plain program has unused runtime code:\n%s", plain)
	}

	runtime := generateSource(t, `var values []int = []int{4}; print(values[0], 4 / 2, char(65), int("2"));`)

	for _, fragment := range []string{"nuruIndexSlice", "nuruDiv", "nuruChar", "nuruAtoi", "nuruFail", `"reflect"`, `"strconv"`} {
		if !bytes.Contains(runtime, []byte(fragment)) {
			t.Fatalf("runtime program does not contain %q\n%s", fragment, runtime)
		}
	}
}

func TestBuildProducesExecutableAndCleansTemporarySource(t *testing.T) {
	temporaryRoot := t.TempDir()

	t.Setenv("TMPDIR", temporaryRoot)
	output := filepath.Join(t.TempDir(), "program")

	if err := Build(context.Background(), generateSource(t, `print("ok");`), output); err != nil {
		t.Fatal(err)
	}

	result, err := exec.Command(output).CombinedOutput()
	if err != nil || string(result) != "ok\n" {
		t.Fatalf("binary result = %q, %v", result, err)
	}

	entries, err := os.ReadDir(temporaryRoot)
	if err != nil || len(entries) != 0 {
		t.Fatalf("temporary build directory leaked: %v, %v", entries, err)
	}
}

func TestBuildReportsToolchainAndCancellationErrors(t *testing.T) {
	t.Run("missing toolchain", func(t *testing.T) {
		t.Setenv("PATH", "")
		err := Build(context.Background(), []byte("package main\nfunc main() {}\n"), filepath.Join(t.TempDir(), "program"))

		if err == nil || !strings.Contains(err.Error(), "go") {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		cancel()
		err := Build(ctx, []byte("package main\nfunc main() {}\n"), filepath.Join(t.TempDir(), "program"))

		if err == nil {
			t.Fatal("Build succeeded with a canceled context")
		}
	})
}

func generateSource(t *testing.T, source string) []byte {
	t.Helper()
	program, diagnostics := nuruparser.Parse("test.nuru", []byte(source))

	if len(diagnostics) != 0 {
		t.Fatalf("parse: %v", diagnostics)
	}

	info, diagnostics := checker.Check(program)

	if len(diagnostics) != 0 {
		t.Fatalf("check: %v", diagnostics)
	}

	generated, err := Generate(program, info)
	if err != nil {
		t.Fatal(err)
	}

	return generated
}
