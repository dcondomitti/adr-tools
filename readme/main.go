package readme

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/bmorton/adr-tools/schema"
	"github.com/google/go-github/v48/github"
	"sort"
	"strings"
	"text/template"
	"time"
)

var ErrNoContentChange = errors.New("no content changed")

//go:embed templates/README.md.tmpl
var readmeTemplate string

type TemplateData struct {
	Decisions   []*schema.Decision
	Name        string
	Description string
}

type Builder struct {
	Path         string
	Owner        string
	Repo         string
	BaseBranch   string
	TargetBranch string
	gh           *github.Client
}

func NewBuilder(gh *github.Client, repository string) (*Builder, error) {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository: %s", repository)
	}

	return &Builder{
		gh:           gh,
		BaseBranch:   "main",
		TargetBranch: fmt.Sprintf("adr-tools/readme-update/%d", time.Now().Unix()),
		Path:         "README.md",
		Owner:        parts[0],
		Repo:         parts[1],
	}, nil
}

func (r Builder) RebuildWithPullRequest(ctx context.Context) error {
	fmt.Printf("Building for %s/%s...\n", r.Owner, r.Repo)
	fmt.Printf("-- Loading decisions...\n")
	decisions, err := r.loadDecisions(ctx)
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

func (r Builder) getRef(ctx context.Context, ref string) (*github.Reference, error) {
	gitRef, _, err := r.gh.Git.GetRef(ctx, r.Owner, r.Repo, ref)
	return gitRef, err
}

func (r Builder) loadDecisions(ctx context.Context) (map[string]*schema.Decision, error) {
	decisions := make(map[string]*schema.Decision)

	// Find ADRs in main and in any PR that's labeled with "toc"
	branches := []string{branchRef(r.BaseBranch, true)}
	issues, _, err := r.gh.Search.Issues(ctx, fmt.Sprintf("repo:%s/%s is:pull-request is:open label:toc", r.Owner, r.Repo), nil)
	if err != nil {
		return nil, err
	}
	for _, issue := range issues.Issues {
		pr, _, err := r.gh.PullRequests.Get(ctx, r.Owner, r.Repo, issue.GetNumber())
		if err != nil {
			return nil, err
		}
		branches = append(branches, pr.GetHead().GetRef())
	}

	for _, branch := range branches {
		_, dir, _, err := r.gh.Repositories.GetContents(ctx, r.Owner, r.Repo, "/decisions",
			&github.RepositoryContentGetOptions{
				Ref: branch,
			},
		)
		if err != nil {
			return decisions, err
		}

		for _, decisionRef := range dir {
			content, err := r.getContent(ctx, branch, decisionRef.GetPath())
			if err != nil {
				return decisions, err
			}
			decision := &schema.Decision{
				RawContent: content,
				Path:       decisionRef.GetPath(),
				GithubURL:  decisionRef.GetHTMLURL(),
			}
			if _, ok := decisions[decision.Filename()]; !ok {
				decisions[decision.Filename()] = decision
			}
		}
	}

	return decisions, nil
}

func (r Builder) getContent(ctx context.Context, sha string, path string) (string, error) {
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

func (r Builder) createPullRequest(ctx context.Context) (*github.PullRequest, error) {
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

func (r Builder) commitToBranch(ctx context.Context, content string, message string, baseRef string) (*github.Reference, error) {
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

func (r Builder) generateReadme(decisions map[string]*schema.Decision) (string, error) {
	t := template.New("README.md")
	t = template.Must(t.Parse(readmeTemplate))

	var decisionSlice []*schema.Decision
	keys := make([]string, 0, len(decisions))
	for k := range decisions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		decisionSlice = append(decisionSlice, decisions[k])
	}

	buf := bytes.NewBufferString("")
	err := t.Execute(buf, TemplateData{
		Decisions:   decisionSlice,
		Name:        "Architecture",
		Description: "Our list of decisions",
	})

	return buf.String(), err
}

func (r Builder) regenerateReadme(ctx context.Context, decisions map[string]*schema.Decision) (string, error) {
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
