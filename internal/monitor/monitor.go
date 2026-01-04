package monitor

import (
	"fmt"
	"time"

	"github.com/omnorm/cd-gun/internal/config"
	"github.com/omnorm/cd-gun/internal/logger"
	"github.com/omnorm/cd-gun/internal/state"
)

// ChangeEvent represents a change detected in a repository
type ChangeEvent struct {
	RepositoryName string
	Files          []string
	OldHash        string
	NewHash        string
	DetectedAt     time.Time
}

// Monitor monitors a git repository for changes
type Monitor struct {
	repo       *config.Repository
	configMgr  *config.Manager
	logger     *logger.Logger
	stateStore *state.Store
	eventChan  chan ChangeEvent
	forceChan  chan struct{}
	ticker     *time.Ticker
}

// NewMonitor creates a new repository monitor
func NewMonitor(repo *config.Repository, configMgr *config.Manager,
	log *logger.Logger, stateStore *state.Store) (*Monitor, error) {

	return &Monitor{
		repo:       repo,
		configMgr:  configMgr,
		logger:     log,
		stateStore: stateStore,
		eventChan:  make(chan ChangeEvent, 1),
		forceChan:  make(chan struct{}, 1),
	}, nil
}

// Start begins monitoring the repository
func (m *Monitor) Start(stopChan chan struct{}) error {
	interval := m.configMgr.GetRepositoryPollInterval(m.repo)
	m.ticker = time.NewTicker(interval)
	defer m.ticker.Stop()

	m.logger.Infof("Starting monitor for repository '%s' (interval: %v)",
		m.repo.Name, interval)

	// Perform initial check
	if err := m.checkRepository(); err != nil {
		m.logger.Warnf("Initial check for '%s' failed: %v", m.repo.Name, err)
	}

	for {
		select {
		case <-stopChan:
			m.logger.Infof("Stopping monitor for repository '%s'", m.repo.Name)
			return nil

		case <-m.forceChan:
			m.logger.Infof("Force check triggered for '%s'", m.repo.Name)
			if err := m.checkRepository(); err != nil {
				m.logger.Errorf("Force check failed for '%s': %v", m.repo.Name, err)
			}

		case <-m.ticker.C:
			if err := m.checkRepository(); err != nil {
				m.logger.Warnf("Check failed for '%s': %v", m.repo.Name, err)
			}
		}
	}
}

// ForceCheck triggers an immediate check of the repository
func (m *Monitor) ForceCheck() {
	select {
	case m.forceChan <- struct{}{}:
	default:
		// Channel full, skip
	}
}

// GetEventChan returns the event channel for this monitor
func (m *Monitor) GetEventChan() <-chan ChangeEvent {
	return m.eventChan
}

// checkRepository checks for changes in the repository
func (m *Monitor) checkRepository() error {
	m.logger.Debugf("Checking repository '%s'", m.repo.Name)

	localPath := m.configMgr.GetRepositoryLocalPath(m.repo.Name)

	// Get git helper
	helper := NewGitHelper(localPath, m.logger)

	// Ensure repository is initialized
	if err := helper.EnsureRepository(m.repo); err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	// Fetch from remote
	if err := helper.Fetch(m.repo); err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	// Get current hash
	currentHash, err := helper.GetHash(m.repo.Branch)
	if err != nil {
		return fmt.Errorf("failed to get current hash: %w", err)
	}

	// Load previous state
	repoState, ok := m.stateStore.GetRepository(m.repo.Name)

	// Check if there's a change
	if !ok || repoState.CurrentHash != currentHash {
		// Check if any watched files changed
		var changedFiles []string

		if ok && repoState.CurrentHash != "" {
			files, err := helper.GetChangedFiles(repoState.CurrentHash, currentHash, m.repo.WatchPaths)
			if err != nil {
				m.logger.Warnf("Failed to get changed files for '%s': %v", m.repo.Name, err)
				changedFiles = m.repo.WatchPaths // Assume all watched paths changed
			} else {
				changedFiles = files
			}
		} else {
			changedFiles = m.repo.WatchPaths // No previous state, assume all paths changed
		}

		if len(changedFiles) > 0 {
			// Update state
			newState := state.RepositoryState{
				LastFetch:   time.Now(),
				CurrentHash: currentHash,
			}
			m.stateStore.UpdateRepository(m.repo.Name, newState)

			// Emit change event
			event := ChangeEvent{
				RepositoryName: m.repo.Name,
				Files:          changedFiles,
				OldHash:        repoState.CurrentHash,
				NewHash:        currentHash,
				DetectedAt:     time.Now(),
			}

			select {
			case m.eventChan <- event:
				m.logger.Infof("Change detected in '%s': %v", m.repo.Name, changedFiles)
			default:
				m.logger.Warnf("Event channel full for '%s'", m.repo.Name)
			}
		}
	}

	return nil
}
