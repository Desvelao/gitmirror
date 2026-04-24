package providers

import (
	"fmt"
	"strings"
)

// Repository represents a GitHub repository.
type Repository struct {
	Name   string `json:"name"`
	SSHURL string `json:"ssh_url"`
	URL    string `json:"clone_url"`
}

type GetRepoURLParams interface {
	GetOwner() string
	GetName() string
}

type Provider interface {
	DiscoverRepositories(username string) ([]Repository, error)
	GetRepoURL(owner string, name string) (string, error)
	CreateRepository(url string, token string) error
}

func GetProvider(providerType string) (Provider, error) {
	switch providerType {
	case "github":
		return githubProvider{}, nil
	case "gitlab":
		return gitlabProvider{}, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerType)
	}
}

func GetProviderByRepoURL(repoURL string) (Provider, error) {
	providerType := ""
	if strings.HasPrefix(repoURL, "https://github.com") || strings.HasPrefix(repoURL, "git@github.com") {
		providerType = "github"
	} else if strings.HasPrefix(repoURL, "https://gitlab.com") || strings.HasPrefix(repoURL, "git@gitlab.com") {
		providerType = "gitlab"
	}
	return GetProvider(providerType)
}
