package runn

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/k1LoW/go-github-client/v53/factory"
)

type gh struct {
	client    *github.Client
	owner     string
	repo      string
	workspace string
	number    int
}

func newGitHubClient(ctx context.Context) (*gh, error) {
	if os.Getenv("GITHUB_ACTIONS") == "" {
		return nil, errors.New("env GITHUB_ACTIONS is not set. this environment is not on GitHub Actions runner")
	}
	splitted := strings.Split(os.Getenv("GITHUB_REPOSITORYR"), "/")
	if len(splitted) != 2 {
		return nil, errors.New("env GITHUB_REPOSITORY is invalid")
	}
	owner := splitted[0]
	repo := splitted[1]
	client, err := factory.NewGithubClient(factory.Timeout(10 * time.Second))
	if err != nil {
		return nil, err
	}
	ws := os.Getenv("GITHUB_WORKSPACE")
	if ws == "" {
		return nil, errors.New("env GITHUB_WORKSPACE is not set")
	}
	c := &gh{
		client:    client,
		owner:     owner,
		repo:      repo,
		workspace: ws,
	}
	n, err := c.detectCurrentPullRequestNumber(ctx)
	if err != nil {
		return nil, err
	}
	c.number = n

	return c, nil
}

func (g *gh) detectCurrentPullRequestNumber(ctx context.Context) (int, error) {
	if os.Getenv("GITHUB_PULL_REQUEST_NUMBER") != "" {
		return strconv.Atoi(os.Getenv("GITHUB_PULL_REQUEST_NUMBER"))
	}
	splitted := strings.Split(os.Getenv("GITHUB_REF"), "/") // refs/pull/8/head or refs/heads/branch/branch/name
	if len(splitted) < 3 {
		return 0, fmt.Errorf("env %s is not set", "GITHUB_REF")
	}
	if strings.Contains(os.Getenv("GITHUB_REF"), "refs/pull/") {
		prNumber := splitted[2]
		return strconv.Atoi(prNumber)
	}
	b := strings.Join(splitted[2:], "/")
	l, _, err := g.client.PullRequests.List(ctx, g.owner, g.repo, &github.PullRequestListOptions{
		State: "open",
	})
	if err != nil {
		return 0, err
	}
	var d *github.PullRequest
	for _, pr := range l {
		if pr.GetHead().GetRef() == b {
			if d != nil {
				return 0, errors.New("could not detect number of pull request")
			}
			d = pr
		}
	}
	if d != nil {
		return d.GetNumber(), nil
	}
	return 0, errors.New("could not detect number of pull request")
}

func (g *gh) createReviewComment(ctx context.Context, path string, start, end int, body string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(g.workspace, abs)
	if err != nil {
		return err
	}
	sig := fmt.Sprintf("<!-- runn: %s -->", rel)
	body = fmt.Sprintf("%s\n%s", body, sig)
	if _, _, err := g.client.PullRequests.CreateComment(ctx, g.owner, g.repo, g.number, &github.PullRequestComment{
		Path:      github.String(rel),
		StartLine: github.Int(start),
		Line:      github.Int(end),
		Body:      github.String(body),
	}); err != nil {
		return err
	}
	return nil
}
