package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Manager handles configuration loading and validation
type Manager struct {
	config      *Config
	configPath  string
	lastModTime time.Time
}

// NewManager creates a new config manager
func NewManager(configPath string) (*Manager, error) {
	m := &Manager{
		configPath: configPath,
	}

	if err := m.Load(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return m, nil
}

// Load reads and parses the configuration file, including any repository files
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if unmarshalErr := yaml.Unmarshal(data, &cfg); unmarshalErr != nil {
		return fmt.Errorf("failed to parse config: %w", unmarshalErr)
	}

	// Load repositories from include patterns
	for _, pattern := range cfg.IncludeRepositories {
		repos, reposErr := m.loadRepositoriesFromPattern(pattern)
		if reposErr != nil {
			return fmt.Errorf("failed to load repositories from pattern '%s': %w", pattern, reposErr)
		}
		cfg.Repositories = append(cfg.Repositories, repos...)
	}

	if validateErr := m.validate(&cfg); validateErr != nil {
		return fmt.Errorf("config validation failed: %w", validateErr)
	}

	if intervalsErr := m.parseIntervals(&cfg); intervalsErr != nil {
		return fmt.Errorf("failed to parse intervals: %w", intervalsErr)
	}

	m.config = &cfg

	// Update last modified time
	fi, err := os.Stat(m.configPath)
	if err == nil {
		m.lastModTime = fi.ModTime()
	}

	return nil
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// IsModified checks if the configuration file has been modified
func (m *Manager) IsModified() bool {
	fi, err := os.Stat(m.configPath)
	if err != nil {
		return false
	}
	return fi.ModTime().After(m.lastModTime)
}

// validate checks that the configuration is valid
func (m *Manager) validate(cfg *Config) error {
	if cfg.Agent.Name == "" {
		cfg.Agent.Name = "cd-gun-agent"
	}

	if cfg.Agent.LogLevel == "" {
		cfg.Agent.LogLevel = "info"
	}

	if cfg.Agent.StateDir == "" {
		cfg.Agent.StateDir = "/var/lib/cd-gun"
	}

	if cfg.Agent.CacheDir == "" {
		cfg.Agent.CacheDir = "/var/lib/cd-gun/repos"
	}

	if cfg.Agent.PollInterval == "" {
		cfg.Agent.PollInterval = "5m"
	}

	if len(cfg.Repositories) == 0 {
		return fmt.Errorf("at least one repository must be configured")
	}

	for i, repo := range cfg.Repositories {
		if repo.Name == "" {
			return fmt.Errorf("repository[%d]: name is required", i)
		}

		if repo.URL == "" {
			return fmt.Errorf("repository[%d]: url is required", i)
		}

		if repo.Branch == "" {
			cfg.Repositories[i].Branch = "main"
		}

		if repo.Auth.Type == "" {
			cfg.Repositories[i].Auth.Type = "none"
		}

		if len(repo.WatchPaths) == 0 {
			return fmt.Errorf("repository[%d]: watch_paths is required", i)
		}

		if repo.PollInterval == "" {
			cfg.Repositories[i].PollInterval = cfg.Agent.PollInterval
		}

		if repo.Action.Type == "" {
			return fmt.Errorf("repository[%d]: action.type is required", i)
		}

		if repo.Action.Type == "shell" && repo.Action.Script == "" {
			return fmt.Errorf("repository[%d]: action.script is required for shell action", i)
		}

		if repo.Action.Type == "webhook" && repo.Action.URL == "" {
			return fmt.Errorf("repository[%d]: action.url is required for webhook action", i)
		}

		if repo.Action.Timeout == "" {
			cfg.Repositories[i].Action.Timeout = "10m"
		}
	}

	return nil
}

// loadRepositoriesFromPattern loads repositories from a pattern (glob or directory)
// Handles both glob patterns (e.g., /etc/cd-gun/*.yaml) and direct file paths
func (m *Manager) loadRepositoriesFromPattern(pattern string) ([]Repository, error) {
	// Check if pattern contains wildcards
	if filepath.IsAbs(pattern) && !containsWildcards(pattern) {
		// Direct file path (no wildcards)
		info, err := os.Stat(pattern)
		if err != nil {
			return nil, fmt.Errorf("path not found: %w", err)
		}

		if info.IsDir() {
			// It's a directory, load all .yaml files from it
			return m.loadRepositoriesFromDir(pattern)
		}

		// It's a file, load repositories from it
		return m.loadRepositoriesFromFile(pattern)
	}

	// It's a glob pattern
	return m.loadRepositoriesFromGlob(pattern)
}

// containsWildcards checks if a path contains glob wildcards
func containsWildcards(path string) bool {
	return strings.Contains(path, "*") || strings.Contains(path, "?") || strings.Contains(path, "[")
}

// loadRepositoriesFromGlob loads repository configurations from files matching a glob pattern
func (m *Manager) loadRepositoriesFromGlob(pattern string) ([]Repository, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern: %w", err)
	}

	var repos []Repository
	for _, match := range matches {
		fileRepos, err := m.loadRepositoriesFromFile(match)
		if err != nil {
			return nil, fmt.Errorf("failed to load repositories from %s: %w", match, err)
		}
		repos = append(repos, fileRepos...)
	}

	return repos, nil
}

// loadRepositoriesFromDir loads all .yaml files from a directory
func (m *Manager) loadRepositoriesFromDir(dirPath string) ([]Repository, error) {
	pattern := filepath.Join(dirPath, "*.yaml")
	return m.loadRepositoriesFromGlob(pattern)
}

// loadRepositoriesFromFile loads repositories from a single file
func (m *Manager) loadRepositoriesFromFile(filePath string) ([]Repository, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Try to parse as array of repositories
	var repos []Repository
	if err := yaml.Unmarshal(data, &repos); err != nil {
		// If that fails, try to parse as a single repository wrapped in array
		var singleRepo Repository
		if err := yaml.Unmarshal(data, &singleRepo); err != nil {
			return nil, fmt.Errorf("failed to parse repositories: %w", err)
		}
		repos = []Repository{singleRepo}
	}

	return repos, nil
}

// parseIntervals parses all interval strings to time.Duration
func (m *Manager) parseIntervals(cfg *Config) error {
	// Parse agent poll interval
	d, err := time.ParseDuration(cfg.Agent.PollInterval)
	if err != nil {
		return fmt.Errorf("invalid agent.poll_interval: %w", err)
	}
	cfg.Agent.parsedInterval = d

	// Parse repository intervals
	for i, repo := range cfg.Repositories {
		d, err := time.ParseDuration(repo.PollInterval)
		if err != nil {
			return fmt.Errorf("invalid repositories[%d].poll_interval: %w", i, err)
		}
		cfg.Repositories[i].parsedInterval = d

		// Parse action timeout
		d, err = time.ParseDuration(repo.Action.Timeout)
		if err != nil {
			return fmt.Errorf("invalid repositories[%d].action.timeout: %w", i, err)
		}
		cfg.Repositories[i].Action.parsedTimeout = d
	}

	return nil
}

// GetRepositoryLocalPath returns the local cache path for a repository
func (m *Manager) GetRepositoryLocalPath(repoName string) string {
	return filepath.Join(m.config.Agent.CacheDir, repoName)
}

// GetPollInterval returns the parsed poll interval for agent
func (m *Manager) GetPollInterval() time.Duration {
	return m.config.Agent.parsedInterval
}

// GetRepositoryPollInterval returns the parsed poll interval for a repository
func (m *Manager) GetRepositoryPollInterval(repo *Repository) time.Duration {
	return repo.parsedInterval
}

// GetActionTimeout returns the parsed timeout for an action
func (m *Manager) GetActionTimeout(action *Action) time.Duration {
	return action.parsedTimeout
}

// ExpandEnv expands environment variables in a string
func ExpandEnv(s string) string {
	return os.ExpandEnv(s)
}
