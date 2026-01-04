package monitor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/omnorm/cd-gun/internal/config"
	"github.com/omnorm/cd-gun/internal/logger"
)

// GitHelper provides git operations
type GitHelper struct {
	repoPath string
	logger   *logger.Logger
}

// NewGitHelper creates a new git helper
func NewGitHelper(repoPath string, log *logger.Logger) *GitHelper {
	return &GitHelper{
		repoPath: repoPath,
		logger:   log,
	}
}

// EnsureRepository ensures the repository is initialized locally
func (g *GitHelper) EnsureRepository(repo *config.Repository) error {
	// Check if repo already exists
	if _, err := os.Stat(g.repoPath); err == nil {
		// Repository exists, verify it
		return g.verifyRepository(repo)
	}

	// Clone the repository
	return g.clone(repo)
}

// clone clones a repository
func (g *GitHelper) clone(repo *config.Repository) error {
	// Create parent directory
	parentDir := strings.TrimSuffix(g.repoPath, "/"+strings.Split(g.repoPath, "/")[len(strings.Split(g.repoPath, "/"))-1])
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	args := []string{"clone"}
	if repo.Branch != "" {
		args = append(args, "--branch", repo.Branch)
	}
	args = append(args, repo.URL, g.repoPath)

	cmd := exec.Command("git", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %w, output: %s", err, string(output))
	}

	g.logger.Debugf("Cloned repository '%s' to '%s'", repo.URL, g.repoPath)
	return nil
}

// verifyRepository verifies that the repository exists and is valid
func (g *GitHelper) verifyRepository(repo *config.Repository) error {
	// Check if it's a valid git repository
	cmd := exec.Command("git", "-C", g.repoPath, "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("invalid git repository at %s: %w", g.repoPath, err)
	}

	return nil
}

// Fetch fetches from remote repository
func (g *GitHelper) Fetch(repo *config.Repository) error {
	cmd := exec.Command("git", "-C", g.repoPath, "fetch", "origin", repo.Branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch failed: %w, output: %s", err, string(output))
	}

	return nil
}

// GetHash gets the commit hash for a branch
func (g *GitHelper) GetHash(branch string) (string, error) {
	cmd := exec.Command("git", "-C", g.repoPath, "rev-parse", fmt.Sprintf("origin/%s", branch))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get hash for branch '%s': %w", branch, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetChangedFiles returns files that changed between two commits
func (g *GitHelper) GetChangedFiles(oldHash, newHash string, watchPaths []string) ([]string, error) {
	if oldHash == "" {
		// No previous state, assume all watched files
		return watchPaths, nil
	}

	// Get list of changed files
	cmd := exec.Command("git", "-C", g.repoPath, "diff", "--name-only",
		fmt.Sprintf("%s..%s", oldHash, newHash))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get diff: %w", err)
	}

	changedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")

	// Filter by watch paths
	var filtered []string
	for _, changed := range changedFiles {
		for _, watch := range watchPaths {
			if matchesPath(changed, watch) {
				filtered = append(filtered, changed)
				break
			}
		}
	}

	return filtered, nil
}

// matchesPath checks if a file path matches a watch pattern
func matchesPath(filePath, pattern string) bool {
	// Remove trailing slash from pattern
	pattern = strings.TrimSuffix(pattern, "/")

	// Exact match
	if filePath == pattern {
		return true
	}

	// Pattern is a directory
	if strings.HasPrefix(filePath, pattern+"/") {
		return true
	}

	// Simple wildcard support
	if strings.HasSuffix(pattern, "/*") {
		dir := strings.TrimSuffix(pattern, "/*")
		if strings.HasPrefix(filePath, dir+"/") {
			return true
		}
	}

	return false
}
