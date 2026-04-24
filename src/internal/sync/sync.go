package sync

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gitmirror/internal/logger"
	"gitmirror/internal/providers"
)

type RepoSync struct {
	Name       string `json:"name,omitempty"`
	URL        string `json:"url"`
	Credential string `json:"credential,omitempty"`
	SSHKey     string `json:"ssh_key,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
}

func (r RepoSync) GetName() string {
	return r.Name
}

func (r RepoSync) GetURL() string {
	return r.URL
}

func (r RepoSync) GetSSHKey() string {
	return r.SSHKey
}

func (r RepoSync) ToString() string {
	return fmt.Sprintf("%s|%s", r.Name, r.URL)
}

func (r RepoSync) GetProtocol() string {
	if r.URL[:4] == "git@" || r.URL[:4] == "ssh:" {
		return "ssh"
	} else if r.URL[:5] == "https" {
		return "https"
	}
	return "unknown"
}

func (r RepoSync) RequiresSSHKey() bool {
	return r.GetProtocol() == "ssh"
}

func (r RepoSync) GetProvider() string {
	return r.Protocol
	// TODO: implement provider detection based on URL pattern
}

type RepoConfigureGitCommand interface {
	GetName() string
	GetURL() string
	GetSSHKey() string
}

type SummarySyncRepoResult struct {
	Timestamp         string   `json:"timestamp"`
	DurationSec       float64  `json:"duration_sec"`
	Name              string   `json:"name"`
	URL               string   `json:"url"`
	Path              string   `json:"path,omitempty"`
	Cloned            bool     `json:"cloned"`
	Fetched           bool     `json:"fetched"`
	MirrorPushSuccess []string `json:"mirror_push_success"`
	MirrorPushFailed  []string `json:"mirror_push_failed"`
	Errors            []string `json:"errors"`
}

type SummarySync struct {
	Timestamp       string                  `json:"timestamp"`
	DurationSec     float64                 `json:"duration_sec"`
	InputParameters map[string]interface{}  `json:"input_parameters"`
	ExecutionMode   string                  `json:"execution_mode"`
	DiscoveredRepos []string                `json:"discovered_repos"`
	Results         []SummarySyncRepoResult `json:"results"`
}

func (s *SummarySync) Init() {
	s.Timestamp = time.Now().UTC().Format(time.RFC3339)
}

func (s *SummarySync) End() {
	startTime, err := time.Parse(time.RFC3339, s.Timestamp)
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing start time: %v", err))
		s.DurationSec = 0
		return
	}
	s.DurationSec = math.Round(time.Since(startTime).Seconds()*100) / 100
}

func (s *SummarySync) Write(filePath string) error {
	logger.Debug(fmt.Sprintf("Writing summary to file: %s", filePath))
	file, err := os.Create(filePath)
	if err != nil {
		logger.Error(fmt.Sprintf("Error creating summary file: %v", err))
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(s); err != nil {
		logger.Error(fmt.Sprintf("Error writing summary to file: %v", err))
		return err
	}
	return nil
}

func (s *SummarySync) AddDiscoveredRepo(repo RepoSync) {
	s.DiscoveredRepos = append(s.DiscoveredRepos, repo.ToString())
}

func (s *SummarySyncRepoResult) Init() {
	s.Timestamp = time.Now().UTC().Format(time.RFC3339)
	s.Cloned = false
	s.Fetched = false
	s.MirrorPushSuccess = []string{}
	s.MirrorPushFailed = []string{}
	s.Errors = []string{}
}

func (s *SummarySyncRepoResult) End() {
	startTime, err := time.Parse(time.RFC3339, s.Timestamp)
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing start time: %v", err))
		s.DurationSec = 0
		return
	}
	s.DurationSec = math.Round(time.Since(startTime).Seconds()*100) / 100
}

type SyncRepoOptions struct {
	LocalCloneDir   string
	Credentials     SyncConfigFileCredentials
	RemoveLocalRepo bool
	Vars            map[string]string
}

// SyncRepo synchronizes a repository between providers.
func SyncRepo(repo RepoSync, mirrors []RepoSync, options SyncRepoOptions) (SummarySyncRepoResult, error) {
	summary := SummarySyncRepoResult{
		Name: repo.Name,
		URL:  repo.URL,
	}
	summary.Init()

	// Always clean up when this function returns
	defer func() {
		// _ = os.RemoveAll(cloneDir)
		logger.Debug(fmt.Sprintf("Cleaning up local repository for %s", repo.ToString()))

		if options.RemoveLocalRepo {
			repoDir := filepath.Join(options.LocalCloneDir, repo.Name+".git")
			logger.Debug(fmt.Sprintf("Removing local repository directory %s for %s", repoDir, repo.ToString()))
			if err := os.RemoveAll(repoDir); err != nil {
				logger.Error(fmt.Sprintf("Error removing local repository directory %s: %v", repoDir, err))
			} else {
				logger.Info(fmt.Sprintf("Removed local repository directory %s for %s", repoDir, repo.ToString()))
			}
		} else {
			logger.Debug(fmt.Sprintf("Skipping removal of local repository for %s due to configuration", repo.ToString()))
		}
	}()

	logger.Info(fmt.Sprintf("Syncing %s", repo.ToString()))
	logger.Debug(fmt.Sprintf("Sync options: LocalCloneDir=%s, Vars=%v", options.LocalCloneDir, options.Vars))

	resolvedRepo, err := mergeRepoSyncCredentials(repo, options.Credentials)
	if err != nil {
		summary.Errors = append(summary.Errors, fmt.Sprintf("Error preparing repository %s: %v", repo.ToString(), err))
		return summary, err
	}

	// Clone the repository if it doesn't exist locally
	if err := cloneRepo(resolvedRepo, options); err != nil {
		summary.Errors = append(summary.Errors, fmt.Sprintf("Error cloning repository %s: %v", resolvedRepo.ToString(), err))
		return summary, err
	}
	summary.Cloned = true

	repoPath, err := filepath.Abs(filepath.Join(options.LocalCloneDir, resolvedRepo.Name+".git"))
	if err != nil {
		summary.Errors = append(summary.Errors, fmt.Sprintf("Error resolving repository path for %s: %v", resolvedRepo.ToString(), err))
		return summary, err
	}
	summary.Path = repoPath

	if err := validateLocalRepo(resolvedRepo, options); err != nil {
		summary.Errors = append(summary.Errors, fmt.Sprintf("Error validating local repository %s: %v", resolvedRepo.ToString(), err))
		return summary, err
	}

	// Fetch the latest changes
	if err := fetchRepo(resolvedRepo, options); err != nil {
		summary.Errors = append(summary.Errors, fmt.Sprintf("Error fetching repository %s: %v", resolvedRepo.ToString(), err))
		return summary, err
	}
	summary.Fetched = true

	logger.Debug(fmt.Sprintf("Finished preparing repository %s for sync, starting to push to %d mirrors", resolvedRepo.ToString(), len(mirrors)))

	// Push to each mirror
	for _, mirror := range mirrors {
		resolvedMirror, err := mergeRepoSyncCredentials(mirror, options.Credentials)
		if err != nil {
			message := fmt.Sprintf("Error preparing mirror %s for %s: %v", resolvedMirror.ToString(), resolvedRepo.ToString(), err)
			logger.Error(message)
			summary.Errors = append(summary.Errors, message)
			continue
		}

		// Check if the mirror is created
		// TODO: check if the mirror repository exists before trying to push, and create it if it doesn't exist. This is needed for providers like GitLab where the repository must be created before pushing.
		// if err := createRepo(resolvedMirror, options); err != nil {
		// 	message := fmt.Sprintf("Error creating mirror %s for %s: %v", resolvedMirror.ToString(), resolvedRepo.ToString(), err)
		// 	logger.Error(message)
		// 	summary.Errors = append(summary.Errors, message)
		// 	summary.MirrorPushFailed = append(summary.MirrorPushFailed, resolvedMirror.ToString())
		// 	continue
		// }

		if err := pushToMirror(resolvedRepo, resolvedMirror, options); err != nil {
			message := fmt.Sprintf("Error pushing %s to mirror %s: %v", resolvedRepo.ToString(), resolvedMirror.ToString(), err)
			logger.Error(message)
			summary.Errors = append(summary.Errors, message)
			summary.MirrorPushFailed = append(summary.MirrorPushFailed, resolvedMirror.ToString())
			return summary, err
		} else {
			message := fmt.Sprintf("Successfully pushed %s to mirror %s", resolvedRepo.ToString(), resolvedMirror.ToString())
			logger.Info(message)
			summary.MirrorPushSuccess = append(summary.MirrorPushSuccess, resolvedMirror.ToString())
		}
	}

	return summary, nil
}

func mergeRepoSyncCredentials(mirror RepoSync, credentials SyncConfigFileCredentials) (RepoSync, error) {
	resolvedMirror := mirror

	if resolvedMirror.Credential != "" {
		logger.Debug(fmt.Sprintf("Resolving credential %s for reposync %s", resolvedMirror.Credential, mirror.URL))
		credential, ok := credentials[resolvedMirror.Credential]
		if !ok {
			return resolvedMirror, fmt.Errorf("credential %q is not defined", resolvedMirror.Credential)
		}

		if resolvedMirror.SSHKey == "" {
			logger.Debug(fmt.Sprintf("Resolving SSH key for reposync %s using credential %s", mirror.URL, resolvedMirror.Credential))
			resolvedMirror.SSHKey = credential.SSHKey
		}
	}

	return resolvedMirror, nil
}

// cloneRepo clones the repository from the origin.
func cloneRepo(repo RepoSync, options SyncRepoOptions) error {
	repoDir := repo.Name + ".git"
	repoDirPath := filepath.Join(options.LocalCloneDir, repoDir)
	logger.Debug(fmt.Sprintf("Checking if repository %s already exists at %s", repo.ToString(), repoDirPath))
	if _, err := os.Stat(repoDirPath); err == nil {
		logger.Info(fmt.Sprintf("Repository %s already cloned, skipping clone", repo.ToString()))
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Failed to inspect repository directory %s: %w", repoDirPath, err)
	}

	logger.Debug(fmt.Sprintf("Cloning %s into %s", repo.ToString(), repoDirPath))
	cmd := exec.Command("git", "clone", "--mirror", repo.URL, repoDir)
	cmd.Dir = options.LocalCloneDir
	configureGitCommand(cmd, repo)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone --mirror %s %s failed: %w\n%s", repo.URL, repoDir, err, string(output))
	}
	logger.Info(fmt.Sprintf("Cloned %s into %s", repo.ToString(), repoDirPath))
	return nil
}

func validateLocalRepo(repo RepoSync, options SyncRepoOptions) error {
	repoDir := filepath.Join(options.LocalCloneDir, repo.Name+".git")
	logger.Debug(fmt.Sprintf("Validating local repository at %s for %s", repoDir, repo.ToString()))

	logger.Debug(fmt.Sprintf("Checking if repository directory %s is a bare repository", repoDir))
	cmd := exec.Command("git", "rev-parse", "--is-bare-repository")
	cmd.Dir = repoDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git rev-parse failed for %s: %w\n%s", repo.ToString(), err, string(output))
	}
	if string(output) != "true\n" {
		return fmt.Errorf("local repository at %s is not a bare repository", repoDir)
	}
	logger.Debug(fmt.Sprintf("Local repository at %s is a bare repository for %s", repoDir, repo.ToString()))

	logger.Debug(fmt.Sprintf("Checking if repository directory %s has correct origin URL for %s", repoDir, repo.ToString()))
	cmd = exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoDir
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git remote get-url failed for %s: %w\n%s", repo.ToString(), err, string(output))
	}
	originURL := string(output)
	if originURL != repo.URL+"\n" {
		return fmt.Errorf("local repository at %s has origin URL %s, expected %s", repoDir, originURL, repo.URL)
	}
	logger.Debug(fmt.Sprintf("Local repository at %s has correct origin URL for %s", repoDir, repo.ToString()))

	logger.Debug(fmt.Sprintf("Validated local repository at %s for %s", repoDir, repo.ToString()))
	return nil
}

// fetchRepo fetches the latest changes from the origin.
func fetchRepo(repo RepoSync, options SyncRepoOptions) error {
	logger.Debug(fmt.Sprintf("Fetching latest changes for %s", repo.ToString()))
	cmd := exec.Command("git", "fetch", "origin", "--prune", "--tags")
	cmd.Dir = filepath.Join(options.LocalCloneDir, repo.Name+".git")
	configureGitCommand(cmd, repo)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch failed for %s: %w\n%s", repo.ToString(), err, string(output))
	}
	logger.Info(fmt.Sprintf("Fetched latest changes for %s", repo.ToString()))
	return nil
}

// pushToMirror pushes the repository to the specified mirror.
func pushToMirror(repo RepoSync, mirror RepoSync, options SyncRepoOptions) error {
	logger.Debug(fmt.Sprintf("Pushing %s to mirror %s", repo.ToString(), mirror.ToString()))
	cmd := exec.Command("git", "push", "--mirror", mirror.URL)
	cmd.Dir = filepath.Join(options.LocalCloneDir, repo.Name+".git")

	configureGitCommand(cmd, mirror)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed for %s to %s: %w\n%s", repo.ToString(), mirror.ToString(), err, string(output))
	}
	logger.Info(fmt.Sprintf("Pushed %s to mirror %s", repo.ToString(), mirror.ToString()))
	logger.Info(fmt.Sprintf("git push output: %s", string(output)))
	return nil
}

func createRepo(repo RepoSync, options SyncRepoOptions) error {
	provider, err := providers.GetProviderByRepoURL(repo.URL)
	if err != nil {
		return fmt.Errorf("unsupported repository URL: %s", repo.URL)
	}

	logger.Debug(fmt.Sprintf("VARS %v", options.Vars))
	if err := provider.CreateRepository(repo.URL, options.Vars["GITLAB_TOKEN"]); err != nil {
		return fmt.Errorf("failed to create repository %s: %w", repo.URL, err)
	}

	return nil
}

func configureGitCommand(cmd *exec.Cmd, repo RepoConfigureGitCommand) {
	if repo.GetURL()[:4] == "ssh:" || repo.GetURL()[:4] == "git@" {
		if repo.GetSSHKey() != "" {
			logger.Debug(fmt.Sprintf("Configuring GIT_SSH_COMMAND for repository %s to use SSH key %s", repo.GetURL(), repo.GetSSHKey()))
			cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_SSH_COMMAND=ssh -i '%s' -o IdentitiesOnly=yes", repo.GetSSHKey()))
		} else {
			logger.Error(fmt.Sprintf("SSH mirror %s requires ssh_key in config", repo.GetURL()))
		}
	}
}

func getRepoNameFromURL(repoURL string) string {
	// This is a simple implementation and may not cover all cases
	parts := strings.Split(repoURL, "/")
	nameWithGit := parts[len(parts)-1]
	return strings.TrimSuffix(nameWithGit, ".git")
}
