package core

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/aaamil13/CodeIndexerMCP/internal/utils"
)

// Watcher watches for file system changes and triggers re-indexing
type Watcher struct {
	indexer       *Indexer
	watcher       *fsnotify.Watcher
	debounceMap   map[string]*time.Timer
	debounceMutex sync.Mutex
	stopChan      chan struct{}
	logger        *utils.Logger
}

// NewWatcher creates a new file system watcher
func NewWatcher(indexer *Indexer) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		indexer:     indexer,
		watcher:     fsWatcher,
		debounceMap: make(map[string]*time.Timer),
		stopChan:    make(chan struct{}),
		logger:      utils.NewLogger("[Watcher]"),
	}, nil
}

// Start starts watching for file changes
func (w *Watcher) Start() error {
	w.logger.Info("Starting file watcher")

	// Add project directory to watch
	if err := w.addDirectoryRecursive(w.indexer.projectPath); err != nil {
		return err
	}

	// Start event loop
	go w.eventLoop()

	return nil
}

// Stop stops the watcher
func (w *Watcher) Stop() error {
	w.logger.Info("Stopping file watcher")
	close(w.stopChan)
	return w.watcher.Close()
}

// eventLoop processes file system events
func (w *Watcher) eventLoop() {
	for {
		select {
		case <-w.stopChan:
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.logger.Error("Watcher error:", err)
		}
	}
}

// handleEvent handles a file system event
func (w *Watcher) handleEvent(event fsnotify.Event) {
	// Get relative path
	relPath, err := filepath.Rel(w.indexer.projectPath, event.Name)
	if err != nil {
		return
	}

	// Check if should ignore
	if w.indexer.ignoreMatcher.ShouldIgnore(relPath) {
		return
	}

	// Check if we can parse this file
	if !w.indexer.parsers.CanParse(event.Name) {
		return
	}

	switch {
	case event.Op&fsnotify.Write == fsnotify.Write:
		w.logger.Debugf("File modified: %s", relPath)
		w.debounceIndex(event.Name)

	case event.Op&fsnotify.Create == fsnotify.Create:
		w.logger.Debugf("File created: %s", relPath)

		// If it's a directory, add it to watch list
		info, err := filepath.Glob(event.Name)
		if err == nil && len(info) > 0 {
			w.watcher.Add(event.Name)
		}

		w.debounceIndex(event.Name)

	case event.Op&fsnotify.Remove == fsnotify.Remove:
		w.logger.Debugf("File removed: %s", relPath)
		w.handleFileRemoval(event.Name)

	case event.Op&fsnotify.Rename == fsnotify.Rename:
		w.logger.Debugf("File renamed: %s", relPath)
		w.handleFileRemoval(event.Name)
	}
}

// debounceIndex debounces file indexing to avoid multiple rapid updates
func (w *Watcher) debounceIndex(filePath string) {
	w.debounceMutex.Lock()
	defer w.debounceMutex.Unlock()

	// Cancel existing timer
	if timer, exists := w.debounceMap[filePath]; exists {
		timer.Stop()
	}

	// Create new timer
	w.debounceMap[filePath] = time.AfterFunc(300*time.Millisecond, func() {
		w.debounceMutex.Lock()
		delete(w.debounceMap, filePath)
		w.debounceMutex.Unlock()

		// Index the file
		if err := w.indexer.IndexFile(filePath); err != nil {
			w.logger.Errorf("Failed to index file %s: %v", filePath, err)
		} else {
			w.logger.Infof("Re-indexed: %s", filePath)
		}
	})
}

// handleFileRemoval handles file removal
func (w *Watcher) handleFileRemoval(filePath string) {
	relPath, err := filepath.Rel(w.indexer.projectPath, filePath)
	if err != nil {
		return
	}

	// Get file from database
	file, err := w.indexer.db.GetFileByPath(w.indexer.project.ID, relPath)
	if err != nil {
		w.logger.Errorf("Failed to get file: %v", err)
		return
	}

	if file != nil {
		// Delete file from database (cascades to symbols, imports, etc.)
		if err := w.indexer.db.DeleteFile(file.ID); err != nil {
			w.logger.Errorf("Failed to delete file: %v", err)
		} else {
			w.logger.Infof("Removed from index: %s", relPath)
		}
	}
}

// addDirectoryRecursive adds a directory and all subdirectories to the watch list
func (w *Watcher) addDirectoryRecursive(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if should ignore
		relPath, _ := filepath.Rel(w.indexer.projectPath, path)
		if relPath != "." && w.indexer.ignoreMatcher.ShouldIgnore(relPath) {
			if info != nil && isDir(info) {
				return filepath.SkipDir
			}
			return nil
		}

		// Only watch directories
		if info != nil && isDir(info) {
			if err := w.watcher.Add(path); err != nil {
				w.logger.Warnf("Failed to watch directory %s: %v", path, err)
			} else {
				w.logger.Debugf("Watching directory: %s", relPath)
			}
		}

		return nil
	})
}

// Helper function to check if FileInfo represents a directory
func isDir(info any) bool {
	// Type assertion to handle os.FileInfo interface
	type fileInfo interface {
		IsDir() bool
	}
	if fi, ok := info.(fileInfo); ok {
		return fi.IsDir()
	}
	return false
}
