package main

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type z_Token struct {
	z_kind   string
	z_text   string
	z_line   int
	z_column int
}

func z_isLetter(z_c byte) bool {
	if (z_c >= byte('a')) && (z_c <= byte('z')) {
		return true
	}
	if (z_c >= byte('A')) && (z_c <= byte('Z')) {
		return true
	}
	return (z_c == byte('_'))
}

func z_isDigit(z_c byte) bool {
	return ((z_c >= byte('0')) && (z_c <= byte('9')))
}

var z_paths []string = os.Args[1:]

var z_input string

var z_tokens []*z_Token = nuruMakeSlice[*z_Token](0, "examples/source-analyzer.nuru:25:22")

var z_counts map[string]int = make(map[string]int)

var z_i int = 0

var z_line int = 1

var z_column int = 1

func main() {
	if len(z_paths) == 0 {
		nuruFail("examples/source-analyzer.nuru:32:5", "usage: source-analyzer <file>")
	}
	z_input = nuruReadFile("examples/source-analyzer.nuru:34:9", nuruIndexSlice(z_paths, 0, "examples/source-analyzer.nuru:34:18"))
	for z_i < len(z_input) {
		var z_c byte = nuruIndexString(z_input, z_i, "examples/source-analyzer.nuru:37:18")
		if ((z_c == byte(' ')) || (z_c == byte('\t'))) || (z_c == byte('\r')) {
			z_i = (z_i + 1)
			z_column = (z_column + 1)
			continue
		}
		if z_c == byte('\n') {
			z_i = (z_i + 1)
			z_line = (z_line + 1)
			z_column = 1
			continue
		}
		var z_start int = z_i
		var z_startColumn int = z_column
		var z_kind string = "symbol"
		if z_isLetter(z_c) {
			z_kind = "identifier"
			for (z_i < len(z_input)) && (z_isLetter(nuruIndexString(z_input, z_i, "examples/source-analyzer.nuru:54:41")) || z_isDigit(nuruIndexString(z_input, z_i, "examples/source-analyzer.nuru:54:62"))) {
				z_i = (z_i + 1)
				z_column = (z_column + 1)
			}
		} else {
			if z_isDigit(z_c) {
				z_kind = "integer"
				for (z_i < len(z_input)) && z_isDigit(nuruIndexString(z_input, z_i, "examples/source-analyzer.nuru:61:43")) {
					z_i = (z_i + 1)
					z_column = (z_column + 1)
				}
			} else {
				z_i = (z_i + 1)
				z_column = (z_column + 1)
			}
		}
		var z_text string = nuruSliceString(z_input, z_start, z_i, "examples/source-analyzer.nuru:70:23")
		z_tokens = append(z_tokens, &z_Token{z_kind: z_kind, z_text: z_text, z_line: z_line, z_column: z_startColumn})
		nuruSetMap(z_counts, z_kind, (z_counts[z_kind] + 1), "examples/source-analyzer.nuru:72:5")
	}
	if len(z_tokens) > 0 {
		nuruIndexSlice(z_tokens, 0, "examples/source-analyzer.nuru:78:5").z_text = nuruSliceString(z_input, 0, len(nuruIndexSlice(z_tokens, 0, "examples/source-analyzer.nuru:78:34").z_text), "examples/source-analyzer.nuru:78:22")
	}
	for z_j := 0; z_j < len(z_tokens); z_j = (z_j + 1) {
		nuruPrint(nuruIndexSlice(z_tokens, z_j, "examples/source-analyzer.nuru:82:11").z_line, nuruIndexSlice(z_tokens, z_j, "examples/source-analyzer.nuru:82:27").z_column, nuruIndexSlice(z_tokens, z_j, "examples/source-analyzer.nuru:82:45").z_kind, nuruIndexSlice(z_tokens, z_j, "examples/source-analyzer.nuru:82:61").z_text)
	}
	nuruPrint("tokens", len(z_tokens))
	nuruPrint("identifiers", z_counts["identifier"])
	nuruPrint("integers", z_counts["integer"])
	nuruPrint("symbols", z_counts["symbol"])
}

func nuruFail(location, message string) {
	fmt.Fprintln(os.Stderr, location+": runtime: "+message)
	os.Exit(1)
}

func nuruReadFile(location, path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		nuruFail(location, err.Error())
	}
	return string(data)
}

func nuruIndexSlice[T any](value []T, index int, location string) T {
	if index < 0 || index >= len(value) {
		nuruFail(location, "index out of bounds")
	}
	return value[index]
}

func nuruIndexString(value string, index int, location string) byte {
	if index < 0 || index >= len(value) {
		nuruFail(location, "index out of bounds")
	}
	return value[index]
}

func nuruSliceString(value string, low, high int, location string) string {
	if high < 0 {
		high = len(value)
	}
	if low < 0 || high < low || high > len(value) {
		nuruFail(location, "slice bounds out of range")
	}
	return value[low:high]
}

func nuruSetMap[K comparable, V any](value map[K]V, key K, item V, location string) {
	if value == nil {
		nuruFail(location, "assignment to uninitialized map")
	}
	value[key] = item
}

func nuruMakeSlice[T any](length int, location string) []T {
	if length < 0 {
		nuruFail(location, "negative slice size")
	}
	return make([]T, length)
}

func nuruPrint(values ...any) {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = nuruDisplay(reflect.ValueOf(value))
	}
	fmt.Println(strings.Join(parts, " "))
}
func nuruDisplay(value reflect.Value) string {
	if !value.IsValid() {
		return "<void>"
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return "<nil>"
		}
		return strings.TrimPrefix(value.Elem().Type().Name(), "z_")
	}
	switch value.Kind() {
	case reflect.Int:
		return strconv.FormatInt(value.Int(), 10)
	case reflect.Uint8:
		return string([]byte{byte(value.Uint())})
	case reflect.String:
		return value.String()
	case reflect.Bool:
		return strconv.FormatBool(value.Bool())
	case reflect.Slice:
		parts := make([]string, value.Len())
		for i := range parts {
			parts[i] = nuruDisplay(value.Index(i))
		}
		return "[" + strings.Join(parts, " ") + "]"
	case reflect.Map:
		return fmt.Sprintf("map[%d entries]", value.Len())
	}
	return "<invalid>"
}
