package sync

import (
	"os"
	"path/filepath"
	"testing"

)

func TestLoadConfig_JSON(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	content := `{
		"local_clone_dir": "/tmp/gitmirror",
		"credentials": {
			"github": {
				"ssh_key": "keys/github"
			}
		},
		"repositories": [
			{
				"name": "repo-1",
				"url": "git@github.com:owner/repo-1.git",
				"mirrors": [
					{
						"name": "gitlab",
						"url": "git@gitlab.com:owner/repo-1.git",
						"credential": "github"
					}
				]
			}
		]
	}`

	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write JSON config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig returned error for JSON: %v", err)
	}

	if config.LocalCloneDir != "/tmp/gitmirror" {
		t.Fatalf("unexpected local clone dir: got %q", config.LocalCloneDir)
	}

	if len(config.Repositories) != 1 {
		t.Fatalf("unexpected repository count: got %d", len(config.Repositories))
	}

	repo := config.Repositories[0]
	if repo.Name != "repo-1" {
		t.Fatalf("unexpected repository name: got %q", repo.Name)
	}

	if len(repo.Mirrors) != 1 {
		t.Fatalf("unexpected mirrors count: got %d", len(repo.Mirrors))
	}

	if repo.Mirrors[0].URL != "git@gitlab.com:owner/repo-1.git" {
		t.Fatalf("unexpected mirror URL: got %q", repo.Mirrors[0].URL)
	}
}

func TestLoadConfig_YAML(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	content := `local_clone_dir: /tmp/gitmirror
credentials:
  github:
    ssh_key: keys/github
repositories:
  - name: repo-1
    url: git@github.com:owner/repo-1.git
    mirrors:
      - name: gitlab
        url: git@gitlab.com:owner/repo-1.git
        credential: github
`

	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write YAML config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig returned error for YAML: %v", err)
	}

	if config.LocalCloneDir != "/tmp/gitmirror" {
		t.Fatalf("unexpected local clone dir: got %q", config.LocalCloneDir)
	}

	if len(config.Repositories) != 1 {
		t.Fatalf("unexpected repository count: got %d", len(config.Repositories))
	}

	repo := config.Repositories[0]
	if repo.Name != "repo-1" {
		t.Fatalf("unexpected repository name: got %q", repo.Name)
	}

	if len(repo.Mirrors) != 1 {
		t.Fatalf("unexpected mirrors count: got %d", len(repo.Mirrors))
	}

	if repo.Mirrors[0].Credential != "github" {
		t.Fatalf("unexpected mirror credential: got %q", repo.Mirrors[0].Credential)
	}
}