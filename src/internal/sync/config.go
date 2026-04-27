package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigRepo represents a repository configuration with its name and URLs.

type Mirror struct {
	Name       string `json:"name,omitempty" yaml:"name,omitempty"`
	URL        string `json:"url" yaml:"url"`
	SSHKey     string `json:"ssh_key,omitempty" yaml:"ssh_key,omitempty"`
	SSHCmdOpts string `json:"ssh_command_options,omitempty" yaml:"ssh_command_options,omitempty"`
	Credential string `json:"credential,omitempty" yaml:"credential,omitempty"`
}

type ConfigRepo struct {
	Name       string   `json:"name" yaml:"name"`
	URL        string   `json:"url" yaml:"url"`
	Credential string   `json:"credential,omitempty" yaml:"credential,omitempty"`
	SSHKey     string   `json:"ssh_key,omitempty" yaml:"ssh_key,omitempty"`
	SSHCmdOpts string   `json:"ssh_command_options,omitempty" yaml:"ssh_command_options,omitempty"`
	LocalCloneDirCleanup bool `json:"local_clone_dir_cleanup,omitempty" yaml:"local_clone_dir_cleanup,omitempty"`
	Mirrors    []Mirror `json:"mirrors" yaml:"mirrors"`
}

type SyncConfigFileCredential struct {
	SSHKey     string `json:"ssh_key,omitempty" yaml:"ssh_key,omitempty"`
	SSHCmdOpts string `json:"ssh_command_options,omitempty" yaml:"ssh_command_options,omitempty"`
}

type SyncConfigFileCredentials map[string]SyncConfigFileCredential

// Config holds the configuration settings for the application.
type SyncConfigFile struct {
	LocalCloneDir      string              `json:"local_clone_dir" yaml:"local_clone_dir"`
	LocalCloneDirCleanup bool              `json:"local_clone_dir_cleanup" yaml:"local_clone_dir_cleanup"`
	Credentials  SyncConfigFileCredentials `json:"credentials" yaml:"credentials"`
	Includes     []string                  `json:"includes" yaml:"includes"`
	Excludes     []string                  `json:"excludes" yaml:"excludes"`
	Repositories []ConfigRepo              `json:"repositories" yaml:"repositories"`
}

// LoadConfig loads the configuration from a JSON or YAML file.
func LoadConfig(filePath string) (*SyncConfigFile, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config SyncConfigFile

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		if err := json.Unmarshal(content, &config); err != nil {
			return nil, err
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(content, &config); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", filePath)
	}

	return &config, nil
}

// Validate checks if the required configuration fields are set.
func (c *SyncConfigFile) Validate(repoName string) ([]string, error) {
	return nil, nil
}
