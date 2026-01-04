package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Store manages the persisted state of cd-gun
type Store struct {
	mu        sync.RWMutex
	state     *State
	filePath  string
	autoSave  bool
	saveTimer *time.Timer
}

// NewStore creates a new state store
func NewStore(stateDir string) (*Store, error) {
	statePath := filepath.Join(stateDir, "state.json")

	// Ensure state directory exists
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	store := &Store{
		filePath: statePath,
		autoSave: true,
	}

	// Try to load existing state
	if err := store.Load(); err != nil {
		// If file doesn't exist, create new state
		store.state = NewState()
	}

	return store, nil
}

// Load reads the state from disk
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	s.state = &state
	return nil
}

// Save writes the state to disk
func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// SaveAsync saves the state to disk asynchronously
func (s *Store) SaveAsync() {
	if !s.autoSave {
		return
	}

	// Cancel any pending save
	if s.saveTimer != nil {
		s.saveTimer.Stop()
	}

	// Schedule save in 1 second
	s.saveTimer = time.AfterFunc(time.Second, func() {
		if saveErr := s.Save(); saveErr != nil {
			// Error saving state, but don't crash the timer
			// The next save attempt will try again
			return
		}
	})
}

// UpdateRepository updates a repository state
func (s *Store) UpdateRepository(name string, state RepositoryState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state.UpdateRepository(name, state)
	s.SaveAsync()
}

// GetRepository gets a repository state
func (s *Store) GetRepository(name string) (RepositoryState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.state.GetRepository(name)
}

// GetState returns a copy of the current state
func (s *Store) GetState() *State {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Make a copy
	stateCopy := *s.state
	stateCopy.Repositories = make(map[string]RepositoryState)
	for k, v := range s.state.Repositories {
		stateCopy.Repositories[k] = v
	}

	return &stateCopy
}

// Close closes the state store and ensures final save
func (s *Store) Close() error {
	if s.saveTimer != nil {
		s.saveTimer.Stop()
	}

	return s.Save()
}
