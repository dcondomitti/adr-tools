package cmd

import (
	"context"
	"github.com/bmorton/adr-tools/readme"
	"github.com/google/go-github/v48/github"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

var RebuildIndexCommand = &cli.Command{
	Name:   "rebuild-index",
	Usage:  "Rebuilds the index of ADRs",
	Action: rebuildIndexAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "github-token",
			Usage:    "GitHub token for architecture repo",
			Required: true,
			EnvVars:  []string{"GITHUB_TOKEN"},
		},
		&cli.StringFlag{
			Name:     "github-repository",
			Usage:    "GitHub architecture repo as owner/repo",
			Required: true,
			EnvVars:  []string{"GITHUB_REPOSITORY"},
		},
	},
}

func rebuildIndexAction(ctx *cli.Context) error {
	token := ctx.String("github-token")
	repository := ctx.String("github-repository")

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	builder, err := readme.NewBuilder(client, repository)
	if err != nil {
		return err
	}

	err = builder.RebuildWithPullRequest(context.Background())
	if err != nil {
		return err
	}

	return nil
}
