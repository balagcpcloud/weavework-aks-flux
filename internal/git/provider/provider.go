package provider

import (
	"fmt"

	"github.com/go-logr/logr"
	giturl "github.com/kubescape/go-git-url"
	giturlapis "github.com/kubescape/go-git-url/apis"
	"golang.org/x/net/context"
)

type ProviderType string

const (
	ProviderGitHub    = ProviderType(giturlapis.ProviderGitHub)
	ProviderGitlab    = ProviderType(giturlapis.ProviderGitLab)
	ProviderBitbucket = ProviderType(giturlapis.ProviderBitBucket)
	ProviderAzure     = ProviderType(giturlapis.ProviderAzure)
)

type Provider interface {
	ListPullRequests(ctx context.Context, repo Repository) ([]PullRequest, error)
	AddCommentToPullRequest(ctx context.Context, repo PullRequest, body []byte) (*Comment, error)

	SetLogger(logr.Logger) error
	SetToken(tokenType, token string) error
	SetHostname(hostname string) error

	Setup() error
}

func New(provider ProviderType, options ...ProviderOption) (Provider, error) {
	var p Provider
	switch provider {
	case ProviderGitHub:
		p = newGitHubProvider()
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}

	for _, opt := range options {
		if err := opt(p); err != nil {
			return nil, err
		}
	}

	if err := p.Setup(); err != nil {
		return p, err
	}

	return p, nil
}

func FromURL(repoURL string, options ...ProviderOption) (Provider, Repository, error) {
	gitURL, err := giturl.NewGitURL(repoURL)
	if err != nil {
		return nil, Repository{}, fmt.Errorf("failed parsing repository url: %w", err)
	}

	targetProvider := ProviderType(gitURL.GetProvider())
	repo := Repository{
		Org:  gitURL.GetOwnerName(),
		Name: gitURL.GetRepoName(),
	}

	// Uncomment this when implementing Azure provider
	// if targetProvider == ProviderAzure {
	// 	repo.Project = gitURL.(*azureparserv1.AzureURL).GetProjectName()
	// }

	provider, err := New(targetProvider, options...)
	if err != nil {
		return nil, repo, err
	}

	return provider, repo, nil
}
