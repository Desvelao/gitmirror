package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/spf13/cobra"

	"gitmirror/internal/logger"
	"gitmirror/internal/providers"
	"gitmirror/internal/sync"
)

var discoverOriginUsername string
var discoverDestinationUsername string
var syncRepoIncludes []string
var syncRepoExcludes []string
var discoverOriginType string
var discoverDestinationType string
var summaryOutputPath string
var localCloneDir string
var origin_ssh_key string
var destination_ssh_key string
var customVars []string
var removeLocalClone bool
var cfgFile string

var syncCmd = &cobra.Command{
	Use:           "sync [origin_repo_url] [destination_repo_url] [--discover-username <username>] [--discover-provider <provider>] [--includes <repo1>] [--excludes <repo>] [--workdir <path>] [--origin-ssh-key <path>] [--destination-ssh-key <path>] [--summary <path>]",
	Short:         "Sync repositories from origin to destination, either specified directly or discovered from a provider",
	Long:          `Sync repositories using a local repository in 3 modes: origin-destination (source repository-destination repository), discover repositories from a provider, or use a configuration file. For discover or config modes, the repositories can be filtered using --includes and --excludes flags. The working directory for cloning and syncing can be set with --workdir. SSH keys for origin and destination can be specified with --origin-ssh-key and --destination-ssh-key. A summary of the sync operation can be output to a JSON file with --summary.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runSync,
}

func resolvePathFromExecutionDir(pathValue string, executionDir string) string {
	if pathValue == "" {
		return ""
	}
	if filepath.IsAbs(pathValue) {
		return pathValue
	}
	return filepath.Join(executionDir, pathValue)
}

func init() {
	syncCmd.Flags().StringVarP(&localCloneDir, "clone-dir", "w", "", "working directory to use for cloning and syncing repositories (overrides config)")
	syncCmd.Flags().StringVarP(&discoverOriginUsername, "discover-origin-username", "u", "", "provider username to discover repositories from (overrides config)")
	syncCmd.Flags().StringVarP(&discoverDestinationUsername, "discover-destination-username", "m", "", "provider username to discover repositories from (overrides config)")
	syncCmd.Flags().StringVarP(&discoverOriginType, "discover-origin", "p", "", "source platform to sync repositories from (overrides config)")
	syncCmd.Flags().StringVarP(&discoverDestinationType, "discover-destination", "q", "", "destination platform to sync repositories to (overrides config)")
	syncCmd.Flags().StringArrayVarP(&syncRepoIncludes, "includes", "i", []string{}, "include repositories to sync (can be specified multiple times)")
	syncCmd.Flags().StringArrayVarP(&syncRepoExcludes, "excludes", "e", []string{}, "exclude repositories from sync (can be specified multiple times)")
	syncCmd.Flags().StringArrayVarP(&customVars, "var", "v", []string{}, "define variables to use in config file (format: key=value, can be specified multiple times)")
	syncCmd.Flags().StringVarP(&origin_ssh_key, "origin-ssh-key", "k", "", "SSH key to use for the origin repository (overrides config)")
	syncCmd.Flags().StringVarP(&destination_ssh_key, "destination-ssh-key", "d", "", "SSH key to use for the destination repository (overrides config)")
	syncCmd.Flags().StringVarP(&summaryOutputPath, "summary", "s", "", "summary output file path (JSON)")
	syncCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "path to config file")
	syncCmd.Flags().BoolVar(&removeLocalClone, "cleanup", false, "remove the local clone of the repository after syncing (use with caution, this will delete the local copy of the repository after syncing, use only if you are sure you don't need it anymore)")

}

func runSync(cmd *cobra.Command, args []string) error {
	executionDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current execution directory: %w", err)
	}

	repos := []sync.RepoSync{}
	destinationMapByOrigin := map[string][]sync.RepoSync{}

	resolvedLocalCloneDir, _ := filepath.Abs(localCloneDir)
	// origin-destination mode
	origin := ""
	destination := ""

	if len(args) >= 1 {
		origin = args[0]
	}
	if len(args) >= 2 {
		destination = args[1]
	}

	customVarsMap := make(map[string]string)
	if len(customVars) > 0 {
		logger.Debug(fmt.Sprintf("Custom variables provided: %v", customVars))
		for _, v := range customVars {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid variable format: %s, expected key=value", v)
			}
			key := parts[0]
			value := parts[1]
			logger.Debug(fmt.Sprintf("Setting custom variable: %s=%s", key, value))
			customVarsMap[key] = value
		}
	}

	logger.Debug("Configuration")
	logger.Debug(fmt.Sprintf("localCloneDir: %s (%s)", localCloneDir, resolvedLocalCloneDir))
	logger.Debug(fmt.Sprintf("discover-origin-username: %s", discoverOriginUsername))
	logger.Debug(fmt.Sprintf("discover-origin-type: %s", discoverOriginType))
	logger.Debug(fmt.Sprintf("discover-destination-username: %s", discoverDestinationUsername))
	logger.Debug(fmt.Sprintf("discover-destination-type: %s", discoverDestinationType))
	logger.Debug(fmt.Sprintf("includes: %v", syncRepoIncludes))
	logger.Debug(fmt.Sprintf("excludes: %v", syncRepoExcludes))
	logger.Debug(fmt.Sprintf("summary: %s", summaryOutputPath))
	logger.Debug(fmt.Sprintf("origin: %s", origin))
	logger.Debug(fmt.Sprintf("destination: %s", destination))
	logger.Debug(fmt.Sprintf("origin_ssh_key: %s", origin_ssh_key))
	logger.Debug(fmt.Sprintf("destination_ssh_key: %s", destination_ssh_key))
	logger.Debug(fmt.Sprintf("customVars: %v", customVarsMap))
	logger.Debug(fmt.Sprintf("remove: %v", removeLocalClone))

	mode := ""

	summary := sync.SummarySync{
		DurationSec: 0,
		InputParameters: map[string]interface{}{
			"local_clone_dir":               resolvedLocalCloneDir,
			"discover_origin_username":      discoverOriginUsername,
			"discover_origin_type":          discoverOriginType,
			"discover_destination_username": discoverDestinationUsername,
			"discover_destination_type":     discoverDestinationType,
			"includes":                      syncRepoIncludes,
			"excludes":                      syncRepoExcludes,
			"origin":                        origin,
			"destination":                   destination,
			"origin_ssh_key":                origin_ssh_key,
			"destination_ssh_key":           destination_ssh_key,
			"custom_vars":                   customVarsMap,
			"remove_local_clone":            removeLocalClone,
		},
	}
	summary.Init()

	syncOptions := sync.SyncRepoOptions{
		LocalCloneDir:   resolvedLocalCloneDir,
		Vars:            customVarsMap,
		RemoveLocalRepo: removeLocalClone,
	}

	// Origin-destination mode: if both origin and destination are provided as positional args, we sync just that pair
	if origin != "" && destination != "" {
		mode = "origin-destination"
		syncOptions.RemoveLocalRepo = true
	} else if discoverOriginUsername != "" {
		mode = "discover"
	} else if cfgFile != "" {
		mode = "config"
	} else {
		return fmt.Errorf("no repositories specified for sync, provide origin and destination as positional arguments, or specify discover options, or provide a config file with repositories to sync")
	}

	summary.ExecutionMode = mode

	if mode == "discover" {

		if discoverOriginUsername == "" {
			return fmt.Errorf("Discover origin username is required")
		}

		if discoverOriginType == "" {
			return fmt.Errorf("Discover origin provider type is required")
		}

		if discoverDestinationType == "" {
			return fmt.Errorf("Discover destination provider type is required")
		}

		if discoverDestinationUsername == "" {
			return fmt.Errorf("Discover destination username is required")
		}

		provider, err := providers.GetProvider(discoverOriginType)

		if err != nil {
			return fmt.Errorf("error getting provider: %w", err)
		}

		_repos, err := provider.DiscoverRepositories(discoverOriginUsername)
		if err != nil {
			return fmt.Errorf("error discovering repositories: %w", err)
		}

		if len(_repos) == 0 {
			logger.Info(fmt.Sprintf("No repositories found for %s", discoverOriginUsername))
			return nil
		} else {
			logger.Info(fmt.Sprintf("Found [%d] repositories to sync of [%s]", len(_repos), discoverOriginUsername))
			for _, repo := range _repos {
				logger.Debug(fmt.Sprintf("Discovered repository: %s", repo.Name))
				repos = append(repos, sync.RepoSync{
					Name:       repo.Name,
					URL:        repo.URL,
					Credential: "",
					SSHKey:     resolvePathFromExecutionDir(origin_ssh_key, executionDir),
				})
			}

		}
	} else if mode == "config" {
		if cfgFile == "" {
			return fmt.Errorf("config file path is required for config mode")
		}

		cfg, err := sync.LoadConfig(cfgFile)
		syncOptions.Credentials = cfg.Credentials
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		} else {

		}
		repoNames := make(map[string]bool)
		for _, repo := range cfg.Repositories {
			if repoNames[repo.Name] {
				return fmt.Errorf("duplicate repository name in config: %s", repo.Name)
			}
			repoNames[repo.Name] = true
			repos = append(repos, sync.RepoSync{
				Name:       repo.Name,
				URL:        repo.URL,
				Credential: repo.Credential,
				SSHKey:     resolvePathFromExecutionDir(repo.SSHKey, executionDir),
			})
			for _, mirror := range repo.Mirrors {
				if destinationMapByOrigin[repo.Name] == nil {
					destinationMapByOrigin[repo.Name] = []sync.RepoSync{}
				}
				destinationMapByOrigin[repo.Name] = append(destinationMapByOrigin[repo.Name], sync.RepoSync{
					Name:       mirror.Name,
					URL:        mirror.URL,
					Credential: mirror.Credential,
					SSHKey:     resolvePathFromExecutionDir(mirror.SSHKey, executionDir),
				})
			}
		}
		logger.Info(fmt.Sprintf("Found [%d] repositories to sync in config", len(repos)))
	}

	if mode == "discover" || mode == "config" {
		// Build filter from positional args and --includes flag
		filterIncludes := make(map[string]bool)

		if len(syncRepoIncludes) > 0 {
			for _, name := range syncRepoIncludes {
				filterIncludes[name] = true
			}
			logger.Debug(fmt.Sprintf("Filtering repositories to include only: %v", syncRepoIncludes))
			filtered := repos[:0]
			for _, r := range repos {
				logger.Debug(fmt.Sprintf("Checking if respository is included in filter: %s", r.Name))
				if filterIncludes[r.Name] {
					logger.Debug(fmt.Sprintf("Including repository due to filter in sync: %s", r.Name))
					filtered = append(filtered, r)
				}
			}
			repos = filtered
		} else {
			logger.Debug("No repositories specified in --includes, all discovered repositories are included for sync")
		}

		if len(syncRepoExcludes) > 0 {
			for _, name := range syncRepoExcludes {
				filterIncludes[name] = false
			}
			logger.Debug(fmt.Sprintf("Filtering repositories to exclude: %v", syncRepoExcludes))
			filtered := repos[:0]
			for _, r := range repos {
				logger.Debug(fmt.Sprintf("Checking if respository is excluded in filter: %s", r.Name))
				if filterIncludes[r.Name] != false {
					logger.Debug(fmt.Sprintf("Excluding repository due to filter in sync: %s", r.Name))
					filtered = append(filtered, r)
				}
			}
			repos = filtered
		}
		if len(repos) == 0 {
			return fmt.Errorf("no repositories to sync after filtering, check your --includes and --excludes filters")
		}

		logger.Info(fmt.Sprintf("Found [%d] repositories to sync after filtering", len(repos)))
		logger.Debug(fmt.Sprintf("Repositories to sync: %v", func() []string {
			names := make([]string, len(repos))
			for i, r := range repos {
				names[i] = r.ToString()
			}
			return names
		}()))
	}
	if mode == "origin-destination" {
		logger.Debug("Both origin and destination repositories specified, syncing just this pair")
		logger.Debug(fmt.Sprintf("Origin repository specified: %s", origin))
		logger.Debug(fmt.Sprintf("Destination repository specified: %s", destination))
		syncRepo := sync.RepoSync{
			Name:   "origin",
			URL:    origin,
			SSHKey: resolvePathFromExecutionDir(origin_ssh_key, executionDir),
		}
		mirrors := []sync.RepoSync{
			{
				Name:   "destination",
				URL:    destination,
				SSHKey: resolvePathFromExecutionDir(destination_ssh_key, executionDir),
			},
		}
		if summ, err := sync.SyncRepo(syncRepo, mirrors, syncOptions); err != nil {
			logger.Error(fmt.Sprintf("Error syncing %s: %v", syncRepo.ToString(), err))
			summ.End()
			summary.Results = append(summary.Results, summ)
		} else {
			logger.Info(fmt.Sprintf("Synced: %s", syncRepo.ToString()))
			summ.End()
			summary.Results = append(summary.Results, summ)
		}
	} else if mode == "discover" {
		summary.DiscoveredRepos = func() []string {
			names := make([]string, len(repos))
			for i, r := range repos {
				names[i] = r.ToString()
			}
			return names
		}()
		logger.Info(fmt.Sprintf("Syncing discovered repositories from %s for %s to %s for %s", discoverOriginType, discoverOriginUsername, discoverDestinationType, discoverDestinationUsername))
		for _, syncRepo := range repos {
			summary.AddDiscoveredRepo(syncRepo)
			logger.Debug(fmt.Sprintf("Processing repository: %s", syncRepo.ToString()))
			logger.Debug(fmt.Sprintf("Resolving mirrors for %s", syncRepo.ToString()))

			provider, err := providers.GetProvider(discoverDestinationType)

			if err != nil {
				return fmt.Errorf("error getting provider: %w", err)
			}

			mirrorURL, err := provider.GetRepoURL(discoverDestinationUsername, syncRepo.Name)

			if err != nil {
				logger.Error(fmt.Sprintf("Error resolving mirror URL for %s: %v", syncRepo.ToString(), err))
				continue
			}

			mirrors := []sync.RepoSync{
				{
					Name:   "destination",
					URL:    mirrorURL,
					SSHKey: resolvePathFromExecutionDir(destination_ssh_key, executionDir),
				},
			}

			logger.Debug(fmt.Sprintf("Found mirrors for %s: %v", syncRepo.ToString(), mirrors))

			if summ, err := sync.SyncRepo(syncRepo, mirrors, syncOptions); err != nil {
				logger.Error(fmt.Sprintf("Error syncing %s: %v", syncRepo.ToString(), err))
				summ.End()
				summary.Results = append(summary.Results, summ)
			} else {
				logger.Info(fmt.Sprintf("Synced: %s", syncRepo.ToString()))
				summ.End()
				summary.Results = append(summary.Results, summ)
			}
		}
	} else if mode == "config" {
		for _, syncRepo := range repos {
			logger.Debug(fmt.Sprintf("Processing repository: %s", syncRepo.ToString()))
			logger.Debug(fmt.Sprintf("Resolving mirrors for %s", syncRepo.ToString()))

			mirrors := destinationMapByOrigin[syncRepo.Name]

			if err != nil {
				logger.Error(fmt.Sprintf("Error resolving mirrors for %s: %v", syncRepo.ToString(), err))
				continue
			}

			logger.Debug(fmt.Sprintf("Found mirrors for %s: %v", syncRepo.ToString(), mirrors))

			if len(mirrors) == 0 {
				logger.Warn(fmt.Sprintf("No mirrors configured for %s, skipping sync", syncRepo.ToString()))
				continue
			}
			if summ, err := sync.SyncRepo(syncRepo, mirrors, syncOptions); err != nil {
				logger.Error(fmt.Sprintf("Error syncing %s: %v", syncRepo.ToString(), err))
				summ.End()
				summary.Results = append(summary.Results, summ)
			} else {
				logger.Info(fmt.Sprintf("Synced: %s", syncRepo.ToString()))
				summ.End()
				summary.Results = append(summary.Results, summ)
			}
		}
	}

	summary.End()
	logger.Debug(fmt.Sprintf("Sync summary: %d repositories processed, duration: %.2f seconds %s", len(summary.Results), summary.DurationSec, summaryOutputPath))
	if summaryOutputPath != "" {
		summary.Write(summaryOutputPath)
	}

	var errorCount int
	for _, result := range summary.Results {
		errorCount += len(result.Errors)
	}

	logger.Debug(fmt.Sprintf("Sync completed with %d repositories, %d errors, duration: %.2f seconds", len(summary.Results), errorCount, summary.DurationSec))

	if errorCount > 0 {
		message := fmt.Sprintf("Sync completed with %d errors", errorCount)
		logger.Error(message)
		return fmt.Errorf(message)
	} else {
		logger.Info("Sync completed successfully without errors")
	}

	return nil
}
