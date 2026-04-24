package providers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gitmirror/internal/logger"
)

type githubProvider struct{}

func (githubProvider) DiscoverRepositories(username string) ([]Repository, error) {
	logger.Info(fmt.Sprintf("Discovering repositories for GitHub user: %s", username))
	url := fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=100&type=all", username)
	logger.Debug(fmt.Sprintf("GitHub API URL: %s", url))
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch repositories: received status code %d", resp.StatusCode)
	}

	var repos []Repository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	logger.Debug(fmt.Sprintf("Discovered %d repositories for GitHub user: %s", len(repos), username))
	return repos, nil
}

func (githubProvider) GetRepoURL(owner string, name string) (string, error) {
	return fmt.Sprintf("git@github.com:%s/%s.git", owner, name), nil
}

func (githubProvider) CreateRepository(url string, token string) error {
	// TODO: Implement GitHub repository creation using GitHub API
	return nil
}
