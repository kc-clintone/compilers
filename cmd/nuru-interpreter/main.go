// Command nuru-interpreter inspects and directly executes Nuru programs.
package main

import (
	"os"

	"github.com/kc-clintone/compilers/internal/cli"
)

func main() {
	os.Exit(cli.RunInterpreter(os.Args[1:]))
}
