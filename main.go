package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/google/go-github/v48/github"
	"github.com/namsral/flag"
	"golang.org/x/oauth2"
	"strings"
	"text/template"
	"time"
)

var ErrNoContentChange = errors.New("no content changed")

//go:embed templates/README.md.tmpl
var readmeTemplate string

func main() {
	var token string
	var repository string
	flag.StringVar(&token, "github-token", "", "GitHub token for architecture repo")
	flag.StringVar(&repository, "repository", "", "GitHub architecture repo as owner/repo")
	flag.Parse()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	builder, err := NewReadmeRebuilder(client, repository)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	err = builder.BuildWithPullRequest(context.Background())
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}

type ReadmeRebuilder struct {
	Path         string
	Owner        string
	Repo         string
	BaseBranch   string
	TargetBranch string
	gh           *github.Client
}

func (r ReadmeRebuilder) BuildWithPullRequest(ctx context.Context) error {
	fmt.Printf("Building for %s/%s...\n", r.Owner, r.Repo)
	fmt.Printf("-- Loading decisions...\n")
	decisions, err := r.loadDecisions(ctx, branchRef(r.BaseBranch, true))
	if err != nil {
		return err
	}

	fmt.Printf("-- Building README.md...\n")
	newContent, err := r.regenerateReadme(ctx, decisions)
	if err != nil {
		return err
	}

	fmt.Printf("-- Committing changes...\n")
	_, err = r.commitToBranch(ctx, newContent, "Refreshed list of decisions", branchRef(r.BaseBranch, false))
	if err != nil {
		return err
	}

	fmt.Printf("-- Creating pull request...\n")
	create, err := r.createPullRequest(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Success!\n%s\n", create.GetHTMLURL())

	return nil
}

func NewReadmeRebuilder(gh *github.Client, repository string) (*ReadmeRebuilder, error) {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository: %s", repository)
	}

	return &ReadmeRebuilder{
		gh:           gh,
		BaseBranch:   "main",
		TargetBranch: fmt.Sprintf("adr-tools/readme-update/%d", time.Now().Unix()),
		Path:         "README.md",
		Owner:        parts[0],
		Repo:         parts[1],
	}, nil
}

func (r ReadmeRebuilder) getRef(ctx context.Context, ref string) (*github.Reference, error) {
	gitRef, _, err := r.gh.Git.GetRef(ctx, r.Owner, r.Repo, ref)
	return gitRef, err
}

func (r ReadmeRebuilder) loadDecisions(ctx context.Context, sha string) ([]*Decision, error) {
	var decisions []*Decision

	_, dir, _, err := r.gh.Repositories.GetContents(ctx, r.Owner, r.Repo, "/decisions",
		&github.RepositoryContentGetOptions{
			Ref: sha,
		},
	)
	if err != nil {
		return decisions, err
	}

	for _, decision := range dir {
		content, err := r.getContent(ctx, sha, decision.GetPath())
		if err != nil {
			return decisions, err
		}
		decisions = append(decisions, &Decision{
			RawContent: content,
			Path:       decision.GetPath(),
			GithubURL:  decision.GetHTMLURL(),
		})
	}

	return decisions, nil
}

func (r ReadmeRebuilder) getContent(ctx context.Context, sha string, path string) (string, error) {
	file, _, _, err := r.gh.Repositories.GetContents(ctx, r.Owner, r.Repo, path,
		&github.RepositoryContentGetOptions{
			Ref: sha,
		},
	)
	if err != nil {
		return "", nil
	}

	return file.GetContent()
}

func (r ReadmeRebuilder) createPullRequest(ctx context.Context) (*github.PullRequest, error) {
	baseBranch := branchRef(r.BaseBranch, true)
	targetBranch := branchRef(r.TargetBranch, true)
	title := "Refreshed list of decisions in README.md"
	newPull := &github.NewPullRequest{
		Title: &title,
		Base:  &baseBranch,
		Head:  &targetBranch,
	}
	create, _, err := r.gh.PullRequests.Create(ctx, r.Owner, r.Repo, newPull)
	if err != nil {
		return nil, err
	}
	return create, nil
}

func (r ReadmeRebuilder) commitToBranch(ctx context.Context, content string, message string, baseRef string) (*github.Reference, error) {
	ref, err := r.getRef(ctx, baseRef)

	blob := &github.Blob{
		Content: &content,
	}
	createBlob, _, err := r.gh.Git.CreateBlob(ctx, r.Owner, r.Repo, blob)
	if err != nil {
		return nil, err
	}

	fileMode := "100644"
	newTree := []*github.TreeEntry{
		{
			Path: &r.Path,
			Mode: &fileMode,
			SHA:  createBlob.SHA,
		},
	}
	tree, _, err := r.gh.Git.CreateTree(ctx, r.Owner, r.Repo, ref.GetObject().GetSHA(), newTree)
	if err != nil {
		return nil, err
	}

	newParentCommit := &github.Commit{
		SHA: ref.GetObject().SHA,
	}

	newCommit := &github.Commit{
		Parents: []*github.Commit{newParentCommit},
		Message: &message,
		Tree:    tree,
	}
	commit, _, err := r.gh.Git.CreateCommit(ctx, r.Owner, r.Repo, newCommit)
	if err != nil {
		return nil, err
	}

	targetBranch := branchRef(r.TargetBranch, true)
	newBranch := &github.Reference{
		Ref: &targetBranch,
		Object: &github.GitObject{
			SHA: commit.SHA,
		},
	}
	newRef, _, err := r.gh.Git.CreateRef(ctx, r.Owner, r.Repo, newBranch)
	return newRef, err
}

func (r ReadmeRebuilder) generateReadme(decisions []*Decision) (string, error) {
	t := template.New("README.md")
	t = template.Must(t.Parse(readmeTemplate))

	buf := bytes.NewBufferString("")
	err := t.Execute(buf, TemplateData{
		Decisions:   decisions,
		Name:        "Architecture",
		Description: "Our list of decisions",
	})

	return buf.String(), err
}

func (r ReadmeRebuilder) regenerateReadme(ctx context.Context, decisions []*Decision) (string, error) {
	newContent, err := r.generateReadme(decisions)
	if err != nil {
		return "", err
	}

	content, err := r.getContent(ctx, branchRef(r.BaseBranch, true), "/README.md")
	if err != nil {
		return "", err
	}

	if newContent == content {
		return "", ErrNoContentChange
	}

	return newContent, nil
}

func branchRef(name string, prefix bool) string {
	var format string
	if prefix {
		format = "refs/heads/%s"
	} else {
		format = "heads/%s"
	}
	return fmt.Sprintf(format, name)
}
