package app

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	"github.com/omnorm/cd-gun/internal/config"
	"github.com/omnorm/cd-gun/internal/executor"
	"github.com/omnorm/cd-gun/internal/logger"
	"github.com/omnorm/cd-gun/internal/monitor"
	"github.com/omnorm/cd-gun/internal/state"
)

// App is the main application structure
type App struct {
	config     *config.Manager
	configChan chan config.Config
	logger     *logger.Logger
	stateStore *state.Store //nolint:unused // Used in handleMonitorEvent and Stop methods
	monitors   map[string]*monitor.Monitor
	executor   *executor.Executor
	mu         sync.RWMutex
	stopChan   chan struct{}
	wg         sync.WaitGroup
	logFile    *os.File // Log file handle (nil if logging to stdout)
}

// NewApp creates a new application instance
func NewApp(configPath string, logLevel string) (*App, error) {
	// Load config first to get log file path
	configMgr, err := config.NewManager(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	cfg := configMgr.GetConfig()

	// Setup log output (file or stdout)
	var logOut *os.File = os.Stdout
	if cfg.Agent.LogFile != "" {
		logOut, err = os.OpenFile(cfg.Agent.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file '%s': %w", cfg.Agent.LogFile, err)
		}
	}

	// Create logger with configured output
	// Use logLevel from config if set, otherwise use command-line flag
	effectiveLogLevel := logLevel
	if cfg.Agent.LogLevel != "" {
		effectiveLogLevel = cfg.Agent.LogLevel
	}
	log := logger.NewLogger(effectiveLogLevel, logOut)

	// Create state store
	stateStore, err := state.NewStore(cfg.Agent.StateDir)
	if err != nil {
		if logOut != os.Stdout {
			logOut.Close()
		}
		return nil, fmt.Errorf("failed to create state store: %w", err)
	}

	app := &App{
		config:     configMgr,
		configChan: make(chan config.Config, 1),
		logger:     log,
		stateStore: stateStore,
		monitors:   make(map[string]*monitor.Monitor),
		stopChan:   make(chan struct{}),
		logFile:    logOut,
	}

	// Create executor
	app.executor, err = executor.NewExecutor(log)
	if err != nil {
		if logOut != os.Stdout {
			logOut.Close()
		}
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	// Initialize monitors
	if err := app.initializeMonitors(); err != nil {
		if logOut != os.Stdout {
			logOut.Close()
		}
		return nil, fmt.Errorf("failed to initialize monitors: %w", err)
	}

	log.Infof("CD-Gun agent '%s' initialized successfully (state store: %v)", cfg.Agent.Name, app.GetStateStore() != nil)

	return app, nil
}

// initializeMonitors creates monitors for all configured repositories
func (a *App) initializeMonitors() error {
	cfg := a.config.GetConfig()

	for _, repo := range cfg.Repositories {
		mon, err := monitor.NewMonitor(&repo, a.config, a.logger, a.stateStore)
		if err != nil {
			return fmt.Errorf("failed to create monitor for '%s': %w", repo.Name, err)
		}

		a.monitors[repo.Name] = mon
	}

	return nil
}

// Start starts the application
func (a *App) Start() error {
	a.logger.Info("Starting CD-Gun agent...")

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGUSR1)

	// Start monitors
	for name, mon := range a.monitors {
		a.wg.Add(1)
		go func(name string, mon *monitor.Monitor) {
			defer a.wg.Done()
			if err := mon.Start(a.stopChan); err != nil {
				a.logger.Errorf("Monitor for '%s' error: %v", name, err)
			}
		}(name, mon)
	}

	// Start main event loop
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.eventLoop(sigChan)
	}()

	a.logger.Info("CD-Gun agent started successfully")

	// Wait for all goroutines to finish
	a.wg.Wait()

	return nil
}

// eventLoop handles signals and events
func (a *App) eventLoop(sigChan chan os.Signal) {
	// Create a map to listen to all monitor event channels
	cases := []reflect.SelectCase{
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(a.stopChan),
		},
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(sigChan),
		},
	}

	// Add monitor event channels
	a.mu.RLock()
	monitorMap := make(map[int]string) // Map case index to monitor name
	for name, mon := range a.monitors {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(mon.GetEventChan()),
		})
		monitorMap[len(cases)-1] = name
	}
	a.mu.RUnlock()

	// Add timer for periodic checks
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	cases = append(cases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ticker.C),
	})
	timerIndex := len(cases) - 1

	for {
		chosen, recv, ok := reflect.Select(cases)

		if !ok && chosen == 0 {
			// Stop channel closed
			return
		}

		switch chosen {
		case 0:
			// Stop signal
			return

		case 1:
			// Signal received
			if !ok {
				continue
			}
			sig, isSignal := recv.Interface().(os.Signal)
			if !isSignal {
				continue
			}
			switch sig {
			case syscall.SIGTERM, syscall.SIGINT:
				a.logger.Info("Received shutdown signal, gracefully stopping...")
				if err := a.Stop(); err != nil {
					a.logger.Errorf("Error during shutdown: %v", err)
				}
				return

			case syscall.SIGHUP:
				a.logger.Info("Received SIGHUP, reloading configuration...")
				a.reloadConfig()

			case syscall.SIGUSR1:
				a.logger.Info("Received SIGUSR1, forcing repository check...")
				a.forceCheck()
			}

		case timerIndex:
			// Periodic check for config changes
			if a.config.IsModified() {
				a.logger.Info("Configuration file changed, reloading...")
				a.reloadConfig()
			}

		default:
			// Monitor event received
			if !ok {
				continue
			}
			changeEvent, ok := recv.Interface().(monitor.ChangeEvent)
			if !ok {
				continue
			}
			a.handleMonitorEvent(changeEvent)
		}
	}
}

// reloadConfig reloads the configuration
func (a *App) reloadConfig() {
	if err := a.config.Load(); err != nil {
		a.logger.Errorf("Failed to reload config: %v", err)
		return
	}

	a.logger.Info("Configuration reloaded successfully")

	// TODO: Re-initialize monitors with new config
	// For now, just log success
}

// forceCheck triggers a forced check of all repositories
func (a *App) forceCheck() {
	a.mu.RLock()
	monitors := a.monitors
	a.mu.RUnlock()

	for _, mon := range monitors {
		mon.ForceCheck()
	}

	a.logger.Info("Forced check initiated for all repositories")
}

// Stop gracefully stops the application
func (a *App) Stop() error {
	a.logger.Info("Stopping CD-Gun agent...")

	close(a.stopChan)

	// Wait for all monitors to stop with timeout
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All goroutines finished
	case <-time.After(30 * time.Second):
		a.logger.Warn("Shutdown timeout exceeded")
	}

	// Save state
	if err := a.stateStore.Close(); err != nil {
		a.logger.Errorf("Failed to save state: %v", err)
		return err
	}

	// Close log file if open
	if a.logFile != nil && a.logFile != os.Stdout {
		if err := a.logFile.Close(); err != nil {
			a.logger.Errorf("Failed to close log file: %v", err)
		}
	}

	a.logger.Info("CD-Gun agent stopped")

	return nil
}

// handleMonitorEvent handles change events from monitors
func (a *App) handleMonitorEvent(event monitor.ChangeEvent) {
	a.logger.Infof("Processing change event from '%s': %v", event.RepositoryName, event.Files)

	cfg := a.config.GetConfig()
	repo := findRepository(cfg, event.RepositoryName)
	if repo == nil {
		a.logger.Warnf("Repository '%s' not found in config", event.RepositoryName)
		return
	}

	// Execute the action
	result, err := a.executor.Execute(&repo.Action, &event, a.config)
	if err != nil {
		a.logger.Errorf("Failed to execute action for '%s': %v", event.RepositoryName, err)
		return
	}

	// Update state with execution result
	repoState, _ := a.stateStore.GetRepository(event.RepositoryName)
	repoState.LastActionExecuted = result.ExecutedAt
	if result.Success {
		repoState.LastActionStatus = "success"
		repoState.LastError = ""
		a.logger.Infof("Action executed successfully for '%s'", event.RepositoryName)
	} else {
		repoState.LastActionStatus = "failure"
		repoState.LastError = result.Error
		a.logger.Errorf("Action failed for '%s': %s", event.RepositoryName, result.Error)
	}

	a.stateStore.UpdateRepository(event.RepositoryName, repoState)
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

// GetStateStore returns the state store (ensures field is not marked as unused)
func (a *App) GetStateStore() *state.Store {
	return a.stateStore
}
