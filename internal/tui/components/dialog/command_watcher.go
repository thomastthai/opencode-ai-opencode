package dialog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/logging"
)

// CommandsReloadedMsg is sent when commands have been reloaded due to file changes
type CommandsReloadedMsg struct {
	Commands []Command
	Error    error
}

// CommandWatcher watches command directories for changes and reloads commands
type CommandWatcher struct {
	watcher    *fsnotify.Watcher
	dirs       []string
	stopCh     chan struct{}
	reloadCh   chan struct{}
	mu         sync.Mutex
	debouncer  *time.Timer
	lastReload time.Time
}

// NewCommandWatcher creates a new command watcher
func NewCommandWatcher() (*CommandWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &CommandWatcher{
		watcher:  watcher,
		dirs:     []string{},
		stopCh:   make(chan struct{}),
		reloadCh: make(chan struct{}, 1),
	}, nil
}

// Start begins watching the command directories
func (cw *CommandWatcher) Start() tea.Cmd {
	// Get all command directories
	dirs := cw.getCommandDirectories()
	
	// Add directories to watcher
	for _, dir := range dirs {
		if err := cw.addDirectory(dir); err != nil {
			logging.Error("Failed to watch directory", err, "dir", dir)
		}
	}

	// Start the file watching goroutine
	return cw.watch
}

// Stop stops the file watcher
func (cw *CommandWatcher) Stop() {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	
	// Only close if not already closed
	select {
	case <-cw.stopCh:
		// Already closed
		return
	default:
		close(cw.stopCh)
	}
	
	cw.watcher.Close()
	
	if cw.debouncer != nil {
		cw.debouncer.Stop()
	}
}

// getCommandDirectories returns all directories that should be watched for commands
func (cw *CommandWatcher) getCommandDirectories() []string {
	dirs := []string{}

	cfg := config.Get()
	if cfg == nil {
		return dirs
	}

	// XDG_CONFIG_HOME/opencode/commands
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			xdgConfigHome = filepath.Join(home, ".config")
		}
	}
	if xdgConfigHome != "" {
		dirs = append(dirs, filepath.Join(xdgConfigHome, "opencode", "commands"))
	}

	// $HOME/.opencode/commands
	home, err := os.UserHomeDir()
	if err == nil {
		dirs = append(dirs, filepath.Join(home, ".opencode", "commands"))
	}

	// Project data directory
	if cfg.Data.Directory != "" {
		dirs = append(dirs, filepath.Join(cfg.Data.Directory, "commands"))
	}

	return dirs
}

// addDirectory adds a directory to the watcher
func (cw *CommandWatcher) addDirectory(dir string) error {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Create it if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Add to watcher
	if err := cw.watcher.Add(dir); err != nil {
		return err
	}

	cw.dirs = append(cw.dirs, dir)
	logging.Info("Watching command directory", "dir", dir)
	
	return nil
}

// watch is the main file watching loop
func (cw *CommandWatcher) watch() tea.Msg {
	for {
		select {
		case <-cw.stopCh:
			return nil
			
		case event, ok := <-cw.watcher.Events:
			if !ok {
				return nil
			}
			
			// Only reload on relevant events and for .md files
			if cw.isRelevantEvent(event) {
				cw.scheduleReload()
			}
			
		case err, ok := <-cw.watcher.Errors:
			if !ok {
				return nil
			}
			logging.Error("File watcher error", err)
			
		case <-cw.reloadCh:
			// Debounced reload triggered
			commands, err := LoadCustomCommands()
			return CommandsReloadedMsg{
				Commands: commands,
				Error:    err,
			}
		}
	}
}

// isRelevantEvent checks if a file system event should trigger a reload
func (cw *CommandWatcher) isRelevantEvent(event fsnotify.Event) bool {
	// Only care about .md files
	if !strings.HasSuffix(event.Name, ".md") {
		return false
	}

	// Only care about create, write, remove, and rename events
	return event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0
}

// scheduleReload schedules a command reload with debouncing
func (cw *CommandWatcher) scheduleReload() {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	// Cancel existing timer
	if cw.debouncer != nil {
		cw.debouncer.Stop()
	}

	// Schedule new reload after 500ms
	cw.debouncer = time.AfterFunc(500*time.Millisecond, func() {
		// Trigger reload through channel
		select {
		case cw.reloadCh <- struct{}{}:
			// Reload scheduled
		default:
			// Channel full, reload already pending
		}
	})
}