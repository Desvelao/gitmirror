package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"gitmirror/internal/logger"
)

type gitlabProvider struct{}

func (gitlabProvider) DiscoverRepositories(username string) ([]Repository, error) {
	// TODO: Implement GitLab repository discovery using GitLab API
	return nil, nil
}

func (gitlabProvider) GetRepoURL(owner string, name string) (string, error) {
	return fmt.Sprintf("git@gitlab.com:%s/%s.git", owner, name), nil
}

func (gitlabProvider) CreateRepository(repoURL string, token string) error {
	repoName := getRepoNameFromURL(repoURL)
	repoNamespaceID := getNamespaceIDFromURL(repoURL)
	logger.Debug(fmt.Sprintf("Creating GitLab repository: Name=%s, NamespaceID=%s", repoName, repoNamespaceID))

	logger.Debug(fmt.Sprintf("Checking if the repository already exists: Name=%s, NamespaceID=%s", repoName, repoNamespaceID))
	client := NewClient(token)
	projectPath := fmt.Sprintf("%s/%s", repoNamespaceID, repoName)
	encodedProjectPath := url.PathEscape(projectPath)
	encodedProjectPath = strings.ReplaceAll(encodedProjectPath, "-", "%2D")
	resp, err := client.doRequest(http.MethodGet, fmt.Sprintf("projects/%s", encodedProjectPath), nil)
	if err != nil {
		return fmt.Errorf("gitlab request failed: %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		logger.Info(fmt.Sprintf("Repository %s already exists in namespace %s", repoName, repoNamespaceID))
	} else if resp.StatusCode == http.StatusNotFound {
		logger.Info(fmt.Sprintf("Repository %s does not exist in namespace %s, creating it", repoName, repoNamespaceID))
		if createErr := client.CreateProject(repoName, repoNamespaceID); createErr != nil {
			return fmt.Errorf("failed to create repository %s: %w", repoURL, createErr)
		}
	} else {
		return fmt.Errorf("failed to create repository: received status code %d", resp.StatusCode)
	}

	return nil
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

func NewClient(token string) *Client {
	return &Client{
		BaseURL:    "https://gitlab.com/api/v4",
		HTTPClient: &http.Client{},
		Token:      token,
	}
}

type ClientResponse struct {
	StatusCode int
	Body       []byte
}

func (c *Client) doRequest(method, endpoint string, body interface{}) (*ClientResponse, error) {
	url := fmt.Sprintf("%s/%s", c.BaseURL, endpoint)

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}
	logger.Debug(fmt.Sprintf("Making request to GitLab API: %s %s, Body=%s", method, url, buf.String()))
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("PRIVATE-TOKEN", c.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Making request to GitLab API: %s %s request failed: %w", method, url, err)
	}

	defer resp.Body.Close()

	responseBody, readErr := io.ReadAll(resp.Body)

	if readErr != nil {
		return nil, fmt.Errorf("Making request to GitLab API: %s %s response body: %w", method, url, readErr)
	}

	bodyStr := string(responseBody)
	logger.Debug(fmt.Sprintf("Making request to GitLab API: %s %s response body: %s", method, url, bodyStr))
	logger.Debug(fmt.Sprintf("Making request to GitLab API: %s %s response status: %d", method, url, resp.StatusCode))

	return &ClientResponse{
		StatusCode: resp.StatusCode,
		Body:       responseBody,
	}, nil
}

func (c *Client) CreateProject(name, namespaceID string) error {
	// TODO: implement this method to create a project in GitLab using the API
	logger.Debug(fmt.Sprintf("Creating GitLab repository: Name=%s, NamespaceID=%s", name, namespaceID))

	resp, err := c.doRequest(http.MethodPost, "projects", map[string]interface{}{
		"name":         name,
		"namespace_id": namespaceID,
		"visibility":   "private",
	})
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create repository: received status code %d", resp.StatusCode)
	}

	// body := map[string]interface{}{
	// 	"name":         name,
	// 	"namespace_id": namespaceID,
	// 	"visibility":   "private",
	// }

	// resp, err := c.doRequest(http.MethodPost, "projects", body)
	// if err != nil {
	// 	return err
	// }

	// if resp.StatusCode != http.StatusCreated {
	// 	return fmt.Errorf("failed to create project: %s", resp.Status)
	// }

	return nil
}

func (c *Client) GetProject(projectID string) (map[string]interface{}, error) {
	// resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("projects/%s", projectID), nil)
	// if err != nil {
	// 	return nil, err
	// }

	// if resp.StatusCode != http.StatusOK {
	// 	return nil, fmt.Errorf("failed to get project: %s", resp.Status)
	// }

	// var project map[string]interface{}
	// if err := json.Unmarshal(resp.Body, &project); err != nil {
	// 	return nil, err
	// }

	// return project, nil
	return nil, nil
}

func getRepoNameFromURL(repoURL string) string {
	// This is a simple implementation and may not cover all cases
	parts := strings.Split(repoURL, "/")
	nameWithGit := parts[len(parts)-1]
	return strings.TrimSuffix(nameWithGit, ".git")
}

func getNamespaceIDFromURL(repoURL string) string {
	// This is a simple implementation and may not cover all cases
	parts := strings.Split(repoURL, ":")
	reportoryPath := parts[len(parts)-1]
	parts = strings.Split(reportoryPath, "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-2]
}
