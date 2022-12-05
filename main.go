package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/bmorton/adr-tools/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:     "adr-tools",
		Usage:    "A tool for working with Architecture Decision Records",
		Commands: []*cli.Command{cmd.RebuildIndexCommand, cmd.NewCommand},
	}

	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}
