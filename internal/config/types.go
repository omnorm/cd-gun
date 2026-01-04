package config

import "time"

// Config is the main configuration structure for CD-Gun
type Config struct {
	Agent               AgentConfig  `yaml:"agent"`
	Repositories        []Repository `yaml:"repositories"`
	IncludeRepositories []string     `yaml:"include_repositories"` // List of glob patterns or file paths for repository configs
}

// AgentConfig contains agent-specific settings
type AgentConfig struct {
	Name           string        `yaml:"name"`
	LogLevel       string        `yaml:"log_level"`
	LogFile        string        `yaml:"log_file"` // Optional: path to log file (if not set, logs to stdout)
	StateDir       string        `yaml:"state_dir"`
	CacheDir       string        `yaml:"cache_dir"`
	PollInterval   string        `yaml:"poll_interval"`
	parsedInterval time.Duration `yaml:"-"`
}

// Repository represents a git repository to monitor
type Repository struct {
	Name           string        `yaml:"name"`
	URL            string        `yaml:"url"`
	Branch         string        `yaml:"branch"`
	Auth           Auth          `yaml:"auth"`
	WatchPaths     []string      `yaml:"watch_paths"`
	PollInterval   string        `yaml:"poll_interval"`
	parsedInterval time.Duration `yaml:"-"`
	Action         Action        `yaml:"action"`
}

// Auth contains authentication configuration for a repository
type Auth struct {
	Type        string `yaml:"type"`        // ssh, https, none
	Credentials string `yaml:"credentials"` // path to credentials file or token
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
}

// Action describes what to do when files change
type Action struct {
	Type          string            `yaml:"type"` // shell, webhook, custom
	Script        string            `yaml:"script"`
	URL           string            `yaml:"url"`
	Handler       string            `yaml:"handler"`
	Timeout       string            `yaml:"timeout"`
	parsedTimeout time.Duration     `yaml:"-"`
	Parallel      bool              `yaml:"parallel"`
	Env           map[string]string `yaml:"env"`
}
