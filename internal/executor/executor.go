package executor

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/omnorm/cd-gun/internal/config"
	"github.com/omnorm/cd-gun/internal/logger"
	"github.com/omnorm/cd-gun/internal/monitor"
	"github.com/omnorm/cd-gun/internal/state"
)

// Executor executes actions when changes are detected
type Executor struct {
	logger     *logger.Logger
	stateStore *state.Store //nolint:unused // Reserved for future use with action result persistence
	httpClient *http.Client
}

// ExecutionResult represents the result of executing an action
type ExecutionResult struct {
	RepositoryName string
	Success        bool
	Output         string
	Error          string
	Duration       time.Duration
	ExecutedAt     time.Time
}

// NewExecutor creates a new executor
func NewExecutor(log *logger.Logger) (*Executor, error) {
	return &Executor{
		logger:     log,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Execute executes an action based on a change event
func (e *Executor) Execute(action *config.Action, event *monitor.ChangeEvent,
	configMgr *config.Manager) (*ExecutionResult, error) {

	result := &ExecutionResult{
		RepositoryName: event.RepositoryName,
		ExecutedAt:     time.Now(),
	}

	startTime := time.Now()

	switch action.Type {
	case "shell":
		err := e.executeShell(action, event, configMgr)
		result.Duration = time.Since(startTime)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			e.logger.Errorf("Shell action failed for '%s': %v", event.RepositoryName, err)
		} else {
			result.Success = true
			e.logger.Infof("Shell action completed successfully for '%s'", event.RepositoryName)
		}

	case "webhook":
		err := e.executeWebhook(action, event)
		result.Duration = time.Since(startTime)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			e.logger.Errorf("Webhook action failed for '%s': %v", event.RepositoryName, err)
		} else {
			result.Success = true
			e.logger.Infof("Webhook action completed successfully for '%s'", event.RepositoryName)
		}

	default:
		return nil, fmt.Errorf("unknown action type: %s", action.Type)
	}

	return result, nil
}

// executeShell executes a shell script
func (e *Executor) executeShell(action *config.Action, event *monitor.ChangeEvent,
	configMgr *config.Manager) error {

	ctx, cancel := context.WithTimeout(context.Background(),
		configMgr.GetActionTimeout(action))
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", action.Script)

	// Set environment variables
	env := e.buildEnvironment(action, event, configMgr)
	cmd.Env = append(cmd.Env, env...)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	e.logger.Debugf("Executing shell action for '%s': %s", event.RepositoryName, action.Script)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("shell command failed: %w\nstdout: %s\nstderr: %s",
			err, stdout.String(), stderr.String())
	}

	if stderr.Len() > 0 {
		e.logger.Warnf("Shell action stderr for '%s': %s", event.RepositoryName, stderr.String())
	}

	return nil
}

// executeWebhook executes a webhook action
func (e *Executor) executeWebhook(action *config.Action, event *monitor.ChangeEvent) error {
	_ = map[string]interface{}{
		"repository": event.RepositoryName,
		"files":      event.Files,
		"old_hash":   event.OldHash,
		"new_hash":   event.NewHash,
		"timestamp":  event.DetectedAt,
	}

	// TODO: Marshal payload to JSON and send HTTP POST
	e.logger.Warnf("Webhook action for '%s' not yet implemented: %s",
		event.RepositoryName, action.URL)

	return nil
}

// buildEnvironment builds environment variables for shell execution
func (e *Executor) buildEnvironment(action *config.Action, event *monitor.ChangeEvent,
	configMgr *config.Manager) []string {

	cfg := configMgr.GetConfig()
	repo := findRepository(cfg, event.RepositoryName)

	env := []string{
		fmt.Sprintf("CDGUN_REPO_NAME=%s", event.RepositoryName),
		fmt.Sprintf("CDGUN_REPO_URL=%s", repo.URL),
		fmt.Sprintf("CDGUN_REPO_PATH=%s", configMgr.GetRepositoryLocalPath(event.RepositoryName)),
		fmt.Sprintf("CDGUN_BRANCH=%s", repo.Branch),
		fmt.Sprintf("CDGUN_CHANGED_FILES=%s", join(event.Files, ",")),
		fmt.Sprintf("CDGUN_OLD_HASH=%s", event.OldHash),
		fmt.Sprintf("CDGUN_NEW_HASH=%s", event.NewHash),
	}

	// Add custom environment variables from config
	for k, v := range action.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env
}

// findRepository finds a repository configuration by name
func findRepository(cfg *config.Config, name string) *config.Repository {
	for _, repo := range cfg.Repositories {
		if repo.Name == name {
			return &repo
		}
	}
	return nil
}

// join joins strings with a separator (simple implementation)
func join(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
