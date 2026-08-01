// Command nuru-compiler checks, transpiles, and builds Nuru programs.
package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/kc-clintone/compilers/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	defer stop()
	os.Exit(cli.RunCompiler(ctx, os.Args[1:], cli.Streams{Stdout: os.Stdout, Stderr: os.Stderr}))
}
