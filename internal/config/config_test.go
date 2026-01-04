package config

import (
	"os"
	"testing"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() (string, error)
		wantErr   bool
	}{
		{
			name: "valid config file",
			setupFunc: func() (string, error) {
				tmpfile, err := os.CreateTemp("", "config*.yaml")
				if err != nil {
					return "", err
				}
				defer tmpfile.Close()

				content := `agent:
  name: "test-agent"
  log_level: "info"
  state_dir: "/tmp/state"
  cache_dir: "/tmp/cache"
  poll_interval: "5m"

repositories:
  - name: "test-repo"
    url: "https://github.com/test/repo.git"
    branch: "main"
    watch_paths:
      - "src/"
    action:
      type: "shell"
      script: "echo 'deploying'"`

				if _, err := tmpfile.WriteString(content); err != nil {
					return "", err
				}
				return tmpfile.Name(), nil
			},
			wantErr: false,
		},
		{
			name: "invalid config file",
			setupFunc: func() (string, error) {
				tmpfile, err := os.CreateTemp("", "config*.yaml")
				if err != nil {
					return "", err
				}
				defer tmpfile.Close()

				content := `invalid: yaml: syntax: here:`
				if _, err := tmpfile.WriteString(content); err != nil {
					return "", err
				}
				return tmpfile.Name(), nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := tt.setupFunc()
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}
			defer os.Remove(tmpfile)

			_, err = NewManager(tmpfile)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewManager() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManagerLoad(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	content := `agent:
  name: "test-agent"
  log_level: "debug"
  state_dir: "/tmp/state"
  cache_dir: "/tmp/cache"
  poll_interval: "3m"

repositories:
  - name: "repo1"
    url: "https://github.com/test/repo.git"
    branch: "main"
    watch_paths:
      - "."
    action:
      type: "shell"
      script: "true"`

	if _, writeErr := tmpfile.WriteString(content); writeErr != nil {
		t.Fatalf("write failed: %v", writeErr)
	}
	tmpfile.Close()

	mgr, mgrErr := NewManager(tmpfile.Name())
	if mgrErr != nil {
		t.Fatalf("NewManager() failed: %v", mgrErr)
	}

	if mgr == nil {
		t.Error("Manager is nil")
	}
}

func TestConfigDefaults(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Minimal config without defaults
	content := `repositories:
  - name: "test"
    url: "https://github.com/test/repo.git"
    watch_paths:
      - "."
    action:
      type: "shell"
      script: "true"`

	if _, writeErr := tmpfile.WriteString(content); writeErr != nil {
		t.Fatalf("write failed: %v", writeErr)
	}
	tmpfile.Close()

	mgr, mgrErr := NewManager(tmpfile.Name())
	if mgrErr != nil {
		t.Fatalf("NewManager() failed: %v", mgrErr)
	}

	// Check that defaults were applied
	if mgr.config.Agent.Name == "" {
		t.Error("Agent name should have default value")
	}
	if mgr.config.Agent.LogLevel == "" {
		t.Error("Log level should have default value")
	}
}
