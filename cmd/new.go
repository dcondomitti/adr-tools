package cmd

import (
	"strings"

	"github.com/bmorton/adr-tools/decisions"
	"github.com/urfave/cli/v2"
)

var NewCommand = &cli.Command{
	Name:   "new",
	Usage:  "Creates a new ADR",
	Action: newAction,
	Flags:  []cli.Flag{},
}

func newAction(ctx *cli.Context) error {
	title := strings.Join(ctx.Args().Slice()[0:], " ")
	d := decisions.NewBuilder(title)

	d.GenerateDecision()
	return nil
}
