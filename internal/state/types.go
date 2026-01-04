package state

import "time"

// RepositoryState represents the state of a monitored repository
type RepositoryState struct {
	Name               string    `json:"name"`
	LastFetch          time.Time `json:"last_fetch"`
	CurrentHash        string    `json:"current_hash"`
	LastActionExecuted time.Time `json:"last_action_executed"`
	LastActionStatus   string    `json:"last_action_status"` // success, failure, running
	LastError          string    `json:"last_error"`
}

// State represents the overall state of the cd-gun agent
type State struct {
	Version      string                     `json:"version"`
	LastUpdated  time.Time                  `json:"last_updated"`
	Repositories map[string]RepositoryState `json:"repositories"`
}

// NewState creates a new empty state
func NewState() *State {
	return &State{
		Version:      "1.0",
		LastUpdated:  time.Now(),
		Repositories: make(map[string]RepositoryState),
	}
}

// UpdateRepository updates the state for a repository
func (s *State) UpdateRepository(name string, state RepositoryState) {
	state.Name = name
	s.Repositories[name] = state
	s.LastUpdated = time.Now()
}

// GetRepository gets the state for a repository
func (s *State) GetRepository(name string) (RepositoryState, bool) {
	rs, ok := s.Repositories[name]
	return rs, ok
}

// ActionResult represents the result of executing an action
type ActionResult struct {
	RepositoryName string
	Success        bool
	Output         string
	Error          string
	Duration       time.Duration
	ExecutedAt     time.Time
}
