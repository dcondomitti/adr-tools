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
		&cli.BoolFlag{
			Name:  "pull-request",
			Usage: "Create a pull request",
			Value: true,
		},
		&cli.StringFlag{
			Name:        "target-branch",
			Usage:       "Target branch for rebuild commit",
			DefaultText: "random branch name",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "Render README to stdout instead of committing",
			Value: false,
		},
	},
}

func rebuildIndexAction(ctx *cli.Context) error {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ctx.String("github-token")},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	builder, err := readme.NewBuilder(client, ctx.String("github-repository"))
	if err != nil {
		return err
	}
	builder.DryRun = ctx.Bool("dry-run")
	builder.CreatePullRequest = ctx.Bool("pull-request")
	if targetBranch := ctx.String("target-branch"); targetBranch != "" {
		builder.TargetBranch = targetBranch
		builder.CreateBranch = false
	}

	err = builder.RebuildWithPullRequest(context.Background())
	if err != nil {
		return err
	}

	return nil
}
