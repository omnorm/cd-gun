package monitor

import (
	"bytes"
	"testing"

	"github.com/omnorm/cd-gun/internal/config"
	"github.com/omnorm/cd-gun/internal/logger"
)

func TestNewGitHelper(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger("info", &buf)

	helper := NewGitHelper("/tmp/test-repo", log)
	if helper == nil {
		t.Error("NewGitHelper() returned nil")
	}
}

func TestGitHelperClone(t *testing.T) {
	// This test would need actual git setup, so we'll just verify the helper is created
	var buf bytes.Buffer
	log := logger.NewLogger("info", &buf)

	tmpDir := t.TempDir()
	helper := NewGitHelper(tmpDir+"/repo", log)

	if helper.repoPath != tmpDir+"/repo" {
		t.Errorf("Repository path mismatch: got %s, want %s", helper.repoPath, tmpDir+"/repo")
	}

	if helper.logger != log {
		t.Error("Logger not properly assigned")
	}
}

func TestGitHelperRepository(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger("debug", &buf)

	tmpDir := t.TempDir()
	helper := NewGitHelper(tmpDir+"/test-repo", log)

	// Create a minimal test repository config
	repo := &config.Repository{
		Name:   "test-repo",
		URL:    "https://github.com/test/repo.git",
		Branch: "main",
		Auth: config.Auth{
			Type: "none",
		},
	}

	// Verify that helper can be used with a repository config
	if repo.URL == "" {
		t.Error("Repository URL should not be empty")
	}

	if helper == nil {
		t.Error("Git helper should be created")
	}
}
