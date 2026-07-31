// Command nuru-compiler inspects, transpiles, and builds Nuru programs.
package main

import (
	"os"

	"github.com/kc-clintone/compilers/internal/cli"
)

func main() {
	os.Exit(cli.RunCompiler(os.Args[1:]))
}
